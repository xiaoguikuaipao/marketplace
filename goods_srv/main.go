package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"grpc/goods_srv/global"
	"grpc/goods_srv/handler"
	"grpc/goods_srv/initialize"
	"grpc/goods_srv/proto"
	"grpc/goods_srv/utils"
	"grpc/goods_srv/utils/register/consul"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {

	//初始化配置文件，DB，日志
	initialize.InitConfig()
	initialize.InitDB()
	initialize.InitLogger()
	zap.S().Info(global.ServerConfig)
	//获得网络ip地址
	interfaceName := "WLAN"
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		panic(err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		panic(err)
	}
	currentAddr := ""
	currentPort, err := utils.GetFreePort()
	if err != nil {
		panic("端口获取失败")
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				currentAddr = ipnet.IP.String()
				zap.S().Infof("当前服务地址为%s:%d", currentAddr, currentPort)
			}
		}
	}
	if currentAddr == "" {
		panic("获取不到当前IP地址")
	}

	//注册grpc的服务端
	server := grpc.NewServer()
	proto.RegisterGoodsServer(server, &handler.GoodsServer{})
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", currentPort))
	if err != nil {
		panic("failed to listen: " + err.Error())
	}
	//注册服务健康检查
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	//将服务注册到consul上
	register := consul.NewRegistryClient(global.ServerConfig.ConsuIInfo.Host, global.ServerConfig.ConsuIInfo.Port)
	if err := register.Register(currentAddr, currentPort, global.ServerConfig.Tags, currentPort, global.ServerConfig.Name); err != nil {
		panic("服务注册consul失败")
	}
	go func() {
		err = server.Serve(listen)
		if err != nil {
			panic("failed to start grpc: " + err.Error())
		}
	}()
	// 接收终止信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if err = register.DeRegister(strconv.Itoa(currentPort)); err != nil {
		zap.S().Info("注销失败")
	}
	zap.S().Info("注销成功")
}
