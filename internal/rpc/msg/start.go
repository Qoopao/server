package msg

import (
	msg "github.com/roc/roc-im-server/internal/kitex_gen/msg/messageservice"
	"log"
)

func Start() {
	svr := msg.NewServer(new(MessageServiceImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
