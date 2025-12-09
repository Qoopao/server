package main

import (
	"log"

	convRpc "github.com/rhp-QE/roc-im-server/internal/rpc/conversation"
)

func main() {
	log.Println("启动 Conversation Service...")

	// 启动 RPC 服务
	convRpc.Start()
}

