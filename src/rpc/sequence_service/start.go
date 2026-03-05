package sequenceservice

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

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/google/uuid"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	"github.com/redis/go-redis/v9"
	"github.com/rhp-QE/roc-foundation-util-go/log/otel"
	"github.com/rhp-QE/roc-foundation-util-go/network"
	foundationregistry "github.com/rhp-QE/roc-foundation-util-go/service_registry/registry"
	"github.com/rhp-QE/roc-foundation-util-go/service_registry/registry/etcd"
	sequence "github.com/rhp-QE/roc-im-server/kitex_gen/sequence/sequenceservice"
	consts "github.com/rhp-QE/roc-im-server/src/const"
	"github.com/rhp-QE/roc-im-server/src/rpc/sequence_service/api"
	"github.com/rhp-QE/roc-im-server/src/rpc/sequence_service/service"
	servicecontext "github.com/rhp-QE/roc-im-server/src/rpc/sequence_service/service_context"
	"github.com/rhp-QE/roc-im-server/src/rpc/sequence_service/storage"
)

// Start 启动sequence服务
func Start() error {
	// 初始化 OTEL
	kit := initOTEL()
	defer kit.Shutdown(context.Background())

	// 创建 Redis 客户端
	redisClient := createRedisClient()

	// 创建服务上下文（托管 Redis 和注册中心）
	serviceCtx := createServiceContext(redisClient)
	defer serviceCtx.Close()

	// 获取本机地址
	host := getLocalIP()

	// 创建服务实例
	instance := createServiceInstance(host)

	// 创建服务器
	svr := createServer(serviceCtx, host)

	// 启动服务器
	startServer(svr)

	// 注册服务到 etcd
	registerService(serviceCtx, instance)

	// 设置优雅关闭处理
	defer setupGracefulShutdown(serviceCtx, instance)

	klog.Infof("Sequence service starting - instance_id: %s, address: %s:%d", instance.InstanceID, host, getServicePort())

	// 等待关闭信号
	waitForShutdown()

	return nil
}

// initOTEL 初始化 OTEL 并设置日志级别
func initOTEL() *otel.OTELKit {
	kit, err := otel.InitOTEL(context.Background(),
		otel.WithServiceName(consts.SequenceServiceName),
		otel.WithEndpoint("localhost:4317"),
		otel.WithInsecure(true),
	)
	if err != nil {
		log.Fatalf("Failed to init OTEL: %v", err)
	}
	kit.Logger.SetLevel(klog.LevelDebug)
	return kit
}

// createRedisClient 创建 Redis 客户端
func createRedisClient() *redis.Client {
	redisAddr := "localhost:6379"
	if addr := os.Getenv("REDIS_ADDRESS"); addr != "" {
		redisAddr = addr
	}

	redisPassword := ""
	if pwd := os.Getenv("REDIS_PASSWORD"); pwd != "" {
		redisPassword = pwd
	}

	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			redisDB = db
		}
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// 测试Redis连接
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	return redisClient
}

// createServiceContext 创建服务上下文
func createServiceContext(redisClient *redis.Client) servicecontext.ServiceContext {
	etcdEndpoints := []string{"localhost:2379"}
	if endpoints := os.Getenv("ETCD_ENDPOINTS"); endpoints != "" {
		etcdEndpoints = []string{endpoints}
	}

	registry, err := etcd.NewEtcdRegistry(
		etcd.WithEndpoints(etcdEndpoints),
		etcd.WithDialTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create etcd registry: %v", err)
	}

	return servicecontext.NewServiceContext(registry, redisClient)
}

// getLocalIP 获取本机IP
func getLocalIP() string {
	host, err := network.GetLocalIP()
	if err != nil {
		log.Fatalf("Failed to get local IP: %v", err)
	}
	return host
}

// getServicePort 获取服务端口
func getServicePort() int {
	port := consts.SequenceServicePortDefault
	if portStr := os.Getenv("SEQUENCE_SERVICE_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}
	return port
}

// createServiceInstance 创建服务实例信息
func createServiceInstance(host string) *foundationregistry.ServiceInstance {
	instanceID := fmt.Sprintf("%s-%s", consts.SequenceServiceName, uuid.New().String()[:8])
	return &foundationregistry.ServiceInstance{
		ServiceName: consts.SequenceServiceName,
		InstanceID:  instanceID,
		Host:        host,
		Port:        getServicePort(),
		Weight:      1,
		Status:      foundationregistry.StatusHealthy,
		Metadata: map[string]string{
			"version": "1.0.0",
		},
	}
}

// createServer 创建服务器
func createServer(serviceCtx servicecontext.ServiceContext, host string) server.Server {
	// 创建存储层和服务层
	seqStorage := storage.NewSequenceStorage(serviceCtx)
	seqService := service.NewSequenceService(seqStorage)
	// 创建接口层（API）
	seqHandler := api.NewSequenceAPI(seqService)

	return sequence.NewServer(
		seqHandler,
		server.WithServiceAddr(&net.TCPAddr{IP: net.ParseIP(host), Port: getServicePort()}),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: consts.SequenceServiceName}),
	)
}

// startServer 启动服务器
func startServer(svr server.Server) {
	go func() {
		if err := svr.Run(); err != nil {
			// 这里必须确保整个进程退出，而不是仅仅打日志
			log.Printf("Failed to start sequence server: %v", err)
			os.Exit(1)
		}
	}()

	// 等待 1 秒确保服务已经启动并监听端口
	time.Sleep(1 * time.Second)
}

// registerService 注册服务到 etcd
func registerService(serviceCtx servicecontext.ServiceContext, instance *foundationregistry.ServiceInstance) {
	// 服务启动后再注册到 etcd，避免保活机制在服务未就绪时删除注册
	if err := serviceCtx.GetRegistry().Register(context.Background(), instance); err != nil {
		klog.Fatalf("Failed to register service instance: %v", err)
	}
	klog.Infof("Successfully registered sequence-service to etcd")
}

// setupGracefulShutdown 设置优雅关闭处理
func setupGracefulShutdown(serviceCtx servicecontext.ServiceContext, instance *foundationregistry.ServiceInstance) {
	if err := serviceCtx.GetRegistry().Deregister(context.Background(), instance); err != nil {
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
