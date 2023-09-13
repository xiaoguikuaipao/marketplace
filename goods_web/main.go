package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"grpc/goods_web/global"
	"grpc/goods_web/initialize"
	"grpc/goods_web/utils/register/consul"

	"go.uber.org/zap"
)

func main() {
	// 初始化zap日志, 日志级别debug, info, warn, error, fatal
	initialize.InitLogger()

	// 初始化router
	Router := initialize.Routers()

	//初始化配置文件
	initialize.InitConfig()

	//初始化用户服务Grpc客户端连接
	initialize.InitSrvConn()

	//初始化翻译器
	if err := initialize.InitTrans("zh"); err != nil {
		panic(err)
	}
	//获得ip地址
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
	if err != nil {
		panic("端口获取失败")
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				currentAddr = ipnet.IP.String()
			}
		}
	}
	if currentAddr == "" {
		panic("获取不到当前IP地址")
	}

	registerClient := consul.NewRegistryClient(global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port)
	err = registerClient.Register(currentAddr, global.ServerConfig.Port, global.ServerConfig.Name, global.ServerConfig.Tags, strconv.Itoa(global.ServerConfig.Port))
	if err != nil {
		panic("服务注册失败")
	}

	port := global.V.GetInt("port")
	zap.S().Infof("启动服务器，端口: %d", port)
	go func() {
		if err := Router.Run(fmt.Sprintf(":%d", port)); err != nil {
			zap.S().Panic("启动失败", zap.Error(err))
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if err := registerClient.Deregister(strconv.Itoa(global.ServerConfig.Port)); err != nil {
		zap.S().Infof("注销失败%s", err.Error())
	} else {
		zap.S().Info("注销成功")
	}
}
