package order

import (
	"context"
	"net/http"
	"strconv"

	"grpc/order_web/forms"
	"grpc/order_web/global"
	"grpc/order_web/middlewares"
	"grpc/order_web/proto"
	"grpc/order_web/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func List(ctx *gin.Context) {
	userId, _ := ctx.Get("userID")
	claims, _ := ctx.Get("claims")
	model := claims.(*middlewares.MyClaims)
	request := proto.OrderFilterRequest{}
	//如果是管理员用户，则不穿userid，获得所有订单
	if model.Role == 1 {
		request.UserId = int32(userId.(int64))
	}
	//获取分页信息
	page := ctx.DefaultQuery("p", "0")
	pageInt, _ := strconv.Atoi(page)
	pageNum := ctx.DefaultQuery("pnum", "0")
	pageNumInt, _ := strconv.Atoi(pageNum)
	request.Page = int32(pageInt)
	request.PageNum = int32(pageNumInt)

	rsp, err := global.OrderSrvClient.OrderList(context.Background(), &request)
	if err != nil {
		zap.L().Error("获取订单列表错误", zap.Error(err))
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	reMap := gin.H{
		"total": rsp.Total,
	}
	orderList := make([]interface{}, 0)
	for _, item := range rsp.Data {
		tmpMap := map[string]interface{}{}
		tmpMap["id"] = item.Id
		tmpMap["status"] = item.Status
		tmpMap["pay_type"] = item.PayType
		tmpMap["user"] = item.UserId
		tmpMap["post"] = item.Post
		tmpMap["total"] = item.Total
		tmpMap["address"] = item.Address
		tmpMap["mobile"] = item.Mobile
		tmpMap["order_sn"] = item.OrderSn
		tmpMap["add_time"] = item.AddTime

		orderList = append(orderList, tmpMap)
	}
	reMap["data"] = orderList
	ctx.JSON(http.StatusOK, reMap)
}

func New(ctx *gin.Context) {
	orderForm := forms.CreateOrderForm{}
	if err := ctx.ShouldBindJSON(&orderForm); err != nil {
		utils.HandleValidatorError(ctx, err)
		return
	}
	userId, _ := ctx.Get("userID")
	rsp, err := global.OrderSrvClient.Create(context.Background(), &proto.OrderRequest{
		UserId:  int32(userId.(int64)),
		Address: orderForm.Address,
		Name:    orderForm.Name,
		Mobile:  orderForm.Mobile,
		Post:    orderForm.Post,
	})
	if err != nil {
		zap.S().Error(err)
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	// 生成支付宝的支付url
	//TODO
	ctx.JSON(http.StatusOK, gin.H{
		"id": rsp.Id,
		//"alipay_url": url,
	})
}

func Detail(ctx *gin.Context) {
	id := ctx.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"msg": "格式错误",
		})
		return
	}
	userId, _ := ctx.Get("userID")
	claims, _ := ctx.Get("claims")
	model := claims.(*middlewares.MyClaims)
	request := proto.OrderRequest{
		Id: int32(i),
	}
	//如果是管理员用户，则不传userid，获得所有订单
	if model.Role == 1 {
		request.UserId = int32(userId.(int64))
	}

	rsp, err := global.OrderSrvClient.OrderDetail(context.Background(), &request)
	if err != nil {
		zap.S().Error(err)
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	reMap := gin.H{}
	reMap["id"] = rsp.OrderInfo.Id
	reMap["status"] = rsp.OrderInfo.Status
	reMap["user"] = rsp.OrderInfo.UserId
	reMap["post"] = rsp.OrderInfo.Post
	reMap["address"] = rsp.OrderInfo.Address
	reMap["mobile"] = rsp.OrderInfo.Mobile
	reMap["pay_type"] = rsp.OrderInfo.PayType
	reMap["order_sn"] = rsp.OrderInfo.OrderSn
	reMap["total"] = rsp.OrderInfo.Total
	reMap["addTime"] = rsp.OrderInfo.AddTime

	goodsList := make([]interface{}, 0)
	for _, item := range rsp.Goods {
		tmpMap := gin.H{
			"id":    item.GoodsId,
			"name":  item.GoodsName,
			"image": item.GoodsImage,
			"price": item.GoodsPrice,
			"nums":  item.Nums,
		}
		goodsList = append(goodsList, tmpMap)
	}
	reMap["goods"] = goodsList
	//订单详情也要加上支付url
	reMap["alipay_url"] = ""
	ctx.JSON(http.StatusOK, reMap)
}
