package main

import (
	"log"

	convmsgconsumer "github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer"
)

func main() {
	log.Println("启动 conv_msg_consumer...")

	if err := convmsgconsumer.Start(); err != nil {
		log.Fatalf("conv_msg_consumer exited with error: %v", err)
	}
}


