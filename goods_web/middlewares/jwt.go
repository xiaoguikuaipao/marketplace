package middlewares

import (
	"errors"
	"net/http"
	"time"

	"grpc/goods_web/global"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const TokenExpireDuration = 2 * time.Hour

type MyClaims struct {
	jwt.RegisteredClaims
	ID       int64  `json:"id"`
	NickName string `json:"nick_name"`
	Role     uint   `json:"role"`
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("x-token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, map[string]string{
				"msg": "请登录",
			})
			c.Abort()
			return
		}

		j := NewJWT()
		// parseToken 解析token包含的信息
		claims, err := j.ParseToken(token)
		if err != nil {
			if err == TokenInvalid {
				c.JSON(http.StatusUnauthorized, map[string]string{
					"msg": "无效的授权",
				})
				c.Abort()
				return
			}

			c.JSON(http.StatusUnauthorized, "身份信息已失效，请重新登录")
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Set("userID", claims.ID)
		c.Next()
	}
}

type JWT struct {
	SigningKey []byte
}

var (
	TokenInvalid = errors.New("token is invalid")
)

func NewJWT() *JWT {
	jwtIns := JWT{SigningKey: []byte(global.ServerConfig.JWTInfo.SigningKey)}
	return &jwtIns
}

func (j *JWT) GenToken(myclaim MyClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, myclaim)
	return token.SignedString(j.SigningKey)
}

func (j *JWT) ParseToken(tokenString string) (*MyClaims, error) {
	var mc = new(MyClaims)
	token, err := jwt.ParseWithClaims(tokenString, mc, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		return nil, err
	}
	if token != nil {
		if claim, ok := token.Claims.(*MyClaims); ok && token.Valid {
			return claim, err
		}
		return nil, TokenInvalid
	} else {
		return nil, TokenInvalid
	}
}
