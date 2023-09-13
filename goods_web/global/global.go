package global

import (
	"grpc/goods_web/config"
	"grpc/goods_web/proto"

	ut "github.com/go-playground/universal-translator"
	"github.com/spf13/viper"
)

var (
	ServerConfig   = &config.ServerConfig{}
	V              *viper.Viper
	Trans          ut.Translator
	GoodsSrvClient proto.GoodsClient
)
