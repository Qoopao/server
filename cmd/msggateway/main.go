package main

import (
	"context"
	"log"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/roc/roc-foundation-util-go/log/otel"
	"github.com/roc/roc-im-server/internal/msggateway"
)

func main() {
	ctx := context.Background()
	log.Println("启动 MsgGateway 服务...")

	// 使用 foundation-util-go 统一初始化 OTEL (Trace + Log)
	kit, err := otel.InitOTEL(ctx,
		otel.WithServiceName("msggateway"),
		otel.WithEndpoint("localhost:4317"),
		otel.WithInsecure(true),
	)
	if err != nil {
		log.Fatalf("Failed to init OTEL: %v", err)
	}
	defer kit.Shutdown(ctx)

	// 设置日志级别为 DEBUG
	kit.Logger.SetLevel(klog.LevelDebug)

	// 创建 WebSocket 服务器
	wsServer := msggateway.NewWsServer(
		msggateway.WithPort(10010),
		msggateway.WithMaxConnNum(10000),
		msggateway.WithWriteBufferSize(1000),
		msggateway.WithHandshakeTimeout(1000),
		msggateway.WithMessageMaxMsgLength(10000),
	)

	// 启动服务
	klog.Info("MsgGateway 启动中...")
	klog.Infof("MsgGateway 监听端口: %d", 10010)
	if err := wsServer.Run(context.Background()); err != nil {
		klog.Fatalf("MsgGateway 启动失败: %v", err)
	}
}
