package storage

import (
	"context"
	"os"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-foundation-util-go/mq"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	servicecontext "github.com/rhp-QE/roc-im-server/src/rpc/message_service/service_context"
)

// MessageStorage 定义消息持久化/投递接口（存储层）
// 在当前阶段，仅负责将消息投递到消息队列
type MessageStorage interface {
	// PublishMessages 将消息批量投递到 MQ
	PublishMessages(ctx context.Context, msgs []*sdkws.MessageData) error
}

type messageStorageImpl struct {
	serviceCtx servicecontext.ServiceContext
	topic      string
}

// NewMessageStorage 创建 MessageStorage 实现
func NewMessageStorage(serviceCtx servicecontext.ServiceContext) MessageStorage {
	topic := os.Getenv("MESSAGE_SERVICE_TOPIC")
	if topic == "" {
		topic = "im_message_topic"
	}

	return &messageStorageImpl{
		serviceCtx: serviceCtx,
		topic:      topic,
	}
}

// PublishMessages 将消息批量投递到 MQ
func (s *messageStorageImpl) PublishMessages(ctx context.Context, msgs []*sdkws.MessageData) error {
	if len(msgs) == 0 {
		return nil
	}

	producer := s.serviceCtx.GetMQProducer()
	if producer == nil {
		klog.CtxErrorf(ctx, "message_storage: mq producer is nil")
		return nil
	}

	mqMsgs := make([]*mq.Message, 0, len(msgs))
	for _, msg := range msgs {
		if msg == nil {
			continue
		}

		body, err := msg.Marshal(nil)
		if err != nil {
			klog.CtxErrorf(ctx, "message_storage: marshal message failed",
				"conv_id", msg.ConvID,
				"server_msg_id", msg.SMessageID,
				"client_msg_id", msg.CMessaegID,
				"error", err.Error())
			continue
		}

		mqMsgs = append(mqMsgs, &mq.Message{
			Topic:     s.topic,
			Key:       msg.ConvID,
			Value:     body,
			Timestamp: time.Now(),
		})
	}

	if len(mqMsgs) == 0 {
		return nil
	}

	klog.CtxDebugf(ctx, "message_storage: publish messages to mq",
		"topic", s.topic,
		"count", len(mqMsgs))

	_, err := producer.SendBatch(ctx, mqMsgs)
	return err
}
