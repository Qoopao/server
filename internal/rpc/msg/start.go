package msg

import (
	"log"
	"net"

	"github.com/cloudwego/kitex/server"
	msg "github.com/roc/roc-im-server/internal/kitex_gen/msg/messageservice"
	"github.com/roc/roc-im-server/pkg/common/storage/controller"
	"github.com/roc/roc-im-server/tools/kvstore"
	"github.com/roc/roc-im-server/tools/mq" // 替换为实际的包路径
	"go.uber.org/zap"
)

func Start() {
	// 初始化日志
	if err := InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer Sync()

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

	svr := msg.NewServer(
		&MessageServiceImpl{
			MsgDatabase: controller.NewCommonMsgDatabase(mqi, store),
		},
		server.WithServiceAddr(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 10100}),
	)

	err = svr.Run()

	if err != nil {
		Logger.Error("Failed to run server", zap.Error(err))
		log.Println(err.Error())
	}
}
