package main

import (
	"log"

	convMsgConsumer "github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer"
)

func main() {
	log.Println("启动 ConvMsgConsumer...")
	if err := convMsgConsumer.Start(); err != nil {
		log.Fatalf("ConvMsgConsumer 启动失败: %v", err)
	}
	select {}
}
