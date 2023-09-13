package handler

import (
	"context"
	"encoding/json"

	"grpc/order_srv/global"
	"grpc/order_srv/models"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"go.uber.org/zap"
)

func OrderTimeout(ctx context.Context, ext ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	//注意幂等性
	for i := range ext {
		var order models.OrderInfo
		_ = json.Unmarshal(ext[i].Body, &order)
		var orderRes models.OrderInfo
		if result := global.DB.Model(&models.OrderInfo{}).Where(&models.OrderInfo{OrderSn: order.OrderSn}).First(&orderRes); result.RowsAffected == 0 {
			return consumer.ConsumeSuccess, nil
		}
		if orderRes.Status != "TRADE_SUCCESS" {
			//转发reback消息
			tx := global.DB.Begin()
			orderRes.Status = "TRADE_CLOSED"
			tx.Save(&orderRes)
			p, err := rocketmq.NewProducer(producer.WithNameServer([]string{"192.168.224.128:9876"}))
			if err != nil {
				tx.Rollback()
				zap.L().Error("order转发者启动失败", zap.Error(err))
				return consumer.ConsumeRetryLater, err
			}
			if err = p.Start(); err != nil {
				tx.Rollback()
				zap.L().Error("order转发者启动失败", zap.Error(err))
				return consumer.ConsumeRetryLater, err
			}

			_, err = p.SendSync(context.Background(), primitive.NewMessage("order_reback", ext[i].Body))
			if err != nil {
				tx.Rollback()
				zap.L().Error("转发消息失败", zap.Error(err))
				return consumer.ConsumeRetryLater, err
			}

			tx.Commit()
		}
		return consumer.ConsumeSuccess, nil
	}
	return consumer.ConsumeSuccess, nil
}
