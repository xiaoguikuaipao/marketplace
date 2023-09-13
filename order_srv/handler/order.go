package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"grpc/order_srv/global"
	"grpc/order_srv/models"
	"grpc/order_srv/proto"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrderServer struct {
	*proto.UnimplementedOrderServer
}

type OrderListener struct {
	Code        codes.Code
	Detail      string
	ID          int32
	OrderAmount float32
}

var OL OrderListener

func (*OrderServer) CartItemList(c context.Context, req *proto.UserInfo) (*proto.CartItemListResponse, error) {
	var shopCarts []models.ShoppingCart
	var rsp proto.CartItemListResponse
	if result := global.DB.Where(&models.ShoppingCart{User: req.Id}).Find(&shopCarts); result.Error != nil {
		return nil, result.Error
	} else {
		rsp.Total = int32(result.RowsAffected)
	}

	for _, shopCart := range shopCarts {
		rsp.Data = append(rsp.Data, &proto.ShopCartInfoResponse{
			Id:      shopCart.ID,
			UserId:  shopCart.User,
			GoodsId: shopCart.Goods,
			Nums:    shopCart.Nums,
			Checked: shopCart.Checked,
		})
	}
	return &rsp, nil

}

func (*OrderServer) CreateCartItem(c context.Context, req *proto.CartItemRequest) (*proto.ShopCartInfoResponse, error) {
	//将商品添加到购物车有两种情况
	var shopCart models.ShoppingCart
	if result := global.DB.Where(&models.ShoppingCart{
		Goods: req.GoodsId,
		User:  req.UserId,
	}).First(&shopCart); result.RowsAffected == 1 {
		shopCart.Nums += req.Nums
	} else {
		shopCart.User = req.UserId
		shopCart.Goods = req.GoodsId
		shopCart.Nums = req.Nums
		shopCart.Checked = false
	}
	global.DB.Save(&shopCart)
	return &proto.ShopCartInfoResponse{Id: shopCart.ID}, nil
}

func (*OrderServer) UpdateCartItem(c context.Context, req *proto.CartItemRequest) (*emptypb.Empty, error) {
	var shopCart models.ShoppingCart
	if result := global.DB.Where("goods = ? and user = ?", req.GoodsId, req.UserId).First(&shopCart); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "购物车记录不存在")
	}

	//细节，前端可能只传Check
	if req.Nums > 0 {
		shopCart.Checked = req.Checked
	}
	shopCart.Nums = req.Nums
	global.DB.Save(&shopCart)
	return &emptypb.Empty{}, nil
}

func (*OrderServer) DeleteCartItem(c context.Context, req *proto.CartItemRequest) (*emptypb.Empty, error) {
	if result := global.DB.Where("goods = ? and user = ?", req.GoodsId, req.UserId).Delete(&models.ShoppingCart{}); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "购物车记录不存在")
	}
	return &emptypb.Empty{}, nil
}

