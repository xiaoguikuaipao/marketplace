package routers

import (
	"grpc/goods_web/api/brand"
	"grpc/goods_web/middlewares"

	"github.com/gin-gonic/gin"
)

func InitBrandsRouter(Router *gin.RouterGroup) {
	BrandsRouter := Router.Group("brands")

	{
		BrandsRouter.GET("/list", brand.List)
		BrandsRouter.POST("/brand", middlewares.JWTAuth(), middlewares.IsAdminAuth(), brand.New)
		BrandsRouter.DELETE("/:id", middlewares.JWTAuth(), middlewares.IsAdminAuth(), brand.Delete)

		BrandsRouter.PATCH("/:id", middlewares.JWTAuth(), middlewares.IsAdminAuth(), brand.Update)
	}
}
