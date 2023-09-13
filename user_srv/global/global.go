package global

import (
	"grpc/user_srv/config"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	DB           *gorm.DB
	V            *viper.Viper
	ServerConfig *config.ServerConfig = new(config.ServerConfig)
)
