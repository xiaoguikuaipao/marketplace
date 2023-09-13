package brand

import (
	"context"
	"net/http"
	"strconv"

	"grpc/goods_web/forms"
	"grpc/goods_web/global"
	"grpc/goods_web/proto"
	"grpc/goods_web/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func New(ctx *gin.Context) {
	brandForm := forms.BrandForm{}
	err := ctx.ShouldBindJSON(&brandForm)
	if err != nil {
		if verr, ok := err.(validator.ValidationErrors); ok {
			utils.HandleValidatorError(ctx, verr)
		} else {
			zap.S().Errorw("创建品牌失败", err.Error())
		}
		return
	}
	brand, err := global.GoodsSrvClient.CreateBrand(context.Background(), &proto.BrandRequest{
		Name: brandForm.Name,
		Logo: brandForm.Logo,
	})
	if err != nil {
		zap.S().Errorw("创建品牌失败", err.Error())
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "品牌新建成功",
		"id":  brand.Id,
	})
}

func Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg":   "无效参数",
			"error": err.Error(),
		})
		return
	}
	_, err = global.GoodsSrvClient.DeleteBrand(context.Background(), &proto.BrandRequest{Id: int32(idInt)})
	if err != nil {
		zap.S().Errorw("删除品牌失败", err.Error())
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "删除成功",
	})
}

func Update(ctx *gin.Context) {
	brandForm := forms.BrandUpdateForm{}
	err := ctx.ShouldBindJSON(&brandForm)
	if err != nil {
		utils.HandleValidatorError(ctx, err)
		return
	}
	id, err := strconv.Atoi(brandForm.Id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "无效参数",
		})
		return
	}
	_, err = global.GoodsSrvClient.UpdateBrand(context.Background(), &proto.BrandRequest{
		Id:   int32(id),
		Name: brandForm.Name,
		Logo: brandForm.Logo,
	})
	if err != nil {
		zap.S().Errorw("更新品牌失败", err.Error())
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "更新品牌成功",
	})

}

func List(ctx *gin.Context) {
	page := ctx.DefaultQuery("page", "1")
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"msg": "无法找到此页面",
		})
	}
	rsp, err := global.GoodsSrvClient.BrandList(context.Background(), &proto.BrandFilterRequest{
		Pages:       int32(pageInt),
		PagePerNums: 5,
	})
	if err != nil {
		zap.S().Errorw("获取品牌列表失败", err.Error())
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	reMap := gin.H{}
	reMap["total"] = rsp.Total
	reList := make([]interface{}, 0)
	for _, data := range rsp.Data {
		tmpMap := gin.H{}
		tmpMap["id"] = data.Id
		tmpMap["name"] = data.Name
		tmpMap["logo"] = data.Logo
		reList = append(reList, tmpMap)
	}
	reMap["data"] = reList
	ctx.JSON(http.StatusOK, reMap)
}
