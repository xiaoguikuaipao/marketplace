package handler

import (
	"context"
	"encoding/json"

	"grpc/inventory_srv/global"
	"grpc/inventory_srv/models"
	"grpc/inventory_srv/proto"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InventoryServer struct {
	*proto.UnimplementedInventoryServer
}

func (s *InventoryServer) SetInv(ctx context.Context, req *proto.GoodsInvInfo) (*emptypb.Empty, error) {
	var inv models.Inventory
	global.DB.Where(&models.Inventory{Goods: req.GoodsId}).First(&inv)
	inv.Goods = req.GoodsId
	inv.Stocks = req.Num
	global.DB.Save(&inv)
	return &emptypb.Empty{}, nil
}

func (s *InventoryServer) InvDetail(ctx context.Context, req *proto.GoodsInvInfo) (*proto.GoodsInvInfo, error) {
	var inv models.Inventory
	if result := global.DB.Where(&models.Inventory{Goods: req.GoodsId}).First(&inv, req.GoodsId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "没有库存信息")
	}
	return &proto.GoodsInvInfo{
		GoodsId: inv.Goods,
		Num:     inv.Stocks,
	}, nil
}

//var m sync.Mutex

func (s *InventoryServer) Sell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//1. 事务问题
	//扣减库存，本地事务[A: 10, B: 5, C: 20]
	//要么全部扣除成功，要么全部扣除失败
	//2. 并发问题
	//两个goroutine同时通过了两个if，来到扣减库存处，但此时库存只有1
	tx := global.DB.Begin()
	sellDetail := models.StockRebackDetail{
		Status:  1,
		OrderSn: req.OrderSn,
	}
	var details []models.GoodsDetail
	//这样锁有问题， 1. 请求的不是同一商品，却被锁住
	//m.Lock()
	for _, good := range req.GoodsInfo {
		var inv models.Inventory
		details = append(details, models.GoodsDetail{
			Goods: good.GoodsId,
			Nums:  good.Num,
		})
		//悲观锁
		//if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&models.Inventory{Goods: good.GoodsId}).First(&inv, good.GoodsId); result.RowsAffected == 0 {
		//	tx.Rollback()
		//	return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		//}

		//乐观锁
		for {
			if result := global.DB.Where(&models.Inventory{Goods: good.GoodsId}).First(&inv, good.GoodsId); result.RowsAffected == 0 {
				tx.Rollback()
				return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
			}
			if inv.Stocks < good.Num {
				tx.Rollback()
				return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
			}
			inv.Stocks -= good.Num
			//坑： gorm的Updates函数对应的sql语句UPDATE xxx set yyy = zzz。如果想设置这个zzz是0，有可能会被gorm自动当成默认值忽略
			//方法：使用select选定某几列。
			if result := tx.Model(&models.Inventory{}).Where("goods=? and version=?", good.GoodsId, inv.Version).Select("Stocks", "Version").Updates(&models.Inventory{
				Stocks:  inv.Stocks,
				Version: inv.Version + 1,
			}); result.RowsAffected == 0 {
				zap.L().Info("库存扣减失败")
			} else {
				break
			}
		}
		//tx.Save(&inv)
	}
	sellDetail.Detail = details
	if result := tx.Model(&models.StockRebackDetail{}).Create(&sellDetail); result.Error != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "保存库存扣减历史失败")
	}
	tx.Commit()
	//m.Unlock()
	return &emptypb.Empty{}, nil
}

func (s *InventoryServer) Reback(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//库存归还：1. 订单超时归还, 2. 订单创建失败 3. 手动归还(前两种必须做到)
	tx := global.DB.Begin()
	for _, good := range req.GoodsInfo {
		var inv models.Inventory
		if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&models.Inventory{Goods: good.GoodsId}).First(&inv, good.GoodsId); result.RowsAffected == 0 {
			tx.Rollback()
			return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		}
		// 这里要加分布式锁
		inv.Stocks += good.Num
		tx.Save(&inv)
	}
	tx.Commit()
	return &emptypb.Empty{}, nil
}

func AutoReback(ctx context.Context, ext ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	type orderInfo struct {
		OrderSn string
	}
	//注意幂等性
	for i := range ext {
		var order orderInfo
		err := json.Unmarshal(ext[i].Body, &order)
		if err != nil {
			zap.L().Error("解析json失败", zap.Error(err))
			return consumer.ConsumeSuccess, err
		}

		tx := global.DB.Begin()
		var sellDetail models.StockRebackDetail
		if result := tx.Model(&models.StockRebackDetail{}).Where(&models.StockRebackDetail{
			OrderSn: order.OrderSn,
			Status:  1,
		}).First(&sellDetail); result.RowsAffected == 0 {
			return consumer.ConsumeSuccess, nil
		}

		//对订单的每个商品归还库存
		for _, orderGood := range sellDetail.Detail {
			result := tx.Model(&models.Inventory{}).Where(&models.Inventory{Goods: orderGood.Goods}).Update("stocks", gorm.Expr("stocks+?", orderGood.Nums))
			if result.RowsAffected == 0 {
				tx.Rollback()
				return consumer.ConsumeRetryLater, nil
			}
		}

		//确认订单归还状态
		result := tx.Model(&models.StockRebackDetail{}).Where(&models.StockRebackDetail{OrderSn: order.OrderSn}).Update("status", 2)
		if result.RowsAffected == 0 {
			tx.Rollback()
			return consumer.ConsumeRetryLater, nil
		}
		tx.Commit()
		return consumer.ConsumeSuccess, nil
	}
	return consumer.ConsumeSuccess, nil
}
