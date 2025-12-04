package main

import (
	"log"

	msgRpc "github.com/roc/roc-im-server/internal/rpc/msg"
)

func main() {
	log.Println("启动 Msg Service...")

	// 启动 RPC 服务
	msgRpc.Start()
}

