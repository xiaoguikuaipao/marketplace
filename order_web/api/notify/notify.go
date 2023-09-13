package notify

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Notify(c *gin.Context) {
	// 生成支付宝客户端

	// var noti, err = client.GetTradeNotification(c.Request)
	//if err != nil ...

	//global.OrderSrvClient.UpdateOrderStatus(context.Background(), &proto.OrderStatus{
	//	OrderSn: noti.OutTradeNo,
	//	Status:  noti.TradeStatus,
	//})

	// 通知支付宝
	c.String(http.StatusOK, "success")
}
