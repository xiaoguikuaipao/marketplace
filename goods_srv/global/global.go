package global

import (
	"grpc/goods_srv/config"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	DB           *gorm.DB
	V            *viper.Viper
	ServerConfig = new(config.ServerConfig)
)
