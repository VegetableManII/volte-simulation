package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	. "volte/common"

	"github.com/spf13/viper"
	"github.com/wonderivan/logger"
)

var (
	cscfConn, eNodeBConn *net.UDPConn
)

/*
	读协程读消息->解析前管道->协议解析->解析后管道->写协程写消息
		readGoroutine --->> chan *Msg --->> parser --->> chan *Msg --->> writeGoroutine
*/
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, CtxString("Entity"), "PGW")
	coreIC := make(chan *Msg, 2)
	coreOC := make(chan *Msg, 2)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	// 开启与eNodeB交互的协程
	go TransaportWithClient(ctx, eNodeBConn, coreIC, coreOC)
	// go common.ExchangeWithClient(ctx, eNodeBConn, coreIC, coreOC) // debug
	// 开启与IMS p-cscf交互的协程

	<-quit
	logger.Warn("[PGW] pgw 功能实体退出...")
	cancel()
	logger.Warn("[PGW] pgw 子协程退出完成...")
}

func init() {
	viper.SetConfigName("config.yml")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".") // 设置配置文件与可执行文件在同一目录可供编译后的程序使用
	if e := viper.ReadInConfig(); e != nil {
		log.Panicln("配置文件读取失败", e)
	}
	host := viper.GetString("EPS.pgw.host")
	// cscf := viper.GetString("IMS.p-cscf.host")
	logger.Info("配置文件读取成功", "")
	// 启动 PGW 的UDP服务器
	eNodeBConn = InitServer(host)
	// 创建连接 CSCF 的客户端
	// cscfConn = ConnectServer(cscf)
}