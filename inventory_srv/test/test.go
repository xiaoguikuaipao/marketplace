package main

import (
	"context"
	"fmt"

	"grpc/inventory_srv/proto"

	_ "github.com/mbobakov/grpc-consul-resolver"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var userConn *grpc.ClientConn
var inventorySrvClient proto.InventoryClient

func main() {
	InitSrvConn()
	TestSetInv()
}
func InitSrvConn() {
	//用了grpc-consul-resolver库，将grpc直接连接到consul，而不再先从consul中读服务地址，再连接服务
	//连接就交给grpc来维护，实际上每次通过客户端调用grpc服务，这个连接会动态的变化，连接到不同的服务
	var err error
	userConn, err = grpc.Dial(
		fmt.Sprintf(
			"consul://192.168.224.128:8500/inventory_srv?wait=14s",
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatal("[InitSrvConn]【连接服务失败】")
	} else {
		fmt.Println(userConn.GetState())
	}

	inventorySrvClient = proto.NewInventoryClient(userConn)
}

func TestSetInv() {
	var i int32
	for i = 1; i <= 4; i++ {
		_, err := inventorySrvClient.SetInv(context.Background(), &proto.GoodsInvInfo{
			GoodsId: i,
			Num:     100,
		})
		if err != nil {
			fmt.Println("设置库存成功")
		}
	}
}
