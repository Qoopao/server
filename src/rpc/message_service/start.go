package messageservice

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
	"github.com/rhp-QE/roc-foundation-util-go/log/otel"
	"github.com/rhp-QE/roc-foundation-util-go/mq"
	"github.com/rhp-QE/roc-foundation-util-go/mq/kafka"
	"github.com/rhp-QE/roc-foundation-util-go/network"
	foundationregistry "github.com/rhp-QE/roc-foundation-util-go/service_registry/registry"
	"github.com/rhp-QE/roc-foundation-util-go/service_registry/registry/etcd"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	mongodb "github.com/rhp-QE/roc-foundation-util-go/storage/mongodb"
	message "github.com/rhp-QE/roc-im-server/kitex_gen/message/messageservice"
	consts "github.com/rhp-QE/roc-im-server/src/const"
	"github.com/rhp-QE/roc-im-server/src/rpc/message_service/api"
	"github.com/rhp-QE/roc-im-server/src/rpc/message_service/service"
	servicecontext "github.com/rhp-QE/roc-im-server/src/rpc/message_service/service_context"
	msgstorage "github.com/rhp-QE/roc-im-server/src/rpc/message_service/storage"
)

// Start 启动 message_service
func Start() error {
	// 初始化 OTEL
	kit := initOTEL()
	defer kit.Shutdown(context.Background())

	// 创建 MQ Producer
	producer := createMQProducer()
	defer producer.Close()

	// 创建 MongoDB 存储
	store := createMongoStorage()
	defer store.Close()

	// 创建 Etcd Registry
	registry := createRegistry()

	// 创建 ServiceContext
	serviceCtx := servicecontext.NewServiceContext(registry, producer, store)
	defer serviceCtx.Close()

	// 获取本机地址
	host := getLocalIP()

	// 创建服务实例信息
	instance := createServiceInstance(host)

	// 创建服务器
	svr := createServer(serviceCtx, host)

	// 启动服务器
	startServer(svr)

	// 注册服务到 etcd（服务启动后再注册）
	registerService(serviceCtx, instance)

	// 设置优雅关闭处理
	defer setupGracefulShutdown(serviceCtx, instance)

	klog.Infof("Message service starting - instance_id: %s, address: %s:%d", instance.InstanceID, host, getServicePort())

	// 等待关闭信号
	waitForShutdown()

	return nil
}

// initOTEL 初始化 OTEL
func initOTEL() *otel.OTELKit {
	kit, err := otel.InitOTEL(context.Background(),
		otel.WithServiceName(consts.MessageServiceName),
		otel.WithEndpoint("localhost:4317"),
		otel.WithInsecure(true),
	)
	if err != nil {
		log.Fatalf("Failed to init OTEL: %v", err)
	}
	kit.Logger.SetLevel(klog.LevelDebug)
	return kit
}

// createMQProducer 创建 Kafka Producer
func createMQProducer() mq.Producer {
	brokers := []string{"localhost:9092"}
	if addr := os.Getenv("KAFKA_BROKERS"); addr != "" {
		brokers = []string{addr}
	}

	producer, err := kafka.NewKafkaProducer([]kafka.ProducerOption{
		kafka.WithProducerBrokers(brokers),
	})
	if err != nil {
		log.Fatalf("Failed to create kafka producer: %v", err)
	}
	return producer
}

// createMongoStorage 创建 MongoDB 存储实例
func createMongoStorage() foundationstorage.Storage {
	uri := os.Getenv("MESSAGE_MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	database := os.Getenv("MESSAGE_MONGODB_DATABASE")
	if database == "" {
		database = "im_message"
	}

	store, err := mongodb.NewMongoStorage(
		[]mongodb.Option{
			mongodb.WithURI(uri),
			mongodb.WithDatabase(database),
		},
	)
	if err != nil {
		log.Fatalf("Failed to create mongo storage: %v", err)
	}
	return store
}

// createRegistry 创建 Etcd Registry
func createRegistry() foundationregistry.Registry {
	endpoints := []string{"localhost:2379"}
	if addr := os.Getenv("ETCD_ENDPOINTS"); addr != "" {
		endpoints = []string{addr}
	}

	registry, err := etcd.NewEtcdRegistry(
		etcd.WithEndpoints(endpoints),
		etcd.WithDialTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create etcd registry: %v", err)
	}
	return registry
}

// getLocalIP 获取本机 IP
func getLocalIP() string {
	host, err := network.GetLocalIP()
	if err != nil {
		log.Fatalf("Failed to get local IP: %v", err)
	}
	return host
}

// getServicePort 获取服务端口
func getServicePort() int {
	port := consts.MessageServicePortDefault
	if portStr := os.Getenv("MESSAGE_SERVICE_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}
	return port
}

// createServiceInstance 创建服务实例信息
func createServiceInstance(host string) *foundationregistry.ServiceInstance {
	instanceID := fmt.Sprintf("%s-%s", consts.MessageServiceName, uuid.New().String()[:8])
	return &foundationregistry.ServiceInstance{
		ServiceName: consts.MessageServiceName,
		InstanceID:  instanceID,
		Host:        host,
		Port:        getServicePort(),
		Status:      foundationregistry.StatusHealthy,
		Weight:      1,
		Metadata: map[string]string{
			"version": "1.0.0",
		},
	}
}

// createServer 创建 kitex 服务器
func createServer(serviceCtx servicecontext.ServiceContext, host string) server.Server {
	// 创建存储层
	msgStorage := msgstorage.NewMessageStorage(serviceCtx)
	// 创建逻辑层
	msgService := service.NewMessageService(msgStorage, serviceCtx)
	// 创建接口层（API）
	msgHandler := api.NewMessageAPI(msgService)

	return message.NewServer(
		msgHandler,
		server.WithServiceAddr(&net.TCPAddr{IP: net.ParseIP(host), Port: getServicePort()}),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: consts.MessageServiceName}),
	)
}

// startServer 在 goroutine 中启动服务器，并等待 1 秒确保就绪
func startServer(svr server.Server) {
	go func() {
		if err := svr.Run(); err != nil {
			// 确保启动失败时整个进程直接退出，而不是仅仅打印日志
			log.Printf("Failed to start message server: %v", err)
			os.Exit(1)
		}
	}()

	time.Sleep(1 * time.Second)
}

// registerService 将服务注册到 etcd
func registerService(serviceCtx servicecontext.ServiceContext, instance *foundationregistry.ServiceInstance) {
	if err := serviceCtx.GetRegistry().Register(context.Background(), instance); err != nil {
		klog.Fatalf("Failed to register message-service instance: %v", err)
	}
	klog.Infof("Successfully registered message-service to etcd")
}

// setupGracefulShutdown 反注册服务实例
func setupGracefulShutdown(serviceCtx servicecontext.ServiceContext, instance *foundationregistry.ServiceInstance) {
	if err := serviceCtx.GetRegistry().Deregister(context.Background(), instance); err != nil {
		klog.Errorf("Failed to deregister message-service instance: %v", err)
	}
}

// waitForShutdown 等待关闭信号
func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	klog.Info("Message service shutting down...")
}
