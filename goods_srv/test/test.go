package main

import (
	"context"
	"fmt"

	"grpc/goods_srv/proto"

	_ "github.com/mbobakov/grpc-consul-resolver"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var userConn *grpc.ClientConn
var goodsSrvClient proto.GoodsClient

func main() {
	InitSrvConn()
	testCreateGoods()
}

func testCreateGoods() {
	Rsp, err := goodsSrvClient.CreateGoods(context.Background(), &proto.CreateGoodsInfo{
		Id:              0001,
		Name:            "商品1",
		GoodsSn:         "",
		Stocks:          10,
		MarketPrice:     10,
		ShopPrice:       10,
		GoodsBrief:      "商品1",
		GoodsDesc:       "商品1",
		ShipFree:        false,
		Images:          nil,
		DescImages:      nil,
		GoodsFrontImage: "",
		IsNew:           false,
		IsHot:           false,
		OnSale:          false,
		CategoryId:      4,
		BrandId:         2,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(Rsp.CategoryId, Rsp.Brand)
}
func testCategoryList() {
	Rsp, err := goodsSrvClient.CategoryBrandList(context.Background(), &proto.CategoryBrandFilterRequest{
		Pages:       1,
		PagePerNums: 3,
	})
	if err != nil {
		panic(err)
	}
	for _, brand := range Rsp.Data {
		fmt.Println(brand)
	}
}
func testBrandList() {
	Rsp, err := goodsSrvClient.BrandList(context.Background(), &proto.BrandFilterRequest{
		Pages:       1,
		PagePerNums: 1,
	})
	if err != nil {
		panic(err)
	}
	for _, brand := range Rsp.Data {
		fmt.Println(brand)
	}
}
func testCreateBrand() {
	rsp, err := goodsSrvClient.CreateBrand(context.Background(), &proto.BrandRequest{
		Id:   2,
		Name: "2",
		Logo: "2",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Id)
}
func testCreateCategory() {
	rsp, err := goodsSrvClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{
		Name:           "1",
		ParentCategory: 0,
		Level:          1,
		IsTab:          false,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Id)
}
func InitSrvConn() {
	//用了grpc-consul-resolver库，将grpc直接连接到consul，而不再先从consul中读服务地址，再连接服务
	//连接就交给grpc来维护，实际上每次通过客户端调用grpc服务，这个连接会动态的变化，连接到不同的服务
	var err error
	userConn, err = grpc.Dial(
		fmt.Sprintf(
			"consul://192.168.224.128:8500/goods_srv?wait=14s",
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatal("[InitSrvConn]【连接服务失败】")
	} else {
		fmt.Println(userConn.GetState())
	}

	goodsSrvClient = proto.NewGoodsClient(userConn)
}
