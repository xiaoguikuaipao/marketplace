package api

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"grpc/user_web/forms"
	"grpc/user_web/global"
	"grpc/user_web/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func GenerateSmsCode(width int) string {
	//生成width长度的验证码
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	rand.NewSource(time.Now().Unix())

	var sb strings.Builder
	for i := 0; i < width; i++ {
		_, _ = fmt.Fprintf(&sb, "%d", numeric[rand.Intn(r)])
	}
	return sb.String()
}

func SendSms(ctx *gin.Context) {
	//表单验证，验证传入的手机号
	sendSmsForm := forms.SendSmsForm{}
	if err := ctx.ShouldBind(&sendSmsForm); err != nil {
		utils.HandleValidatorError(ctx, err)
		return
	}

	code := GenerateSmsCode(6)
	mobile := sendSmsForm.Mobile

	//将手机号作为key，code作为value存入redis中
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf(
			"%s:%d",
			global.ServerConfig.RedisInfo.Host,
			global.ServerConfig.RedisInfo.Port,
		),
	})
	rdb.Set(context.Background(), mobile, code, 300*time.Second)

	ctx.JSON(http.StatusOK, gin.H{
		"msg": "验证短信发送成功",
	})
}
