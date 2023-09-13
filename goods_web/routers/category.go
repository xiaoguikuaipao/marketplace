package routers

import (
	"grpc/goods_web/api/category"
	"grpc/goods_web/middlewares"

	"github.com/gin-gonic/gin"
)

func InitCategoryRouter(Router *gin.RouterGroup) {
	CategoryRouter := Router.Group("categorys")
	{
		CategoryRouter.GET("/", category.List)                                // 商品类别列表页
		CategoryRouter.DELETE("/:id", middlewares.JWTAuth(), category.Delete) // 删除分类
		CategoryRouter.GET("/:id", category.Detail)                           // 获取分类详情
		CategoryRouter.POST("", middlewares.JWTAuth(), category.New)          //新建分类
		CategoryRouter.PUT("/:id", middlewares.JWTAuth(), category.Update)    //修改分类信息
	}
}
