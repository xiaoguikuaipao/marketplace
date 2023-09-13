package goods

import (
	"context"
	"net/http"
	"strconv"

	"grpc/goods_web/forms"
	"grpc/goods_web/global"
	"grpc/goods_web/proto"
	"grpc/goods_web/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func List(ctx *gin.Context) {

	request := &proto.GoodsFilterRequest{}
	priceMinStr := ctx.DefaultQuery("pmin", "0")
	priceMin, _ := strconv.Atoi(priceMinStr)
	request.PriceMin = int32(priceMin)

	priceMaxStr := ctx.DefaultQuery("pmax", "0")
	priceMax, _ := strconv.Atoi(priceMaxStr)
	request.PriceMin = int32(priceMax)

	isHot := ctx.DefaultQuery("ishot", "0")
	if isHot == "1" {
		request.IsHot = true
	}
	isNew := ctx.DefaultQuery("isnew", "0")
	if isNew == "1" {
		request.IsNew = true
	}
	isTab := ctx.DefaultQuery("istab", "0")
	if isTab == "1" {
		request.IsTab = true
	}

	categoryId := ctx.DefaultQuery("c", "0")
	categoryIdInt, _ := strconv.Atoi(categoryId)
	request.TopCategory = int32(categoryIdInt)

	brandId := ctx.DefaultQuery("brand", "0")
	brandIdInt, _ := strconv.Atoi(brandId)
	request.Brand = int32(brandIdInt)

	page := ctx.DefaultQuery("p", "0")
	pageInt, _ := strconv.Atoi(page)
	request.Pages = int32(pageInt)

	pageNums := ctx.DefaultQuery("pnums", "0")
	pageNumsInt, _ := strconv.Atoi(pageNums)
	request.PagePerNums = int32(pageNumsInt)

	//TODO
	// 用elasticsearch搜索关键词
	keywords := ctx.DefaultQuery("kw", "")
	request.KeyWords = keywords

	//请求商品srv
	rsp, err := global.GoodsSrvClient.GoodsList(context.Background(), request)
	if err != nil {
		zap.S().Errorw("[List]【查询商品列表失败】")
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	reMap := map[string]interface{}{
		"total": rsp.Total,
		//"data":  rsp.Data,
	}
	goodsList := make([]interface{}, 0)
	// 目的是自定义返回的json数据格式
	for _, value := range rsp.Data {
		goodsList = append(goodsList, map[string]interface{}{
			"id":          value.Id,
			"name":        value.Name,
			"goods_brief": value.GoodsBrief,
			"desc":        value.GoodsDesc,
			"ship_free":   value.ShipFree,
			"images":      value.Images,
			"desc_images": value.DescImages,
			"front_image": value.GoodsFrontImage,
			"shop_price":  value.ShopPrice,
			"category": map[string]interface{}{
				"id":   value.Category.Id,
				"name": value.Category.Name,
			},
			"brand": map[string]interface{}{
				"id":   value.Brand.Id,
				"name": value.Brand.Name,
				"logo": value.Brand.Logo,
			},
			"is_hot":    value.IsHot,
			"is_new":    value.IsNew,
			"on_sale":   value.OnSale,
			"click_num": value.ClickNum,
			"fav_num":   value.FavNum,
			"sold_num":  value.SoldNum,
		})
	}
	reMap["data"] = goodsList
	// 如果这样返回，json的格式就是rsp.data里的goodInfoResponse的json格式(是由proto文件自动生成的), 驼峰式。
	ctx.JSON(http.StatusOK, reMap)
}

func New(ctx *gin.Context) {
	goodsForm := forms.GoodsForm{}
	if err := ctx.ShouldBindJSON(&goodsForm); err != nil {
		utils.HandleValidatorError(ctx, err)
		return
	}
	goodsClient := global.GoodsSrvClient
	rsp, err := goodsClient.CreateGoods(context.Background(), &proto.CreateGoodsInfo{
		Name:            goodsForm.Name,
		GoodsSn:         goodsForm.GoodsSn,
		Stocks:          goodsForm.Stocks,
		MarketPrice:     goodsForm.MarketPrice,
		ShopPrice:       goodsForm.ShopPrice,
		GoodsBrief:      goodsForm.GoodsBrief,
		ShipFree:        *goodsForm.ShipFree,
		Images:          goodsForm.Images,
		DescImages:      goodsForm.DescImages,
		GoodsFrontImage: goodsForm.FrontImage,
		CategoryId:      goodsForm.CategoryId,
		BrandId:         goodsForm.Brand,
	})
	if err != nil {
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	//如何设置库存
	//TODO
	ctx.JSON(http.StatusOK, rsp)
}

func Detail(ctx *gin.Context) {
	id := ctx.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	r, err := global.GoodsSrvClient.GetGoodsDetail(context.Background(), &proto.GoodInfoRequest{Id: int32(i)})
	if err != nil {
		utils.HandleGrpcErrorToHttp(err, ctx)
	}

	//关于库存信息、点赞数、点击数等
	// 1. 拿着这个r中的信息请求库存服务等，返回
	// 2. 让前端自己去访问库存服务，异步返回

	rsp := map[string]interface{}{
		"id":   r.Id,
		"name": r.Name,
		"desc": r.GoodsBrief,
		"category": map[string]interface{}{
			"id": r.Category.Id,
		},
		"brand": map[string]interface{}{
			"id": r.Brand.Id,
		},
	}
	ctx.JSON(http.StatusOK, rsp)
}

func Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	_, err = global.GoodsSrvClient.DeleteGoods(context.Background(), &proto.DeleteGoodsInfo{Id: int32(i)})

	if err != nil {
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.Status(http.StatusOK)
	return
}

func Stocks(ctx *gin.Context) {
	id := ctx.Param("id")
	_, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	//TODO 库存
	return
}

func UpdateStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	updateForm := forms.GoodsStatusForm{}
	if err := ctx.BindJSON(&updateForm); err != nil {
		utils.HandleValidatorError(ctx, err)
		return
	}
	_, err = global.GoodsSrvClient.UpdateGoods(context.Background(), &proto.CreateGoodsInfo{
		Id:     int32(i),
		IsNew:  *updateForm.IsNew,
		IsHot:  *updateForm.IsHot,
		OnSale: *updateForm.OnSale,
	})
	if err != nil {
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "修改成功",
	})
}

func Update(ctx *gin.Context) {
	id := ctx.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	goodsForm := forms.GoodsForm{}
	if err := ctx.BindJSON(&goodsForm); err != nil {
		utils.HandleValidatorError(ctx, err)
		return
	}
	goodsClient := global.GoodsSrvClient
	_, err = goodsClient.UpdateGoods(context.Background(), &proto.CreateGoodsInfo{
		Id:              int32(i),
		Name:            goodsForm.Name,
		GoodsSn:         goodsForm.GoodsSn,
		Stocks:          goodsForm.Stocks,
		MarketPrice:     goodsForm.MarketPrice,
		ShopPrice:       goodsForm.ShopPrice,
		GoodsBrief:      goodsForm.GoodsBrief,
		ShipFree:        *goodsForm.ShipFree,
		Images:          goodsForm.Images,
		DescImages:      goodsForm.DescImages,
		GoodsFrontImage: goodsForm.FrontImage,
		CategoryId:      goodsForm.CategoryId,
		BrandId:         goodsForm.Brand,
	})
	if err != nil {
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "更新成功",
	})
}
