package initialize

import (
	"fmt"

	"grpc/goods_web/global"
	"grpc/goods_web/proto"

	_ "github.com/mbobakov/grpc-consul-resolver"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitSrvConn() {
	//用了grpc-consul-resolver库，将grpc直接连接到consul，而不再先从consul中读服务地址，再连接服务
	//连接就交给grpc来维护，实际上每次通过客户端调用grpc服务，这个连接会动态的变化，连接到不同的服务
	userConn, err := grpc.Dial(
		fmt.Sprintf(
			"consul://%s:%d/%s?wait=14s",
			global.ServerConfig.ConsulInfo.Host,
			global.ServerConfig.ConsulInfo.Port,
			global.ServerConfig.GoodsSrvInfo.Name,
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatal("[InitSrvConn]【连接商品服务失败】")
	}

	goodsSrvClient := proto.NewGoodsClient(userConn)
	global.GoodsSrvClient = goodsSrvClient
}
