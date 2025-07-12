package pushhandler

import (
	"context"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/transport"
	"github.com/roc/roc-im-server/internal/kitex_gen/push"
	"github.com/roc/roc-im-server/internal/kitex_gen/push/pushservice"
	"github.com/roc/roc-im-server/internal/kitex_gen/sdkws"
	"github.com/roc/roc-im-server/pkg/user"
	"github.com/roc/roc-im-server/tools/mq"
)

func handle() {
	mqi, err := mq.NewSaramaMQ([]string{"localhost:9092"})
	if err != nil {
		panic(err)
	}

	ctx, _ := context.WithCancel(context.Background())

	err = mqi.Subscribe("message_topic", func(ctx context.Context, msg *mq.Message) error {
		return pushHandler(ctx, msg)
	},
		mq.WithAutoAck(true),
		mq.WithGroupID("push_handler_group"),
	)

	if err != nil {
		println("failed to subscribe  %w", err)
	}

	// 等待上下文取消
	<-ctx.Done()

	return
}

func pushHandler(ctx context.Context, msg *mq.Message) error {
	println("Processing message:", string(msg.Body))

	msgData := sdkws.MsgData{}
	if err := msgData.Unmarshal(msg.Body); err != nil {
		println("Failed to unmarshal message:", err.Error())
		return err
	}

	userService := user.NewUserService()
	address, _ := userService.UserAddress(ctx, msgData.RecvID)

	pushServiceClient, _ := pushservice.NewClient("push_service", client.WithHostPorts(address), client.WithTransportProtocol(transport.GRPC))
	pushServiceClient.PushMsg(ctx, &push.PushMsgReq{
		UserIDs:        []string{msgData.RecvID},
		ConversationID: "conversation_id_mock",
		MsgData:        &msgData,
	})

	return nil
}
