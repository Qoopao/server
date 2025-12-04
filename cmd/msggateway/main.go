package main

import (
	"context"
	"log"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"github.com/roc/roc-im-server/internal/msggateway"
	"github.com/roc/roc-im-server/tools/klog_otel"
)

func main() {
	ctx := context.Background()
	log.Println("启动 MsgGateway 服务...")

	// 初始化 OTEL Trace Provider
	p := provider.NewOpenTelemetryProvider(
		provider.WithServiceName("msggateway"),
		provider.WithExportEndpoint("localhost:4317"),
		provider.WithInsecure(),
	)
	defer p.Shutdown(ctx)

	// 初始化 OTEL Log Provider
	otelLogger, lp, err := klog_otel.NewKitexLoggerWithConfig(ctx, "msggateway", "localhost:4317", true)
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	defer lp.Shutdown(ctx)

	// 设置日志级别为 DEBUG
	otelLogger.SetLevel(klog.LevelDebug)

	klog.SetLogger(otelLogger)

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
