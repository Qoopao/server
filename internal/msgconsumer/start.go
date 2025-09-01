package msgconsumer

import (
	"github.com/roc/roc-im-server/pkg/common/storage/controller"
	"github.com/roc/roc-im-server/tools/kvstore"
	"github.com/roc/roc-im-server/tools/mq"
)

func Start() {
	go func() {
		var (
			mqi   mq.MQ
			err   error
			store kvstore.KVStore
		)

		mqi, err = mq.NewSaramaMQ([]string{"localhost:9092"})
		if err != nil {
			panic("[error] mq create error" + err.Error())
		}

		store, err = kvstore.NewKVStore(kvstore.Config{
			Address:  "localhost:6379",
			Password: "redis123",
			DB:       0,
		})

		consumer := ConsumerMessage{
			MessageDB: controller.NewCommonMsgDatabase(mqi, store),
			mqi:       mqi,
		}

		consumer.Run()
	}()
}
