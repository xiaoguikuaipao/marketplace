package initialize

import (
	"grpc/order_srv/global"
	"grpc/order_srv/handler"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"go.uber.org/zap"
)

func InitProducer() {

	p, err := rocketmq.NewTransactionProducer(&handler.OL, producer.WithNameServer([]string{"192.168.224.128:9876"}), producer.WithInstanceName("归还消息生产者"))
	if err != nil {
		zap.L().Panic("创建归还生产者失败", zap.Error(err))
		return
	}
	global.TP = p

	delay, err := rocketmq.NewProducer(producer.WithNameServer([]string{"192.168.224.128:9876"}), producer.WithInstanceName("延时消息生产者"))
	if err != nil {
		zap.L().Panic("启动延时消息生产者失败", zap.Error(err))
		return
	}
	global.P = delay
}
