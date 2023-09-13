package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"grpc/user_web/forms"
	"grpc/user_web/global"
	"grpc/user_web/global/response"
	"grpc/user_web/middlewares"
	"grpc/user_web/proto"
	"grpc/user_web/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetUserList 获取用户列表接口
func GetUserList(ctx *gin.Context) {

	//获取用户列表时先获取到访问者的身份
	claim, _ := ctx.Get("claims")
	currentUser := claim.(*middlewares.MyClaims)
	zap.S().Infof("访问用户：%d", currentUser.ID)

	pn := ctx.DefaultQuery("pn", "0")
	pnInt, _ := strconv.Atoi(pn)
	pSize := ctx.DefaultQuery("psize", "10")
	pSizeInt, _ := strconv.Atoi(pSize)

	rsp, err := global.UserSrvClient.GetUserList(context.Background(), &proto.PageInfo{
		Pn:    uint32(pnInt),
		PSize: uint32(pSizeInt),
	})
	if err != nil {
		zap.S().Errorw("[GetUserList]【查询用户列表失败】")
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	result := make([]interface{}, 0)
	for _, value := range rsp.Data {

		userRsp := response.UserResponse{
			Id:       value.Id,
			Mobile:   value.Mobile,
			NickName: value.NickName,
			Birthday: response.JsonTime(time.Unix(int64(value.Birthday), 0)),
			Gender:   value.Gender,
			Role:     value.Role,
		}

		result = append(result, userRsp)
	}
	ctx.JSON(http.StatusOK, result)
}

// PasswordLogin 密码登录接口
func PasswordLogin(c *gin.Context) {
	//表单验证
	passwordLoginForm := &forms.PasswordLoginForm{}
	if err := c.ShouldBind(&passwordLoginForm); err != nil {
		utils.HandleValidatorError(c, err)
		return
	}

	//验证码验证
	if !store.Verify(passwordLoginForm.CaptchaId, passwordLoginForm.Captcha, false) {
		c.JSON(http.StatusBadRequest, gin.H{
			"captcha": "验证码错误",
		})
		return
	}

	//具体登录逻辑
	rsp, err := global.UserSrvClient.GetUserByMobile(context.Background(), &proto.MobileRequest{Mobile: passwordLoginForm.Mobile})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				c.JSON(http.StatusBadRequest, map[string]string{
					"mobile": "用户不存在",
				})
			default:
				c.JSON(http.StatusInternalServerError, map[string]string{
					"mobile": "登录失败",
				})
			}
			return
		}
		//获取到信息还要进一步进行密码验证
	} else {
		PassRsp, PassErr := global.UserSrvClient.CheckPassword(context.Background(), &proto.CheckPasswordInfo{
			Password:          passwordLoginForm.Password,
			EncryptedPassword: rsp.Password,
		})
		if PassErr != nil {
			//用户服务调用失败，内部错误
			c.JSON(http.StatusInternalServerError, map[string]string{
				"msg": "登录失败",
			})
		} else {
			//如果调用成功，判断密码逻辑
			if PassRsp.Success == true {
				//生成Token返回
				j := middlewares.NewJWT()
				claim := middlewares.MyClaims{
					ID:       int64(rsp.Id),
					NickName: rsp.NickName,
					Role:     uint(rsp.Role),
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    "xyz",
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(middlewares.TokenExpireDuration)),
					},
				}
				token, err := j.GenToken(claim)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"msg": "生成token失败",
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"id":         rsp.Id,
					"nick_name":  rsp.NickName,
					"token":      token,
					"expired_at": time.Now().Add(middlewares.TokenExpireDuration).Unix(),
				})
			} else {
				c.JSON(http.StatusBadRequest, map[string]string{
					"password": "密码错误",
				})
			}
		}
	}

}

// Register 用户注册接口
func Register(ctx *gin.Context) {
	registerForm := &forms.RegisterForm{}
	if err := ctx.ShouldBind(&registerForm); err != nil {
		utils.HandleValidatorError(ctx, err)
		return
	}

	//验证码校验(redis)
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf(
			"%s:%d",
			global.ServerConfig.RedisInfo.Host,
			global.ServerConfig.RedisInfo.Port,
		),
	})
	value, err := rdb.Get(context.Background(), registerForm.Mobile).Result()
	if err != nil {
		if err == redis.Nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": "验证码错误",
			})
			return
		}
		zap.S().Errorw("[Register] redis读取失败", "msg", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "注册失败",
		})
	} else {
		if value != registerForm.Code {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": "验证码错误",
			})
			return
		}
	}

	user, err := global.UserSrvClient.CreateUser(context.Background(), &proto.CreateUserInfo{
		NickName: registerForm.Mobile,
		Password: registerForm.Password,
		Mobile:   registerForm.Mobile,
	})

	if err != nil {
		zap.S().Errorw("[Register 【新建用户失败】]", "msg", err.Error())
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	//生成Token返回, 注册完成后直接赋予登录状态
	j := middlewares.NewJWT()
	claim := middlewares.MyClaims{
		ID:       int64(user.Id),
		NickName: user.NickName,
		Role:     uint(1),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "xyz",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(middlewares.TokenExpireDuration)),
		},
	}
	token, err := j.GenToken(claim)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "生成token失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":         user.Id,
		"nick_name":  user.NickName,
		"token":      token,
		"expired_at": time.Now().Add(middlewares.TokenExpireDuration).Unix(),
	})
}

func Update(ctx *gin.Context) {
	userId, ok := ctx.Get("userID")
	if !ok {
		ctx.JSON(http.StatusNotFound, gin.H{
			"msg": "获取不到用户信息",
		})
		return
	}
	updateForm := forms.UpdateUserInfoForm{}
	err := ctx.ShouldBindJSON(&updateForm)
	fmt.Println(updateForm)
	if err != nil {
		zap.S().Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "无效参数",
		})
		return
	}
	userInfo, err := global.UserSrvClient.GetUserById(context.Background(), &proto.IdRequest{Id: int32(userId.(int64))})
	if err != nil {
		zap.S().Errorw("获取不到用户信息", err.Error())
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	fmt.Println(userInfo)
	date, err := time.Parse("2006-01-02", updateForm.Birthday)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "时间错误",
		})
		return
	}
	var nickname string
	var gender string
	if updateForm.Nickname == "" {
		nickname = userInfo.NickName
	} else {
		nickname = updateForm.Nickname
	}
	if updateForm.Gender == "" {
		gender = userInfo.Gender
	} else {
		gender = updateForm.Gender
	}
	fmt.Println(nickname, gender)
	_, err = global.UserSrvClient.UpdateUser(context.Background(), &proto.UpdateUserInfo{
		Id:       int32(userId.(int64)),
		NickName: nickname,
		Gender:   gender,
		Birthday: uint64(date.Unix()),
	})
	if err != nil {
		zap.S().Error(err)
		utils.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "更改信息成功",
	})
}
