package main

import (
	"log"

	backserviceRpc "github.com/rhp-QE/roc-im-server/src/rpc/backservice"
)

func main() {
	log.Println("启动 BackService...")

	// 启动 RPC 服务
	backserviceRpc.Start()
}
