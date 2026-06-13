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
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	mongodb "github.com/rhp-QE/roc-foundation-util-go/storage/mongodb"
	"github.com/rhp-QE/roc-im-server/src/common/kitexinfra"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
	"github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/application"
	deps "github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/infrastructure/deps"
	"github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/infrastructure/persistence"
	mqadapter "github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/interfaces/mq"
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
	// Kitex resolver 负责下游 sequence/backbon 服务发现。
	resolver, err := kitexinfra.NewEtcdResolverFromEnv()
	if err != nil {
		log.Fatalf("[ConvMsgConsumer] failed to create kitex etcd resolver: %v", err)
	}

	// 创建 ServiceContext
	svcCtx := deps.NewServiceContext(store, consumer, resolver)
	defer svcCtx.Close()

	// 组装各层：persistence → application → MQ handler。
	convRepo := persistence.NewConvMsgRepository(svcCtx)
	consumerUsecase := application.NewConsumerUsecase(convRepo, svcCtx)
	consumerHandler := mqadapter.NewConsumerHandler(consumerUsecase)

	// 启动消费循环（MQ handler 负责订阅消息队列）
	go func() {
		if err := consumerHandler.Subscribe(ctx, svcCtx.GetConsumer()); err != nil {
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
func createMongoStorage() foundationstorage.Storage {
	uri := os.Getenv("CONV_MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	database := os.Getenv("CONV_MONGODB_DATABASE")
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
		log.Fatalf("[ConvMsgConsumer] failed to create mongo repo: %v", err)
	}

	// 为集合创建必要索引
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 实际集合名为 prefix + "messages"（即 conv_messages）
	// 注意：_id 字段已有 MongoDB 自动创建的唯一索引，无需手动创建
	// 索引：按 conv_id+seq 查询单链
	if err := store.CreateIndex(ctx, "messages", bson.D{
		{Key: "conv_id", Value: 1},
		{Key: "seq", Value: 1},
	}, false); err != nil {
		log.Printf("[ConvMsgConsumer] create index conv_id+seq failed: %v", err)
	}

	// 实际集合名为 prefix + "conversations"（即 conv_conversations）
	// 注意：_id 字段已有 MongoDB 自动创建的唯一索引，无需手动创建

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
