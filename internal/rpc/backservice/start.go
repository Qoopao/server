package backservice

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/google/uuid"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	"github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen"
	backbonservice "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/backbonservice"
	"github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/backservice"
	"github.com/roc/roc-foundation-util-go/log/otel"
	"github.com/roc/roc-foundation-util-go/network"
	"github.com/roc/roc-foundation-util-go/service_registry/discovery"
	"github.com/roc/roc-foundation-util-go/service_registry/loadbalancer"
	foundationregistry "github.com/roc/roc-foundation-util-go/service_registry/registry"
	"github.com/roc/roc-foundation-util-go/service_registry/registry/etcd"
)

// Start 启动 BackService 服务器
func Start() {
	// 使用 foundation-util-go 统一初始化 OTEL（初始化 klog）
	kit, err := otel.InitOTEL(context.Background(),
		otel.WithServiceName("backservice"),
		otel.WithEndpoint("localhost:4317"),
		otel.WithInsecure(true),
	)
	if err != nil {
		// 注意：OTEL 初始化失败时，klog 还未初始化，使用标准库 log
		log.Fatalf("Failed to init OTEL: %v", err)
	}
	defer kit.Shutdown(context.Background())

	// 设置日志级别为 DEBUG
	kit.Logger.SetLevel(klog.LevelDebug)

	// 获取服务端口，支持环境变量配置
	port := 10300
	if portStr := os.Getenv("BACKSERVICE_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// 获取本机IP
	localIP, err := network.GetLocalIP()
	if err != nil {
		klog.Fatalf("Failed to get local IP: %v", err)
	}

	// 创建服务实例信息
	instanceID := fmt.Sprintf("backservice-%s", uuid.New().String()[:8])

	// 创建 Etcd Registry
	registry, err := etcd.NewEtcdRegistry(
		etcd.WithEndpoints([]string{"localhost:2379"}),
		etcd.WithDialTimeout(5*time.Second),
	)
	if err != nil {
		klog.Fatalf("Failed to create etcd registry: %v", err)
	}
	defer registry.Close()

	// 注册服务实例
	instance := &foundationregistry.ServiceInstance{
		ServiceName: "backservice-im",
		InstanceID:  instanceID,
		Host:        localIP,
		Port:        port,
		Weight:      1,
		Status:      foundationregistry.StatusHealthy,
		Metadata: map[string]string{
			"version": "1.0.0",
		},
	}

	// 向 backbonservice 注册服务（service = "backservice", method = "*"）
	// 在后台异步注册，避免阻塞启动（backbon-service 可能还未启动）
	// 注册逻辑会在 registerToBackbonService 内部执行（等待 1 秒后注册到 etcd）
	go func() {
		ctx := context.Background()
		if err := registerToBackbonService(ctx, registry, instance); err != nil {
			klog.Warnf("Failed to register to backbonservice: %v", err)
		}
	}()

	// 创建服务器
	svr := backservice.NewServer(
		NewBackServiceImpl(),
		server.WithServiceAddr(&net.TCPAddr{IP: net.ParseIP(localIP), Port: port}),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "backservice"}),
	)

	// 优雅关闭处理
	defer func() {
		if err := registry.Deregister(context.Background(), instance); err != nil {
			klog.Errorf("Failed to deregister service instance: %v", err)
		}
	}()

	klog.Infof("BackService starting - instance_id: %s, address: %s:%d", instanceID, localIP, port)

	err = svr.Run()

	if err != nil {
		klog.Errorf("Failed to run server: %v", err)
	}
}

// registerToBackbonService 向 backbonservice 注册服务
func registerToBackbonService(ctx context.Context, registry foundationregistry.Registry, instance *foundationregistry.ServiceInstance) error {
	// 等待 1 秒确保服务已经启动并监听端口
	time.Sleep(1 * time.Second)

	// 注册服务实例到 etcd
	if err := registry.Register(ctx, instance); err != nil {
		return fmt.Errorf("failed to register service instance to etcd: %w", err)
	}
	klog.Infof("Successfully registered backservice-im to etcd")

	// 创建服务发现客户端
	lb := loadbalancer.NewRoundRobinLoadBalancer()
	discoveryClient := discovery.NewDiscovery(registry, lb)

	// 获取 backbon-service 实例
	backbonInstance, err := discoveryClient.GetInstance(ctx, "backbon-service")
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
	registerReq := &kitex_gen.RegisterServiceReq{
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
