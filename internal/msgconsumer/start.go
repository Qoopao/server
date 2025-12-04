package msgconsumer

import (
	foundationcache "github.com/roc/roc-foundation-util-go/cache"
	"github.com/roc/roc-foundation-util-go/cache/redis"
	foundationmq "github.com/roc/roc-foundation-util-go/mq"
	"github.com/roc/roc-foundation-util-go/mq/kafka"
	"github.com/roc/roc-im-server/pkg/common/storage/controller"
)

func Start() {
	var (
		producer foundationmq.Producer
		consumer foundationmq.Consumer
		cache    foundationcache.Cache
		err      error
	)

	producer, err = kafka.NewKafkaProducer([]kafka.ProducerOption{
		kafka.WithProducerBrokers([]string{"localhost:9092"}),
	})
	if err != nil {
		panic("[error] producer create error: " + err.Error())
	}

	consumer, err = kafka.NewKafkaConsumer([]kafka.ConsumerOption{
		kafka.WithConsumerBrokers([]string{"localhost:9092"}),
		kafka.WithConsumerGroupID("msgconsumer-group"),
	})
	if err != nil {
		panic("[error] consumer create error: " + err.Error())
	}

	cache, err = redis.NewRedisCache([]redis.Option{
		redis.WithAddress("localhost:6379"),
		redis.WithPassword("redis123"),
		redis.WithDB(0),
	})
	if err != nil {
		panic("[error] cache create error: " + err.Error())
	}

	msgConsumer := ConsumerMessage{
		MessageDB: controller.NewCommonMsgDatabase(producer, consumer, cache),
		consumer:  consumer,
	}

	// 直接运行，阻塞在这里
	msgConsumer.Run()
}
