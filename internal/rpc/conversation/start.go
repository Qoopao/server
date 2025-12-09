package conversation

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/cloudwego/kitex/server"
	"github.com/google/uuid"
	foundationcache "github.com/rhp-QE/roc-foundation-util-go/cache"
	"github.com/rhp-QE/roc-foundation-util-go/cache/redis"
	foundationmq "github.com/rhp-QE/roc-foundation-util-go/mq"
	"github.com/rhp-QE/roc-foundation-util-go/mq/kafka"
	foundationregistry "github.com/rhp-QE/roc-foundation-util-go/service_registry/registry"
	"github.com/rhp-QE/roc-foundation-util-go/service_registry/registry/etcd"
	conversation "github.com/rhp-QE/roc-im-server/internal/kitex_gen/conversation/conversationservice"
	"github.com/rhp-QE/roc-im-server/pkg/common/storage/controller"
	"go.uber.org/zap"
)

// Logger 全局日志实例
var Logger *zap.Logger

// InitLogger 初始化日志
func InitLogger() error {
	var err error
	Logger, err = zap.NewProduction()
	if err != nil {
		return err
	}
	return nil
}

// Sync 同步日志
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

func Start() {
	// 初始化日志
	if err := InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer Sync()

	var (
		err      error
		producer foundationmq.Producer
		consumer foundationmq.Consumer
		cache    foundationcache.Cache
	)

	producer, err = kafka.NewKafkaProducer([]kafka.ProducerOption{
		kafka.WithProducerBrokers([]string{"localhost:9092"}),
	})
	if err != nil {
		panic(err.Error())
	}

	consumer, err = kafka.NewKafkaConsumer([]kafka.ConsumerOption{
		kafka.WithConsumerBrokers([]string{"localhost:9092"}),
		kafka.WithConsumerGroupID("conversation-service-group"),
	})
	if err != nil {
		panic(err.Error())
	}

	cache, err = redis.NewRedisCache([]redis.Option{
		redis.WithAddress("localhost:6379"),
		redis.WithPassword("redis123"),
		redis.WithDB(0),
	})
	if err != nil {
		panic(err.Error())
	}

	// 获取服务端口，支持环境变量配置
	port := 10200
	if portStr := os.Getenv("CONVERSATION_SERVICE_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// 获取本机IP
	localIP, err := getLocalIP()
	if err != nil {
		log.Fatalf("Failed to get local IP: %v", err)
	}

	// 创建服务实例信息
	instanceID := fmt.Sprintf("conversation-service-%s", uuid.New().String()[:8])

	// 创建 Etcd Registry
	registry, err := etcd.NewEtcdRegistry(
		etcd.WithEndpoints([]string{"localhost:2379"}),
		etcd.WithDialTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create etcd registry: %v", err)
	}
	defer registry.Close()

	// 注册服务实例
	instance := &foundationregistry.ServiceInstance{
		ServiceName: "conversation-service",
		InstanceID:  instanceID,
		Host:        localIP,
		Port:        port,
		Weight:      1,
		Metadata: map[string]string{
			"version": "1.0.0",
		},
	}

	if err := registry.Register(context.Background(), instance); err != nil {
		log.Fatalf("Failed to register service instance: %v", err)
	}

	// 创建服务器
	svr := conversation.NewServer(
		&ConversationServiceImpl{
			MessageDB: controller.NewCommonMsgDatabase(producer, consumer, cache),
		},
		server.WithServiceAddr(&net.TCPAddr{IP: net.ParseIP(localIP), Port: port}),
	)

	// 优雅关闭处理
	defer func() {
		if err := registry.Deregister(context.Background(), instance); err != nil {
			Logger.Error("Failed to deregister service instance", zap.Error(err))
		}
	}()

	Logger.Info("Conversation service starting",
		zap.String("instance_id", instanceID),
		zap.String("address", fmt.Sprintf("%s:%d", localIP, port)))

	err = svr.Run()

	if err != nil {
		Logger.Error("Failed to run server", zap.Error(err))
		log.Println(err.Error())
	}
}

// getLocalIP 获取本机IP地址
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no non-loopback IP found")
}

// createServiceRegistry 创建服务注册中心
