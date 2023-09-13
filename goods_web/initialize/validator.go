package initialize

import (
	"fmt"
	"reflect"
	"strings"

	"grpc/goods_web/global"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	ent "github.com/go-playground/validator/v10/translations/en"
	zht "github.com/go-playground/validator/v10/translations/zh"
)

func InitTrans(locale string) (err error) {
	//修改gin框架中的validator引擎的翻译引擎
	if v, ok := (binding.Validator.Engine()).(*validator.Validate); ok {
		//注册一个获取json的tag的自定义方法
		v.RegisterTagNameFunc(func(field reflect.StructField) string {
			//格式 `json:"abc,def"` 获取到的就是""里的string，第一个是json名字
			name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
			//field.tag.get如果json为空，返回的是"-"
			if name == "-" {
				return ""
			}
			return name
		})

		zhT := zh.New()
		enT := en.New()

		//第一个参数是备用语言，后面是支持语言
		uni := ut.New(enT, zhT, enT)
		//传入的locale应该是zh 或者 en等
		global.Trans, ok = uni.GetTranslator(locale)
		if !ok {
			return fmt.Errorf("error to uni.GetTranslator(%s)", locale)
		}
		switch locale {
		case "en":
			_ = ent.RegisterDefaultTranslations(v, global.Trans)
		case "zh":
			_ = zht.RegisterDefaultTranslations(v, global.Trans)
		default:
			_ = ent.RegisterDefaultTranslations(v, global.Trans)
		}
		return
	}
	return err
}
