package main

import (
	"log"

	convRpc "github.com/rhp-QE/roc-im-server/src/rpc/conversation_service"
)

func main() {
	log.Println("启动 Conversation Service...")
	if err := convRpc.Start(); err != nil {
		log.Fatalf("Conversation Service 启动失败: %v", err)
	}
}
