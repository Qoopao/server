package main

import (
	"log"

	"github.com/rhp-QE/roc-im-server/internal/msgconsumer"
)

func main() {
	log.Println("启动 MsgConsumer...")

	// 启动消费者（会阻塞）
	msgconsumer.Start()

	// 阻塞，防止进程退出
	select {}
}
