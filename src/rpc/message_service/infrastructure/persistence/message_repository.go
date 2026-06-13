package persistence

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-foundation-util-go/mq"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	deps "github.com/rhp-QE/roc-im-server/src/rpc/message_service/infrastructure/deps"
)

// MessageRepository 定义消息事实仓储接口。
// 这里封装 Mongo message fact 和 MQ 投递 outbox 入口，application 层不直接依赖底层存储 SDK。
type MessageRepository interface {
	// SaveMessageFact 写入消息事实。
	// 发送成功语义依赖该写入先完成；重复写入同一 server_msg_id 时返回已存在消息，保持幂等。
	SaveMessageFact(ctx context.Context, msg *sdkws.MessageData) (*sdkws.MessageData, bool, error)

	// FindMessageByClientMsgID 通过 sender + client_msg_id 查询已提交消息，用于客户端超时重试去重。
	FindMessageByClientMsgID(ctx context.Context, sendID string, clientMsgID string) (*sdkws.MessageData, bool, error)

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

type messageRepository struct {
	serviceCtx deps.ServiceContext
	topic      string
}

// NewMessageRepository 创建 MessageRepository 实现
func NewMessageRepository(serviceCtx deps.ServiceContext) MessageRepository {
	topic := os.Getenv("MESSAGE_SERVICE_TOPIC")
	if topic == "" {
		topic = "im_message_topic"
	}

	return &messageRepository{
		serviceCtx: serviceCtx,
		topic:      topic,
	}
}

// PublishMessages 将消息批量投递到 MQ
func (s *messageRepository) PublishMessages(ctx context.Context, msgs []*sdkws.MessageData) error {
	if len(msgs) == 0 {
		return nil
	}

	producer := s.serviceCtx.GetMQProducer()
	if producer == nil {
		klog.CtxErrorf(ctx, "message_repository: mq producer is nil")
		return errors.New("mq producer is nil")
	}

	mqMsgs := make([]*mq.Message, 0, len(msgs))
	for _, msg := range msgs {
		if msg == nil {
			continue
		}

		body, err := msg.Marshal(nil)
		if err != nil {
			klog.CtxErrorf(ctx, "message_repository: marshal message failed",
				"conv_id", msg.ConvID,
				"server_msg_id", msg.SMessageID,
				"client_msg_id", msg.CMessaegID,
				"error", err.Error())
			return err
		}

		mqMsgs = append(mqMsgs, &mq.Message{
			Topic:     s.topic,
			Key:       msg.ConvID,
			Value:     body,
			Timestamp: time.Now(),
		})
	}

	if len(mqMsgs) == 0 {
		return errors.New("no valid message to publish")
	}

	klog.CtxDebugf(ctx, "message_repository: publish messages to mq",
		"topic", s.topic,
		"count", len(mqMsgs))

	_, err := producer.SendBatch(ctx, mqMsgs)
	return err
}
