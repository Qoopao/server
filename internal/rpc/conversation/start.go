package conversation

import (
	conversation "github.com/roc/roc-im-server/internal/kitex_gen/conversation/conversationservice"
	"log"
)

func Start() {
	svr := conversation.NewServer(new(ConversationServiceImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
