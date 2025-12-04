package msg

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	// "log/slog"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/google/uuid"
	foundationcache "github.com/roc/roc-foundation-util-go/cache"
	"github.com/roc/roc-foundation-util-go/cache/redis"
	foundationmq "github.com/roc/roc-foundation-util-go/mq"
	"github.com/roc/roc-foundation-util-go/mq/kafka"
	foundationregistry "github.com/roc/roc-foundation-util-go/service_registry/registry"
	"github.com/roc/roc-foundation-util-go/service_registry/registry/etcd"
	msg "github.com/roc/roc-im-server/internal/kitex_gen/msg/messageservice"
	"github.com/roc/roc-im-server/pkg/common/storage/controller"
	"go.uber.org/zap"

	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	"github.com/roc/roc-foundation-util-go/log/otel"
	// "go.opentelemetry.io/contrib/bridges/otelslog"
	// "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	// "go.opentelemetry.io/otel/log/global"
	// sdklog "go.opentelemetry.io/otel/sdk/log"
)

// var otelLogger *slog.Logger

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

	// 创建 Kafka Producer
	producer, err = kafka.NewKafkaProducer([]kafka.ProducerOption{
		kafka.WithProducerBrokers([]string{"localhost:9092"}),
	})
	if err != nil {
		panic(err.Error())
	}

	// 创建 Kafka Consumer
	consumer, err = kafka.NewKafkaConsumer([]kafka.ConsumerOption{
		kafka.WithConsumerBrokers([]string{"localhost:9092"}),
		kafka.WithConsumerGroupID("msg-service-group"),
	})
	if err != nil {
		panic(err.Error())
	}

	// 创建 Redis Cache
	cache, err = redis.NewRedisCache([]redis.Option{
		redis.WithAddress("localhost:6379"),
		redis.WithPassword("redis123"),
		redis.WithDB(0),
	})
	if err != nil {
		panic(err.Error())
	}

	// 获取服务端口，支持环境变量配置
	port := 10100
	if portStr := os.Getenv("MSG_SERVICE_PORT"); portStr != "" {
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
	instanceID := fmt.Sprintf("msg-service-%s", uuid.New().String()[:8])

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
		ServiceName: "msg-service",
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

	// 使用 foundation-util-go 统一初始化 OTEL
	kit, err := otel.InitOTEL(context.Background(),
		otel.WithServiceName("msg-service"),
		otel.WithEndpoint("localhost:4317"),
		otel.WithInsecure(true),
	)
	if err != nil {
		log.Fatalf("Failed to init OTEL: %v", err)
	}
	defer kit.Shutdown(context.Background())

	// 创建服务器
	svr := msg.NewServer(
		&MessageServiceImpl{
			MsgDatabase: controller.NewCommonMsgDatabase(producer, consumer, cache),
		},
		server.WithServiceAddr(&net.TCPAddr{IP: net.ParseIP(localIP), Port: port}),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "msg-service"}),
	)

	// 设置日志级别为 DEBUG
	kit.Logger.SetLevel(klog.LevelDebug)

	// 优雅关闭处理
	defer func() {
		if err := registry.Deregister(context.Background(), instance); err != nil {
			Logger.Error("Failed to deregister service instance", zap.Error(err))
		}
	}()

	Logger.Info("Message service starting",
		zap.String("instance_id", instanceID),
		zap.String("address", fmt.Sprintf("%s:%d", localIP, port)))

	err = svr.Run()

	if err != nil {
		Logger.Error("Failed to run server", zap.Error(err))
		log.Println(err.Error())
	}
}

// func initLog(ctx context.Context) *sdklog.LoggerProvider {
// 	logExporter, err := otlploggrpc.New(ctx, otlploggrpc.WithInsecure())
// 	if err != nil {
// 		panic("failed to initialize exporter")
// 	}

// 	lp := sdklog.NewLoggerProvider(
// 		sdklog.WithProcessor(
// 			sdklog.NewSimpleProcessor(logExporter),
// 		),
// 	)

// 	global.SetLoggerProvider(lp)

// 	// init otelLogger, then use it directly anywhere in your app to record your log, and the
// 	// log content will be sent to apmplus backend.
// 	otelLogger = otelslog.NewLogger("slog")
// 	return lp
// }

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
