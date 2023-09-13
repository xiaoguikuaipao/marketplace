package initialize

import (
	"net/http"

	"grpc/goods_web/middlewares"
	"grpc/goods_web/routers"

	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"success": true,
		})
	})
	{
		//配置跨域请求中间件
		r.Use(middlewares.Cors())

		ApiGroup := r.Group("v1")
		routers.InitGoodsRouter(ApiGroup)
		routers.InitBrandsRouter(ApiGroup)
		routers.InitCategoryRouter(ApiGroup)
		routers.InitCB(ApiGroup)
	}
	return r
}
