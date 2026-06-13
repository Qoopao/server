package sequenceservice

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	kitexregistry "github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	"github.com/redis/go-redis/v9"
	"github.com/rhp-QE/roc-foundation-util-go/log/otel"
	"github.com/rhp-QE/roc-foundation-util-go/network"
	sequence "github.com/rhp-QE/roc-im-server/kitex_gen/sequence/sequenceservice"
	"github.com/rhp-QE/roc-im-server/src/common/kitexinfra"
	consts "github.com/rhp-QE/roc-im-server/src/const"
	"github.com/rhp-QE/roc-im-server/src/rpc/sequence_service/application"
	deps "github.com/rhp-QE/roc-im-server/src/rpc/sequence_service/infrastructure/deps"
	"github.com/rhp-QE/roc-im-server/src/rpc/sequence_service/infrastructure/persistence"
	rpcadapter "github.com/rhp-QE/roc-im-server/src/rpc/sequence_service/interfaces/rpc"
)

// Start 启动sequence服务
func Start() error {
	// 初始化 OTEL
	kit := initOTEL()
	defer kit.Shutdown(context.Background())

	// 创建 Redis 客户端
	redisClient := createRedisClient()

	// 创建服务上下文（只托管 Redis，服务注册由 Kitex server 托管）
	serviceCtx := createServiceContext(redisClient)
	defer serviceCtx.Close()

	// 获取本机地址
	host := getLocalIP()

	// 创建服务器
	svr := createServer(serviceCtx, host)

	// 启动服务器
	startServer(svr)

	klog.Infof("Sequence service started, address: %s:%d", host, getServicePort())

	// 等待关闭信号
	waitForShutdown()

	if err := svr.Stop(); err != nil {
		klog.Errorf("Sequence service stop failed: %v", err)
	}

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
func createServiceContext(redisClient *redis.Client) deps.ServiceContext {
	return deps.NewServiceContext(redisClient)
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

// createServer 创建服务器
func createServer(serviceCtx deps.ServiceContext, host string) server.Server {
	kitexRegistry, err := kitexinfra.NewEtcdRegistryFromEnv()
	if err != nil {
		log.Fatalf("Failed to create kitex etcd registry: %v", err)
	}

	// 创建存储层和服务层
	seqRepo := persistence.NewSequenceRepository(serviceCtx)
	seqUsecase := application.NewSequenceUsecase(seqRepo)
	seqHandler := rpcadapter.NewSequenceHandler(seqUsecase)

	return sequence.NewServer(
		seqHandler,
		server.WithServiceAddr(&net.TCPAddr{IP: net.ParseIP(host), Port: getServicePort()}),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: consts.SequenceServiceName}),
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
			// 这里必须确保整个进程退出，而不是仅仅打日志
			log.Printf("Failed to start sequence server: %v", err)
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
