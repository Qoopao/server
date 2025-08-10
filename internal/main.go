package main

import (
	conversation "github.com/roc/roc-im-server/internal/kitex_gen/conversation/conversationservice"
	"log"
)

func main() {
	svr := conversation.NewServer(new(ConversationServiceImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
