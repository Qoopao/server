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
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/google/uuid"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	backservice "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/back/backservice"
	backbon "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/backbon"
	backbonservice "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/backbon/backbonservice"
	"github.com/rhp-QE/roc-foundation-util-go/log/otel"
	"github.com/rhp-QE/roc-foundation-util-go/network"
	foundationregistry "github.com/rhp-QE/roc-foundation-util-go/service_registry/registry"
	"github.com/rhp-QE/roc-foundation-util-go/service_registry/registry/etcd"
	"github.com/rhp-QE/roc-im-server/internal/rpc/backservice/servicecontext"
)

// Start 启动 BackService 服务器
func Start() {
	// 初始化 OTEL
	kit := initOTEL()
	defer kit.Shutdown(context.Background())

	// 获取配置信息
	port := getServicePort()
	localIP := getLocalIP()

	// 创建 Etcd Registry
	registry := createEtcdRegistry()
	defer registry.Close()

	// 创建服务上下文
	serviceCtx := servicecontext.NewServiceContext(registry)
	defer serviceCtx.Close()

	// 创建服务实例
	instance := createServiceInstance(localIP, port)

	// 创建服务器
	svr := createBackServiceServer(localIP, port, serviceCtx)

	// 启动服务器
	startServer(svr)

	// 注册服务到 etcd
	registerService(registry, instance)

	// 设置优雅关闭处理
	defer setupGracefulShutdown(registry, instance)

	klog.Infof("BackService starting - instance_id: %s, address: %s:%d", instance.InstanceID, localIP, port)

	// 等待关闭信号
	waitForShutdown()
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
	port := 10300
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

// createEtcdRegistry 创建 Etcd Registry
func createEtcdRegistry() foundationregistry.Registry {
	registry, err := etcd.NewEtcdRegistry(
		etcd.WithEndpoints([]string{"localhost:2379"}),
		etcd.WithDialTimeout(5*time.Second),
	)
	if err != nil {
		klog.Fatalf("Failed to create etcd registry: %v", err)
	}
	return registry
}

// createServiceInstance 创建服务实例信息
func createServiceInstance(host string, port int) *foundationregistry.ServiceInstance {
	instanceID := fmt.Sprintf("backservice-%s", uuid.New().String()[:8])
	return &foundationregistry.ServiceInstance{
		ServiceName: "backservice-im",
		InstanceID:  instanceID,
		Host:        host,
		Port:        port,
		Weight:      1,
		Status:      foundationregistry.StatusHealthy,
		Metadata: map[string]string{
			"version": "1.0.0",
		},
	}
}

// createBackServiceServer 创建 BackService 服务器
func createBackServiceServer(host string, port int, serviceCtx servicecontext.ServiceContext) server.Server {
	return backservice.NewServer(
		NewBackServiceImpl(serviceCtx),
		server.WithServiceAddr(&net.TCPAddr{IP: net.ParseIP(host), Port: port}),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "backservice"}),
	)
}

// startServer 启动服务器
func startServer(svr server.Server) {
	go func() {
		if err := svr.Run(); err != nil {
			klog.Fatalf("Failed to start backservice server: %v", err)
		}
	}()

	// 等待 1 秒确保服务已经启动并监听端口
	time.Sleep(1 * time.Second)
}

// registerService 注册服务到 etcd
func registerService(registry foundationregistry.Registry, instance *foundationregistry.ServiceInstance) {
	// 服务启动后再注册到 etcd，避免保活机制在服务未就绪时删除注册
	if err := registry.Register(context.Background(), instance); err != nil {
		klog.Fatalf("Failed to register service instance: %v", err)
	}
	klog.Infof("Successfully registered backservice-im to etcd")

	// 异步注册到 backbonservice
	go func() {
		ctx := context.Background()
		if err := registerToBackbonService(ctx, registry, instance); err != nil {
			klog.Warnf("Failed to register to backbonservice: %v", err)
		}
	}()
}

// setupGracefulShutdown 设置优雅关闭处理
func setupGracefulShutdown(registry foundationregistry.Registry, instance *foundationregistry.ServiceInstance) {
	if err := registry.Deregister(context.Background(), instance); err != nil {
		klog.Errorf("Failed to deregister service instance: %v", err)
	}
}

// waitForShutdown 等待关闭信号
func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	klog.Info("Shutting down...")
}

// registerToBackbonService 向 backbonservice 注册服务
func registerToBackbonService(ctx context.Context, registry foundationregistry.Registry, instance *foundationregistry.ServiceInstance) error {
	// 创建服务上下文（用于服务发现）
	serviceCtx := servicecontext.NewServiceContext(registry)

	// 获取 backbon-service 实例
	backbonInstance, err := serviceCtx.GetDiscovery().GetInstance(ctx, "backbon-service")
	if err != nil {
		return fmt.Errorf("no available instance (backbon-service may not be started yet): %w", err)
	}

	// 创建 backbonservice 客户端
	hostPort := fmt.Sprintf("%s:%d", backbonInstance.Host, backbonInstance.Port)
	backbonClient, err := backbonservice.NewClient(
		"backbon-service",
		client.WithHostPorts(hostPort),
		client.WithSuite(tracing.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "backbon-service"}),
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

	// 注册服务：service = "backservice", methods = ["*"] (全部方法)
	registerReq := &backbon.RegisterServiceReq{
		Service: "backservice-im",
		Methods: []string{"*"}, // 支持所有方法
	}

	registerResp, err := backbonClient.RegisterService(ctx, registerReq)
	if err != nil {
		return fmt.Errorf("failed to call RegisterService: %w", err)
	}

	if !registerResp.GetSuccess() {
		return fmt.Errorf("RegisterService failed: %s", registerResp.GetError())
	}

	klog.Infof("Successfully registered service 'backservice' with all methods to backbonservice")
	return nil
}
