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
type MessageStorage interface {
	// PublishMessages 将消息批量投递到 MQ
	PublishMessages(ctx context.Context, msgs []*sdkws.MessageData) error

	// FetchConvMessageListWithRange 查询会话消息列表
	// convID: 会话ID
	// left: 查询范围的左边界（seq），包含
	// right: 查询范围的右边界（seq），包含
	// 返回：消息列表（按 seq 升序）
	FetchConvMessageListWithRange(ctx context.Context, convID string, left int64, right int64) ([]*sdkws.MessageData, error)

	// FetchConvLatestMessageList 查询会话内最新 n 条消息
	// convID: 会话ID
	// limit: 返回消息数量限制
	// 返回：消息列表（按 seq 升序）
	FetchConvLatestMessageList(ctx context.Context, convID string, limit int64) ([]*sdkws.MessageData, error)

	// BatchGetMessages 根据消息ID列表批量获取消息详情
	// messageIDs: 消息ID列表
	// convID: 会话ID（可选，用于过滤）
	// 返回：消息ID到消息数据的映射
	BatchGetMessages(ctx context.Context, messageIDs []string, convID string) (map[string]*sdkws.MessageData, error)
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
