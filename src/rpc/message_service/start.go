package messageservice

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
	"github.com/rhp-QE/roc-foundation-util-go/log/otel"
	"github.com/rhp-QE/roc-foundation-util-go/mq"
	"github.com/rhp-QE/roc-foundation-util-go/mq/kafka"
	"github.com/rhp-QE/roc-foundation-util-go/network"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	mongodb "github.com/rhp-QE/roc-foundation-util-go/storage/mongodb"
	message "github.com/rhp-QE/roc-im-server/kitex_gen/message/messageservice"
	"github.com/rhp-QE/roc-im-server/src/common/kitexinfra"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
	consts "github.com/rhp-QE/roc-im-server/src/const"
	"github.com/rhp-QE/roc-im-server/src/rpc/message_service/application"
	deps "github.com/rhp-QE/roc-im-server/src/rpc/message_service/infrastructure/deps"
	persistence "github.com/rhp-QE/roc-im-server/src/rpc/message_service/infrastructure/persistence"
	rpcadapter "github.com/rhp-QE/roc-im-server/src/rpc/message_service/interfaces/rpc"
	"go.mongodb.org/mongo-driver/bson"
)

// Start 启动 message_service
func Start() error {
	// 初始化 OTEL
	kit := initOTEL()
	defer kit.Shutdown(context.Background())

	// 创建 MQ Producer
	producer := createMQProducer()

	// 创建 MongoDB 存储
	store := createMongoStorage()

	// Kitex 通过 resolver 完成服务发现，业务层不再持有注册发现抽象。
	sequenceResolver, err := kitexinfra.NewEtcdResolverFromEnv()
	if err != nil {
		log.Fatalf("Failed to create kitex etcd resolver: %v", err)
	}

	// 创建 ServiceContext
	serviceCtx := deps.NewServiceContext(producer, store, sequenceResolver)
	defer serviceCtx.Close()

	// 获取本机地址
	host := getLocalIP()

	// 创建服务器
	svr := createServer(serviceCtx, host)

	// 启动服务器
	startServer(svr)

	klog.Infof("Message service started, address: %s:%d", host, getServicePort())

	// 等待关闭信号
	waitForShutdown()

	if err := svr.Stop(); err != nil {
		klog.Errorf("Message service stop failed: %v", err)
	}

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
		database = orm.DBMessage
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
	ensureMessageIndexes(store)
	return store
}

func ensureMessageIndexes(store foundationstorage.Storage) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := store.CreateIndex(ctx, orm.CollectionMessages, bson.D{
		{Key: "conv_id", Value: 1},
		{Key: "seq", Value: 1},
	}, false); err != nil {
		log.Printf("Message service create index conv_id+seq failed: %v", err)
	}

	if err := store.CreateIndex(ctx, orm.CollectionMessages, bson.D{
		{Key: "send_id", Value: 1},
		{Key: "client_msg_id", Value: 1},
	}, false); err != nil {
		log.Printf("Message service create index send_id+client_msg_id failed: %v", err)
	}
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

// createServer 创建 kitex 服务器
func createServer(serviceCtx deps.ServiceContext, host string) server.Server {
	kitexRegistry, err := kitexinfra.NewEtcdRegistryFromEnv()
	if err != nil {
		log.Fatalf("Failed to create kitex etcd registry: %v", err)
	}

	// 创建基础设施、应用、RPC 适配层。
	msgRepository := persistence.NewMessageRepository(serviceCtx)
	msgService := application.NewMessageService(msgRepository, serviceCtx)
	msgHandler := rpcadapter.NewMessageHandler(msgService)

	return message.NewServer(
		msgHandler,
		server.WithServiceAddr(&net.TCPAddr{IP: net.ParseIP(host), Port: getServicePort()}),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: consts.MessageServiceName}),
		server.WithRegistry(kitexRegistry),
		server.WithRegistryInfo(&kitexregistry.Info{
			Weight: 1,
			Tags: map[string]string{
				"version": "1.0.0",
			},
		}),
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

// waitForShutdown 等待关闭信号
func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	klog.Info("Message service shutting down...")
}
