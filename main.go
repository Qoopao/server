package main

import (
	"context"

	"github.com/roc/roc-im-server/internal/kitex_gen/sdkws"
	"github.com/roc/roc-im-server/internal/msggateway"
	msgRpc "github.com/roc/roc-im-server/internal/rpc/msg"
	// "github.com/roc/roc-im-server/test/kafaka"
)

func test_mar() {
	var msg sdkws.MsgData
	msg.SendTime = 1000
	msg.ClientMsgID = "123"
	msg.ContentType = 1
	msg.SessionType = 1
	msg.RecvID = "123"
	msg.Seq = 1
	msg.SendID = "123"
	msg.Content = []byte("hello")

	// data := make([]byte, 0, 3)
	da, err := msg.Marshal(nil)
	if err != nil {

		// data = da
	}
	var msgTmp sdkws.MsgData
	msgTmp.Unmarshal(da)
	panic(err)
}

func main() {

	// test_mar()

	go func() {
		msgRpc.Start()
	}()

	// go func() {
	// 	kafaka_test.KfakTest()
	// }()

	// go func() {
	// 	kafaka_test.Consumer()
	// }()

	var wsServer = msggateway.NewWsServer(
		msggateway.WithPort(10010),
		msggateway.WithMaxConnNum(10000),
		msggateway.WithWriteBufferSize(1000),
		msggateway.WithHandshakeTimeout(1000),
		msggateway.WithMessageMaxMsgLength(10000),
	)

	wsServer.Run(context.Background())

	println("over")
}
