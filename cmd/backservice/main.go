package main

import (
	"log"

	backserviceRpc "github.com/roc/roc-im-server/internal/rpc/backservice"
)

func main() {
	log.Println("启动 BackService...")

	// 启动 RPC 服务
	backserviceRpc.Start()
}
