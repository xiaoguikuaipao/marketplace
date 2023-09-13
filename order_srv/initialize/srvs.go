package initialize

import (
	"fmt"

	"grpc/order_srv/global"
	"grpc/order_srv/proto"

	_ "github.com/mbobakov/grpc-consul-resolver"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitSrvs() {
	//用了grpc-consul-resolver库，将grpc直接连接到consul，而不再先从consul中读服务地址，再连接服务
	//连接就交给grpc来维护，实际上每次通过客户端调用grpc服务，这个连接会动态的变化，连接到不同的服务
	userConn, err := grpc.Dial(
		fmt.Sprintf(
			"consul://%s:%d/%s?wait=14s",
			global.ServerConfig.ConsuIInfo.Host,
			global.ServerConfig.ConsuIInfo.Port,
			global.ServerConfig.GoodsSrvInfo.Name,
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.L().Fatal("[InitSrvConn]【连接商品服务失败】", zap.Error(err))
		return
	}

	goodsSrvClient := proto.NewGoodsClient(userConn)
	global.GoodsSrvClient = goodsSrvClient

	invConn, err := grpc.Dial(
		fmt.Sprintf(
			"consul://%s:%d/%s?wait=14s",
			global.ServerConfig.ConsuIInfo.Host,
			global.ServerConfig.ConsuIInfo.Port,
			global.ServerConfig.InventorySrvInfo.Name,
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatal("[InitSrvConn]【连接库存服务失败】")
	}

	inventorySrvClient := proto.NewInventoryClient(invConn)
	global.InventorySrvClient = inventorySrvClient
}