// ExecuteLocalTransaction 发送half消息之后执行的本地事务
func (o *OrderListener) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
	var orderInfo models.OrderInfo
	_ = json.Unmarshal(msg.Body, &orderInfo)

	goodsIds := make([]int32, 0)
	var shopCarts []models.ShoppingCart
	goodsNums := make(map[int32]int32)
	if result := global.DB.Where(&models.ShoppingCart{
		User:    orderInfo.User,
		Checked: true,
	}).Find(&shopCarts); result.RowsAffected == 0 {
		o.Code = codes.InvalidArgument
		o.Detail = "没有选中商品"
		return primitive.RollbackMessageState
	}
	//根据用户id筛选出用户的购物车商品列表
	//将每一项的商品id添加到数组中
	//记录每一件购物车商品数量
	for _, shopCart := range shopCarts {
		goodsIds = append(goodsIds, shopCart.Goods)
		goodsNums[shopCart.Goods] = shopCart.Nums
	}

	//商品服务调用
	goods, err := global.GoodsSrvClient.BatchGetGoods(context.Background(), &proto.BatchGoodsIdInfo{Id: goodsIds})
	if err != nil {
		o.Code = codes.Internal
		o.Detail = "批量查询商品信息失败"
		return primitive.RollbackMessageState
	}
	var orderAmount float32
	var orderGoods []*models.OrderGoods
	var goodsInvInfo []*proto.GoodsInvInfo
	for _, good := range goods.Data {
		orderAmount += good.ShopPrice * float32(goodsNums[good.Id]) //通过商品id查询商品购买数量)
		orderGoods = append(orderGoods, &models.OrderGoods{
			//注意此时OrderID 还获取不到
			//Order:
			Goods:      good.Id,
			GoodsName:  good.Name,
			GoodsImage: good.GoodsFrontImage,
			GoodsPrice: good.ShopPrice,
			Nums:       goodsNums[good.Id],
		})

		goodsInvInfo = append(goodsInvInfo, &proto.GoodsInvInfo{
			GoodsId: good.Id,
			Num:     goodsNums[good.Id],
		})
	}

	//库存服务调用
	_, err = global.InventorySrvClient.Sell(context.Background(), &proto.SellInfo{
		OrderSn:   orderInfo.OrderSn,
		GoodsInfo: goodsInvInfo,
	})
	// 有没有可能，err != nil， 但实际上在库存服务端却成功了？
	if err != nil {
		if statusErr, ok := status.FromError(err); ok && (statusErr.Code() == codes.InvalidArgument || statusErr.Code() == codes.ResourceExhausted) {
			o.Code = statusErr.Code()
			o.Detail = statusErr.Message()
			return primitive.RollbackMessageState
		}
		o.Detail = "遇到未知错误"
		return primitive.CommitMessageState
	}

	////测试错误
	//o.Code = codes.Internal
	//o.Detail = "未知错误"
	//return primitive.CommitMessageState

	//生成订单信息表
	tx := global.DB.Begin()
	orderInfo.OrderMount = orderAmount
	o.OrderAmount = orderAmount
	if result := tx.Save(&orderInfo); result.RowsAffected == 0 {
		tx.Rollback()
		o.Code = codes.Internal
		o.Detail = "已扣减，订单信息生成失败"
		return primitive.CommitMessageState
	}
	o.ID = orderInfo.ID

	//生成订单商品表
	for _, orderGood := range orderGoods {
		orderGood.Order = orderInfo.ID
	}
	if result := tx.CreateInBatches(orderGoods, 100); result.RowsAffected == 0 {
		tx.Rollback()
		o.Code = codes.Internal
		o.Detail = "已扣减，生成订单商品表失败"
		return primitive.CommitMessageState
	}

	//删除购物车记录
	if result := tx.Where(&models.ShoppingCart{
		User:    orderInfo.User,
		Checked: true,
	}).Where("Goods in ?", goodsIds).Delete(&models.ShoppingCart{}); result.RowsAffected == 0 {
		tx.Rollback()
		o.Code = codes.Internal
		o.Detail = "已扣减，购物车记录删除失败"
		return primitive.CommitMessageState
	}

	// 发送延时消息归还库存
	//delay, err := rocketmq.NewProducer(producer.WithNameServer([]string{"192.168.224.128:9876"}), producer.WithInstanceName("延时消息生产者"))
	//if err != nil {
	//	zap.L().Error("启动延时消息生产者失败", zap.Error(err))
	//	o.Code = codes.Internal
	//	o.Detail = "启动延时消息生产者失败"
	//	tx.Rollback()
	//	return primitive.CommitMessageState
	//}
	if err = global.P.Start(); err != nil {
		o.Code = codes.Internal
		o.Detail = "启动延时消息生产者失败"
		zap.L().Error("启动延时消息生产者失败", zap.Error(err))
		tx.Rollback()
		return primitive.CommitMessageState
	}
	delayMsg := primitive.NewMessage("order_timeout", msg.Body)
	delayMsg.WithDelayTimeLevel(16)
	_, err = global.P.SendSync(context.Background(), delayMsg)
	if err != nil {
		o.Code = codes.Internal
		o.Detail = "发送延时消息失败"
		tx.Rollback()
		zap.L().Error("发送延时消息失败", zap.Error(err))
		return primitive.CommitMessageState
	}

	tx.Commit()
	return primitive.RollbackMessageState
}

func (o *OrderListener) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
	var orderInfo models.OrderInfo
	_ = json.Unmarshal(msg.Body, &orderInfo)
	if result := global.DB.Where(models.OrderInfo{OrderSn: orderInfo.OrderSn}).First(&orderInfo); result.RowsAffected == 0 {
		if strings.Contains(o.Detail, "已扣减") {
			return primitive.CommitMessageState
		}
	}
	return primitive.RollbackMessageState
}

