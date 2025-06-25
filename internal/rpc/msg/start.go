package msg

import (
	"log"
	"net"

	"github.com/cloudwego/kitex/server"
	msg "github.com/roc/roc-im-server/internal/kitex_gen/msg/messageservice"
	"github.com/roc/roc-im-server/pkg/common/storage/controller"
	"github.com/roc/roc-im-server/tools/mq" // 替换为实际的包路径
)

func Start() {
	mqi, err := mq.NewSaramaMQ([]string{"localhost:9092"})
	svr := msg.NewServer(
		&MessageServiceImpl{
			MsgDatabase: controller.NewCommonMsgDatabase(mqi),
		},
		server.WithServiceAddr(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 10100}),
	)

	err = svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
