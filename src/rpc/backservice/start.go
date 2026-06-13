package backservice

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/cloudwego/kitex/client"
	kitexdiscovery "github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/klog"
	kitexregistry "github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	backservice "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/back/backservice"
	backbon "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/backbon"
	backbonservice "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/backbon/backbonservice"
	"github.com/rhp-QE/roc-foundation-util-go/log/otel"
	"github.com/rhp-QE/roc-foundation-util-go/network"
	"github.com/rhp-QE/roc-im-server/src/common/kitexinfra"
	consts "github.com/rhp-QE/roc-im-server/src/const"
	deps "github.com/rhp-QE/roc-im-server/src/rpc/backservice/infrastructure/deps"
	rpcadapter "github.com/rhp-QE/roc-im-server/src/rpc/backservice/interfaces/rpc"
)

const (
	backbonRegisterMaxAttempts = 30
	backbonRegisterInterval    = time.Second
	backbonRegisterTimeout     = 5 * time.Second
)

// Start 启动 BackService 服务器
func Start() {
	// 初始化 OTEL
	kit := initOTEL()
	defer kit.Shutdown(context.Background())

	// 获取配置信息
	port := getServicePort()
	localIP := getLocalIP()

	// Kitex resolver 负责下游服务发现；服务端注册由 Kitex server 管理。
	resolver, err := kitexinfra.NewEtcdResolverFromEnv()
	if err != nil {
		klog.Fatalf("Failed to create kitex etcd resolver: %v", err)
	}

	// 创建服务上下文
	serviceCtx := deps.NewServiceContext(resolver)
	defer serviceCtx.Close()

	// 创建服务器
	svr := createBackServiceServer(localIP, port, serviceCtx)

	// 启动服务器
	startServer(svr)

	// 注册业务路由到 backbon-service；这不是服务发现基础能力，保留业务注册语义。
	go registerToBackbonServiceWithRetry(context.Background(), resolver)

	klog.Infof("BackService started, address: %s:%d", localIP, port)

	// 等待关闭信号
	waitForShutdown()

	if err := svr.Stop(); err != nil {
		klog.Errorf("BackService stop failed: %v", err)
	}
}

// initOTEL 初始化 OTEL 并设置日志级别
func initOTEL() *otel.OTELKit {
	kit, err := otel.InitOTEL(context.Background(),
		otel.WithServiceName("backservice"),
		otel.WithEndpoint("localhost:4317"),
		otel.WithInsecure(true),
	)
	if err != nil {
		// 注意：OTEL 初始化失败时，klog 还未初始化，使用标准库 log
		log.Fatalf("Failed to init OTEL: %v", err)
	}
	kit.Logger.SetLevel(klog.LevelDebug)
	return kit
}

// getServicePort 获取服务端口，支持环境变量配置
func getServicePort() int {
	port := consts.BackservicePortDefault
	if portStr := os.Getenv("BACKSERVICE_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}
	return port
}

// getLocalIP 获取本机IP
func getLocalIP() string {
	localIP, err := network.GetLocalIP()
	if err != nil {
		klog.Fatalf("Failed to get local IP: %v", err)
	}
	return localIP
}

// createBackServiceServer 创建 BackService 服务器
func createBackServiceServer(host string, port int, serviceCtx deps.ServiceContext) server.Server {
	kitexRegistry, err := kitexinfra.NewEtcdRegistryFromEnv()
	if err != nil {
		log.Fatalf("Failed to create kitex etcd registry: %v", err)
	}

	return backservice.NewServer(
		rpcadapter.NewBackHandler(serviceCtx),
		server.WithServiceAddr(&net.TCPAddr{IP: net.ParseIP(host), Port: port}),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: consts.BackserviceIMName}),
		server.WithRegistry(kitexRegistry),
		server.WithRegistryInfo(&kitexregistry.Info{
			Weight: 1,
			Tags: map[string]string{
				"version": "1.0.0",
			},
		}),
	)
}

// startServer 启动服务器
func startServer(svr server.Server) {
	go func() {
		if err := svr.Run(); err != nil {
			// 确保启动失败时整个进程直接退出，而不是仅仅打印日志
			log.Printf("Failed to start backservice server: %v", err)
			os.Exit(1)
		}
	}()

	// 等待 1 秒确保服务已经启动并监听端口
	time.Sleep(1 * time.Second)
}

// waitForShutdown 等待关闭信号
func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	klog.Info("Shutting down...")
}

func registerToBackbonServiceWithRetry(ctx context.Context, resolver kitexdiscovery.Resolver) {
	for attempt := 1; attempt <= backbonRegisterMaxAttempts; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, backbonRegisterTimeout)
		err := registerToBackbonService(attemptCtx, resolver)
		cancel()
		if err == nil {
			return
		}

		if attempt == backbonRegisterMaxAttempts {
			klog.Errorf("Failed to register to backbonservice after %d attempts: %v", backbonRegisterMaxAttempts, err)
			return
		}

		klog.Warnf("Failed to register to backbonservice (attempt %d/%d): %v", attempt, backbonRegisterMaxAttempts, err)

		select {
		case <-ctx.Done():
			klog.Warnf("Stopped registering to backbonservice: %v", ctx.Err())
			return
		case <-time.After(backbonRegisterInterval):
		}
	}
}

// registerToBackbonService 向 backbonservice 注册服务
func registerToBackbonService(ctx context.Context, resolver kitexdiscovery.Resolver) error {
	opts := []client.Option{
		client.WithSuite(tracing.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: consts.BackserviceIMName}),
		client.WithRPCTimeout(10 * time.Second),
	}
	if resolver != nil {
		opts = append(opts, client.WithResolver(resolver))
	}

	backbonClient, err := backbonservice.NewClient(
		consts.BackbonServiceName,
		opts...,
	)
	if err != nil {
		return fmt.Errorf("failed to create backbonservice client: %w", err)
	}
	// 注意：kitex 客户端可能没有 Close 方法，如果需要关闭可以检查是否有 Close 方法
	// defer func() {
	// 	if closer, ok := backbonClient.(interface{ Close() error }); ok {
	// 		_ = closer.Close()
	// 	}
	// }()

	// 注册服务：service = "backservice-im", methods = ["*"] (全部方法)
	registerReq := &backbon.RegisterServiceReq{
		Service: consts.BackserviceIMName,
		Methods: []string{"*"}, // 支持所有方法
	}

	registerResp, err := backbonClient.RegisterService(ctx, registerReq)
	if err != nil {
		return fmt.Errorf("failed to call RegisterService: %w", err)
	}

	if !registerResp.GetSuccess() {
		return fmt.Errorf("RegisterService failed: %s", registerResp.GetError())
	}

	klog.Infof("Successfully registered service %q with all methods to backbonservice", consts.BackserviceIMName)
	return nil
}
