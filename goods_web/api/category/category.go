package category

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"grpc/goods_web/forms"
	"grpc/goods_web/global"
	"grpc/goods_web/proto"
	"grpc/goods_web/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

func List(ctx *gin.Context) {
	rsp, err := global.GoodsSrvClient.GetAllCategorysList(context.Background(), &emptypb.Empty{})
	if err != nil {
		zap.S().Error(err)
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	data := make([]interface{}, 0)
	err = json.Unmarshal([]byte(rsp.JsonData), &data)
	if err != nil {
		zap.S().Error(err.Error())
	}
	ctx.JSON(http.StatusOK, data)

}
func Detail(ctx *gin.Context) {
	id := ctx.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	reMap := gin.H{}
	subCategoies := make([]interface{}, 0)
	rsp, err := global.GoodsSrvClient.GetSubCategory(context.Background(), &proto.CategoryListRequest{
		Id: int32(i),
	})
	if err != nil {
		zap.S().Error(err.Error())
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	for _, value := range rsp.SubCategorys {
		subCategoies = append(subCategoies, gin.H{
			"id":              value.Id,
			"name":            value.Name,
			"is_tab":          value.IsTab,
			"level":           value.Level,
			"parent_category": value.ParentCategory,
		})
	}
	reMap["id_tab"] = rsp.Info.IsTab
	reMap["id"] = rsp.Info.Id
	reMap["level"] = rsp.Info.Level
	reMap["name"] = rsp.Info.Name
	reMap["parent_category"] = rsp.Info.ParentCategory
	reMap["sub_category"] = subCategoies
	ctx.JSON(http.StatusOK, reMap)

}

func Update(ctx *gin.Context) {
	id := ctx.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"err": err.Error(),
		})
		return
	}
	updateForm := forms.CategoryForm{}
	err = ctx.BindJSON(&updateForm)
	if err != nil {
		utils.HandleValidatorError(ctx, err)
		zap.S().Error(err.Error())
		return
	}
	_, err = global.GoodsSrvClient.UpdateCategory(context.Background(), &proto.CategoryInfoRequest{
		Id:             int32(i),
		Name:           updateForm.Name,
		ParentCategory: updateForm.ParentCategory,
		Level:          updateForm.Level,
		IsTab:          *updateForm.IsTab,
	})
	if err != nil {
		zap.S().Error(err.Error())
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "更新分类成功",
	})
}

func Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"err": err.Error(),
		})
		return
	}
	_, err = global.GoodsSrvClient.DeleteCategory(context.Background(), &proto.DeleteCategoryRequest{Id: int32(i)})
	if err != nil {
		zap.S().Error(err.Error())
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "删除成功",
	})
}

func New(ctx *gin.Context) {
	categoryForm := forms.CategoryForm{}
	err := ctx.BindJSON(&categoryForm)
	if err != nil {
		zap.S().Error(err.Error())
		utils.HandleValidatorError(ctx, err)
		return
	}
	rsp, err := global.GoodsSrvClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{
		Name:           categoryForm.Name,
		ParentCategory: categoryForm.ParentCategory,
		Level:          categoryForm.Level,
		IsTab:          *categoryForm.IsTab,
	})
	if err != nil {
		utils.HandleGrpcErrorToHttp(err, ctx)
		zap.S().Error(err.Error())
		return
	}
	reMap := gin.H{}
	reMap["name"] = rsp.Name
	ctx.JSON(http.StatusOK, reMap)
}
