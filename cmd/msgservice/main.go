package main

import (
	"log"

	msgRpc "github.com/rhp-QE/roc-im-server/src/rpc/message_service"
)

func main() {
	log.Println("启动 Message Service...")
	if err := msgRpc.Start(); err != nil {
		log.Fatalf("Message Service 启动失败: %v", err)
	}
}
