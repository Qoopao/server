package main

import (
	"context"
	"errors"
	"net"

	"github.com/roc/roc-im-server/internal/kitex_gen/sdkws"
	"github.com/roc/roc-im-server/internal/msggateway"
	"github.com/roc/roc-im-server/internal/pushhandler"
	msgRpc "github.com/roc/roc-im-server/internal/rpc/msg"
	"github.com/roc/roc-im-server/test/redis"
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

func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil { // 优先IPv4
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", errors.New("no non-loopback IP found")
}

func main() {

	// test_mar()
	redis.Redis_Test()

	res, _ := GetLocalIP()
	println("Local IP:", res)

	// 开启 push_handler
	go func() {
		pushhandler.Start()
	}()

	// 开启 message_rpc
	go func() {
		msgRpc.Start()
	}()

	// go func() {
	// 	kafaka_test.KfakTest()
	// }()

	// go func() {
	// 	kafaka_test.Consumer()
	// }()

	// 开启 msggateway
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
