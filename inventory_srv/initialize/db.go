package initialize

import (
	"fmt"
	"log"
	"os"
	"time"

	"grpc/inventory_srv/global"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func InitDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		global.ServerConfig.MysqlInfo.User,
		global.ServerConfig.MysqlInfo.Password,
		global.ServerConfig.MysqlInfo.Host,
		global.ServerConfig.MysqlInfo.Port,
		global.ServerConfig.MysqlInfo.Name)
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			Colorful:      true,
			LogLevel:      logger.Info,
		})
	var err error
	global.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: true,
			NameReplacer:  nil,
			NoLowerCase:   false,
		},
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}
}
