package conversation

import (
	"log"
	"net"

	"github.com/cloudwego/kitex/server"
	conversation "github.com/roc/roc-im-server/internal/kitex_gen/conversation/conversationservice"
)

func Start() {
	svr := conversation.NewServer(
		new(ConversationServiceImpl),
		server.WithServiceAddr(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 10200}),
	)

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
