package conversation

import (
	"log"
	"net"

	"github.com/cloudwego/kitex/server"
	conversation "github.com/roc/roc-im-server/internal/kitex_gen/conversation/conversationservice"
	"github.com/roc/roc-im-server/pkg/common/storage/controller"
	"github.com/roc/roc-im-server/tools/kvstore"
	"github.com/roc/roc-im-server/tools/mq"
)

func Start() {
	var (
		err   error
		mqi   mq.MQ
		store kvstore.KVStore
	)

	mqi, err = mq.NewSaramaMQ([]string{"localhost:9092"})
	if err != nil {
		panic(err.Error())
	}

	store, err = kvstore.NewKVStore(kvstore.Config{
		Address:  "localhost:6379",
		Password: "Rhp.Roc.666",
		DB:       0,
	})
	if err != nil {
		panic(err.Error())
	}

	svr := conversation.NewServer(
		&ConversationServiceImpl{
			MessageDB: controller.NewCommonMsgDatabase(mqi, store),
		},
		server.WithServiceAddr(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 10200}),
	)

	err = svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
