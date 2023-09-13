package global

import (
	"grpc/order_srv/config"
	"grpc/order_srv/proto"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	DB                 *gorm.DB
	V                  *viper.Viper
	ServerConfig       = new(config.ServerConfig)
	GoodsSrvClient     proto.GoodsClient
	InventorySrvClient proto.InventoryClient
	TP                 rocketmq.TransactionProducer
	P                  rocketmq.Producer
)
