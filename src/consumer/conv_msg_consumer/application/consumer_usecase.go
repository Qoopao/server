package application

import (
	"context"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	deps "github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/infrastructure/deps"
	"github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/infrastructure/persistence"
)

// ConsumerUsecase 会话消息消费应用用例。
type ConsumerUsecase interface {
	// HandleMessage 处理消息业务逻辑：保存消息 → 更新会话 → 更新用户最近会话链 → 推送消息
	HandleMessage(ctx context.Context, msg *sdkws.MessageData) error
}

type consumerUsecase struct {
	repo       persistence.ConvMsgRepository
	serviceCtx deps.ServiceContext
}

// NewConsumerUsecase 创建 ConsumerUsecase 实例
func NewConsumerUsecase(repo persistence.ConvMsgRepository, serviceCtx deps.ServiceContext) ConsumerUsecase {
	return &consumerUsecase{
		repo:       repo,
		serviceCtx: serviceCtx,
	}
}

// HandleMessage 处理消息业务逻辑：保存消息 → 更新会话 → 更新用户最近会话链
// TODO: 消费失败 则需要投递到消息队列 重新消费
func (s *consumerUsecase) HandleMessage(ctx context.Context, msg *sdkws.MessageData) error {
	// 1. 保存消息详情（一条消息一条文档，写入单链）
	if err := s.saveMessage(ctx, msg); err != nil {
		return err
	}

	// 2. 修改/创建会话详情（保存完整的 ConversationData，更新 last_message_seq 和 updated_at）
	if err := s.updateConversation(ctx, msg); err != nil {
		return err
	}

	// 3. 如果是单聊则修改用户最近会话链
	if err := s.updateUserRecentConvsIfSmallConv(ctx, msg); err != nil {
		return err
	}

	// 4. 调用长链服务向用户推送消息
	s.pushMessage(ctx, msg)

	klog.CtxDebugf(ctx, "[ConvMsgConsumer] message processed successfully",
		"conv_id", msg.ConvID,
		"msg_id", msg.SMessageID,
		"seq", msg.Seq)
	return nil
}
