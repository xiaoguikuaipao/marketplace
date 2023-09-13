package shop_cart

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"grpc/order_web/forms"
	"grpc/order_web/global"
	"grpc/order_web/proto"
	"grpc/order_web/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func List(ctx *gin.Context) {
	userId, _ := ctx.Get("userID")
	rsp, err := global.OrderSrvClient.CartItemList(context.Background(), &proto.UserInfo{Id: int32(userId.(int64))})
	if err != nil {
		fmt.Println(err)
		zap.L().Error("[List]【查询购物车列表失败】", zap.Error(err))
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	ids := make([]int32, 0)
	for _, item := range rsp.Data {
		ids = append(ids, item.GoodsId)
	}
	if len(ids) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"total": 0,
		})
		return
	}

	//请求商品服务获取商品信息
	goodsRsp, err := global.GoodsSrvClient.BatchGetGoods(context.Background(), &proto.BatchGoodsIdInfo{Id: ids})
	if err != nil {
		zap.S().Errorw("[List]【批量查询商品失败】")
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	reMap := gin.H{
		"total": rsp.Total,
	}
	goodsList := make([]interface{}, 0)
	for _, item := range rsp.Data {
		for _, good := range goodsRsp.Data {
			if good.Id == item.GoodsId {
				tmpMap := map[string]interface{}{}
				tmpMap["id"] = item.Id
				tmpMap["goods_id"] = item.GoodsId
				tmpMap["good_name"] = good.Name
				tmpMap["good_image"] = good.GoodsFrontImage
				tmpMap["good_price"] = good.ShopPrice
				tmpMap["nums"] = item.Nums
				tmpMap["check"] = item.Checked

				goodsList = append(goodsList, tmpMap)
			}
		}
	}
	reMap["data"] = goodsList
	ctx.JSON(http.StatusOK, reMap)
}

func New(ctx *gin.Context) {
	itemForm := forms.ShopCartItemForm{}
	if err := ctx.ShouldBindJSON(&itemForm); err != nil {
		utils.HandleValidatorError(ctx, err)
		return
	}

	//先确认商品是否存在
	//检查添加数量< 库存数量
	goods, err := global.InventoryClient.InvDetail(context.Background(), &proto.GoodsInvInfo{GoodsId: itemForm.GoodsId})
	if err != nil {
		zap.L().Error("[New]【获取商品信息失败】", zap.Error(err))
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	if goods.Num < itemForm.Nums {
		ctx.JSON(http.StatusOK, gin.H{
			"nums": "库存不足",
		})
		return
	}

	userId, _ := ctx.Get("userID")
	rsp, err := global.OrderSrvClient.CreateCartItem(context.Background(), &proto.CartItemRequest{
		UserId:  int32(userId.(int64)),
		GoodsId: itemForm.GoodsId,
		Nums:    itemForm.Nums,
		Checked: false,
	})
	if err != nil {
		zap.S().Errorw("添加到购物车失败")
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id": rsp.Id,
	})
}

func Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"msg": "url格式错误",
		})
		return
	}

	userId, _ := ctx.Get("userID")
	_, err = global.OrderSrvClient.DeleteCartItem(context.Background(), &proto.CartItemRequest{
		UserId:  int32(userId.(int64)),
		GoodsId: int32(i),
	})
	if err != nil {
		zap.S().Errorw("删除失败")
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.Status(http.StatusOK)
}

func Update(ctx *gin.Context) {
	itemForm := forms.ShopCartItemUpdateForm{}
	if err := ctx.ShouldBindJSON(&itemForm); err != nil {
		zap.S().Error(err)
		utils.HandleValidatorError(ctx, err)
		return
	}
	userId, _ := ctx.Get("userID")
	id := ctx.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"msg": "无效参数",
		})
	}
	request := proto.CartItemRequest{
		UserId:  int32(userId.(int64)),
		GoodsId: int32(i),
		Nums:    itemForm.Nums,
		Checked: false,
	}
	if itemForm.Checked != nil {
		request.Checked = *itemForm.Checked
	}
	_, err = global.OrderSrvClient.UpdateCartItem(context.Background(), &request)
	if err != nil {
		zap.S().Error(err)
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
}
