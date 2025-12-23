package convmsgconsumer

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	foundationmq "github.com/rhp-QE/roc-foundation-util-go/mq"
	"github.com/rhp-QE/roc-foundation-util-go/mq/kafka"
	foundationregistry "github.com/rhp-QE/roc-foundation-util-go/service_registry/registry"
	"github.com/rhp-QE/roc-foundation-util-go/service_registry/registry/etcd"
	"github.com/rhp-QE/roc-foundation-util-go/storage"
	mongodb "github.com/rhp-QE/roc-foundation-util-go/storage/mongodb"
	convapi "github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/api"
	"github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/service"
	servicecontext "github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/service_context"
	convstorage "github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/storage"
	"go.mongodb.org/mongo-driver/bson"
)

// Start 启动 conv_msg_consumer：从消息 MQ 消费消息，写入 MongoDB
func Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建 MongoDB 存储
	store := createMongoStorage()
	// 创建 MQ Consumer
	consumer := createMQConsumer()
	// 创建 Etcd Registry
	registry := createEtcdRegistry()
	defer registry.Close()

	// 创建 ServiceContext
	svcCtx := servicecontext.NewServiceContext(store, consumer, registry)
	defer svcCtx.Close()

	// 组装各层：storage → service → api
	convStorage := convstorage.NewConvMsgStorage(svcCtx)
	convService := service.NewConvMsgConsumerService(convStorage, svcCtx)
	convAPI := convapi.NewConvMsgConsumerAPI(convService)

	// 启动消费循环（api 层负责订阅消息队列）
	go func() {
		if err := convAPI.Subscribe(ctx, svcCtx.GetConsumer()); err != nil {
			klog.CtxErrorf(ctx, "[ConvMsgConsumer] subscribe failed", "error", err.Error())
			cancel()
		}
	}()

	klog.Info("[ConvMsgConsumer] started, waiting for messages...")

	// 优雅退出：监听系统信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	select {
	case <-ctx.Done():
	case <-sigCh:
		klog.Info("[ConvMsgConsumer] received interrupt signal, shutting down...")
		cancel()
	}

	// 等待资源关闭
	time.Sleep(1 * time.Second)
	return nil
}

// createMongoStorage 创建 MongoDB 存储实例
func createMongoStorage() storage.Storage {
	uri := os.Getenv("CONV_MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	database := os.Getenv("CONV_MONGODB_DATABASE")
	if database == "" {
		database = "im_message"
	}

	store, err := mongodb.NewMongoStorage(
		[]mongodb.Option{
			mongodb.WithURI(uri),
			mongodb.WithDatabase(database),
		},
		storage.WithCollectionPrefix("im_db_"),
	)
	if err != nil {
		log.Fatalf("[ConvMsgConsumer] failed to create mongo storage: %v", err)
	}

	// 为集合创建必要索引
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 实际集合名为 prefix + "messages"（即 conv_messages）
	// 索引1：按 _id（即 msg_id）查询消息（_id 本身已有主键索引，这里显式创建便于查询优化）
	if err := store.CreateIndex(ctx, "messages", bson.D{
		{Key: "_id", Value: 1},
	}, false); err != nil {
		log.Printf("[ConvMsgConsumer] create index messages _id failed: %v", err)
	}
	// 索引2：按 conv_id+seq 查询单链
	if err := store.CreateIndex(ctx, "messages", bson.D{
		{Key: "conv_id", Value: 1},
		{Key: "seq", Value: 1},
	}, false); err != nil {
		log.Printf("[ConvMsgConsumer] create index conv_id+seq failed: %v", err)
	}

	// 实际集合名为 prefix + "conversations"（即 conv_conversations）
	// 索引：按 _id（即 conv_id）查询会话（_id 本身已有主键索引，这里显式创建便于查询优化）
	if err := store.CreateIndex(ctx, "conversations", bson.D{
		{Key: "_id", Value: 1},
	}, false); err != nil {
		log.Printf("[ConvMsgConsumer] create index conversations _id failed: %v", err)
	}

	// 实际集合名为 prefix + "user_recent_conversations"（即 conv_user_recent_conversations）
	// 索引：按 user_id+version 查询和排序（客户端增量同步）
	if err := store.CreateIndex(ctx, "user_recent_conversations", bson.D{
		{Key: "user_id", Value: 1},
		{Key: "version", Value: -1}, // 降序，最新的在前
	}, false); err != nil {
		log.Printf("[ConvMsgConsumer] create index user_id+version failed: %v", err)
	}

	return store
}

// createMQConsumer 创建 Kafka Consumer
func createMQConsumer() foundationmq.Consumer {
	brokers := []string{"localhost:9092"}
	if v := os.Getenv("KAFKA_BROKERS"); v != "" {
		brokers = []string{v}
	}

	topic := os.Getenv("MESSAGE_SERVICE_TOPIC")
	if topic == "" {
		topic = "im_message_topic"
	}

	groupID := os.Getenv("CONV_MSG_CONSUMER_GROUP")
	if groupID == "" {
		groupID = "conv-msg-consumer-group"
	}

	consumer, err := kafka.NewKafkaConsumer([]kafka.ConsumerOption{
		kafka.WithConsumerBrokers(brokers),
		kafka.WithConsumerGroupID(groupID),
		kafka.WithConsumerTopics([]string{topic}),
	})
	if err != nil {
		log.Fatalf("[ConvMsgConsumer] failed to create kafka consumer: %v", err)
	}
	return consumer
}

// createEtcdRegistry 创建 Etcd Registry
func createEtcdRegistry() foundationregistry.Registry {
	endpoints := []string{"localhost:2379"}
	if v := os.Getenv("ETCD_ENDPOINTS"); v != "" {
		endpoints = []string{v}
	}

	registry, err := etcd.NewEtcdRegistry(
		etcd.WithEndpoints(endpoints),
		etcd.WithDialTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("[ConvMsgConsumer] failed to create etcd registry: %v", err)
	}
	return registry
}
