package global

import (
	"grpc/order_web/config"
	"grpc/order_web/proto"

	ut "github.com/go-playground/universal-translator"
	"github.com/spf13/viper"
)

var (
	ServerConfig    = &config.ServerConfig{}
	V               *viper.Viper
	Trans           ut.Translator
	GoodsSrvClient  proto.GoodsClient
	OrderSrvClient  proto.OrderClient
	InventoryClient proto.InventoryClient
)