// Create 创建订单
func (*OrderServer) Create(c context.Context, req *proto.OrderRequest) (*proto.OrderInfoResponse, error) {
	var rsp proto.OrderInfoResponse
	OL = OrderListener{}
	//0.. 从购物车中获取选中商品
	//1. 商品的金额要自己查询 - 访问商品服务

	//分布式事务begin
	//2. 库存扣减 - 访问库存服务

	// 本地事务begin
	// 4. 订单的基本信息表 ， 订单的商品信息表
	//5. 从购物车删除记录
	// 本地事务end
	// 分布式事务end
	//orderListener := OrderListener{}
	//p, err := rocketmq.NewTransactionProducer(&orderListener, producer.WithNameServer([]string{"192.168.224.128:9876"}), producer.WithInstanceName("归还消息生产者"))
	//if err != nil {
	//	zap.L().Error("创建归还生产者失败", zap.Error(err))
	//	return nil, err
	//}
	if err := global.TP.Start(); err != nil {
		zap.L().Error("生产者启动失败", zap.Error(err))
		return nil, err
	}

	// 先生成订单编号和基本信息
	orderInfo := models.OrderInfo{
		User:         req.UserId,
		OrderSn:      GenerateOrderSn(req.UserId),
		Address:      req.Address,
		SignerName:   req.Name,
		SingerMobile: req.Mobile,
		Post:         req.Post,
	}
	jsonString, _ := json.Marshal(orderInfo)

	// 发送half归还信息，然后转到本地事务处理
	_, err := global.TP.SendMessageInTransaction(context.Background(), primitive.NewMessage("order_reback", jsonString))
	if err != nil {
		zap.L().Error("发送half归还消息失败", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "发送消息失败: %s", err.Error())
	} else {
		zap.L().Info("发送half归还消息成功")
	}
	if OL.Detail != "" {
		return nil, status.Errorf(codes.Internal, "新建订单失败: %s", OL.Detail)
	}
	rsp.Id = OL.ID
	rsp.OrderSn = orderInfo.OrderSn
	rsp.Total = OL.OrderAmount

	return &rsp, nil
}

func (*OrderServer) OrderList(c context.Context, req *proto.OrderFilterRequest) (*proto.OrderListResponse, error) {
	var orders []models.OrderInfo
	var rsp proto.OrderListResponse
	var total int64

	global.DB.Model(&models.OrderInfo{}).Where(&models.OrderInfo{User: req.UserId}).Count(&total)

	rsp.Total = int32(total)

	global.DB.Where(&models.OrderInfo{User: req.UserId}).Scopes(paginate(int(req.Page), int(req.PageNum))).Find(&orders)
	for _, order := range orders {
		rsp.Data = append(rsp.Data, &proto.OrderInfoResponse{
			Id:      order.ID,
			UserId:  order.User,
			OrderSn: order.OrderSn,
			PayType: order.PayType,
			Status:  order.Status,
			Post:    order.Post,
			Total:   order.OrderMount,
			Address: order.Address,
			Name:    order.SignerName,
			Mobile:  order.SingerMobile,
			AddTime: order.CreatedAt.Format("2006-01-02 11:04:05"),
		})
	}
	return &rsp, nil
}

func (*OrderServer) OrderDetail(c context.Context, req *proto.OrderRequest) (*proto.OrderInfoDetailResponse, error) {
	var order models.OrderInfo
	var rsp proto.OrderInfoDetailResponse

	//web层应该检查权限，避免爬到别的用户的订单
	if result := global.DB.Where(&models.OrderInfo{
		BaseModel: models.BaseModel{ID: req.Id},
		User:      req.UserId,
	}).Find(&order); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "订单记录不存在")
	}

	rsp.OrderInfo = &proto.OrderInfoResponse{
		Id:      order.ID,
		UserId:  order.User,
		OrderSn: order.OrderSn,
		PayType: order.PayType,
		Status:  order.Status,
		Post:    order.Post,
		Total:   order.OrderMount,
		Address: order.Address,
		Name:    order.SignerName,
		Mobile:  order.SingerMobile,
		AddTime: order.CreatedAt.Format("2006-01-12"),
	}

	var orderGoods []models.OrderGoods
	if result := global.DB.Where(&models.OrderGoods{Order: order.ID}).Find(&orderGoods); result.Error != nil {
		return nil, result.Error
	}
	for _, orderGood := range orderGoods {
		rsp.Goods = append(rsp.Goods, &proto.OrderItemResponse{
			GoodsId:    orderGood.Goods,
			GoodsName:  orderGood.GoodsName,
			GoodsImage: orderGood.GoodsImage,
			GoodsPrice: orderGood.GoodsPrice,
			Nums:       orderGood.Nums,
		})
	}
	return &rsp, nil

}

func (*OrderServer) UpdateOrderStatus(c context.Context, req *proto.OrderStatus) (*emptypb.Empty, error) {
	result := global.DB.Model(&models.OrderInfo{}).Where("order_sn = ?", req.OrderSn).Update("status", req.Status)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "订单不存在")
	}
	return &emptypb.Empty{}, nil
}

func GenerateOrderSn(userId int32) string {
	//年月日时分秒+用户id+2位随机数
	now := time.Now()
	rand.Seed(uint64(time.Now().UnixNano()))
	orderSn := fmt.Sprintf("%d%d%d%d%d%d%d%d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), userId, rand.Intn(90)+10)
	return orderSn
}
