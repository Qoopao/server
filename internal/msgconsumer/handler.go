package msgconsumer

import (
	"context"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/transport"
	"github.com/openimsdk/tools/log"
	"github.com/roc/roc-im-server/internal/kitex_gen/push"
	"github.com/roc/roc-im-server/internal/kitex_gen/push/pushservice"
	"github.com/roc/roc-im-server/internal/kitex_gen/sdkws"
	"github.com/roc/roc-im-server/pkg/common/storage/controller"
	"github.com/roc/roc-im-server/pkg/user"
	"github.com/roc/roc-im-server/tools/mq"
	"github.com/roc/roc-im-server/tools/utils"
)

type ConsumerMessage struct {
	MessageDB controller.CommonMsgDatabase
	mqi       mq.MQ
}

func (m *ConsumerMessage) Run() {

	var err error

	m.mqi, err = mq.NewSaramaMQ([]string{"localhost:9092"})
	if err != nil {
		panic(err)
	}

	ctx, _ := context.WithCancel(context.Background())

	err = m.mqi.Subscribe("message_topic", func(ctx context.Context, msg *mq.Message) error {
		return m.pushHandler(ctx, msg)
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

func (m *ConsumerMessage) pushHandler(ctx context.Context, msg *mq.Message) error {
	var (
		err     error
		userIDs []string
		message *sdkws.MsgData
	)

	log.ZDebug(ctx, "consume messsage_topic ", "msg", string(msg.Body))

	messageID := string(msg.Body)
	if message, err = m.MessageDB.GetMsgInfo(ctx, messageID); err != nil {
		return err
	}

	userService := user.NewUserService()

	// 获取消息接收者
	if userIDs, err = userService.GetUserIDsFromConv(ctx, message.ConvID); err != nil {
		return err
	}

	// 对userIDs进行去重
	userIDs = utils.RemoveDuplicate(userIDs)

	// 剔除发送者
	userIDs = utils.RemoveElement(userIDs, message.SendID)

	for _, userID := range userIDs {
		address, _ := userService.UserAddress(ctx, userID)
		// 修改用户混链
		m.MessageDB.UpdateUserConvList(ctx, userID, message.ConvID)

		// 推送消息
		pushServiceClient, _ := pushservice.NewClient("push_service", client.WithHostPorts(address), client.WithTransportProtocol(transport.GRPC))
		pushServiceClient.PushMsg(ctx, &push.PushMsgReq{
			UserIDs:        []string{userID},
			ConversationID: message.ConvID,
			MsgData:        message,
		})
	}

	return nil
}
