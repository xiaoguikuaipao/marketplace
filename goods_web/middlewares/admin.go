package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func IsAdminAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, _ := ctx.Get("claims")
		currentUser := claims.(*MyClaims)

		if currentUser.Role == 2 {
			ctx.JSON(http.StatusForbidden, gin.H{
				"msg": "权限不足",
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
