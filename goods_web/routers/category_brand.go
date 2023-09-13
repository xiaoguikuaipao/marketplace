package routers

import (
	"grpc/goods_web/api/category_brand"

	"github.com/gin-gonic/gin"
)

func InitCB(Router *gin.RouterGroup) {
	CategoryBrandRouter := Router.Group("categorybrands")
	{
		CategoryBrandRouter.GET("/list", category_brand.CategoryBrandList)     // 类别品牌列表页
		CategoryBrandRouter.DELETE("/:id", category_brand.DeleteCategoryBrand) // 删除类别品牌
		CategoryBrandRouter.POST("", category_brand.NewCategoryBrand)          //新建类别品牌
		CategoryBrandRouter.PUT("/:id", category_brand.UpdateCategoryBrand)    //修改类别品牌
		CategoryBrandRouter.GET("/:id", category_brand.GetCategoryBrandList)   //获取分类的品牌
	}

}
