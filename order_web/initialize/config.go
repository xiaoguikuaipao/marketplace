package initialize

import (
	"grpc/order_web/global"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func InitConfig() {
	global.V = viper.New()
	global.V.SetConfigFile(`D:\code\grpc\order_web\config.yaml`)
	if err := global.V.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := global.V.Unmarshal(global.ServerConfig); err != nil {
		panic(err)
	}
	zap.S().Infof("配置信息: %v", *global.ServerConfig)

	global.V.WatchConfig()
	global.V.OnConfigChange(func(in fsnotify.Event) {
		zap.S().Infof("配置信息发生变更...")
		_ = global.V.ReadInConfig()
		_ = global.V.Unmarshal(global.ServerConfig)
		zap.S().Infof("配置信息: %v", *global.ServerConfig)
	})
}
