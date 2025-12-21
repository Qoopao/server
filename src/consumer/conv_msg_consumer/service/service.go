package service

import (
	"context"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	servicecontext "github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/service_context"
	convstorage "github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/storage"
)

// ConvMsgConsumerService 会话消息消费者业务接口
type ConvMsgConsumerService interface {
	// HandleMessage 处理消息业务逻辑：保存消息 → 更新会话 → 更新用户最近会话链 → 推送消息
	HandleMessage(ctx context.Context, msg *sdkws.MessageData) error
}

type convMsgConsumerServiceImpl struct {
	storage    convstorage.ConvMsgStorage
	serviceCtx servicecontext.ServiceContext
}

// NewConvMsgConsumerService 创建 ConvMsgConsumerService 实例
func NewConvMsgConsumerService(storage convstorage.ConvMsgStorage, serviceCtx servicecontext.ServiceContext) ConvMsgConsumerService {
	return &convMsgConsumerServiceImpl{
		storage:    storage,
		serviceCtx: serviceCtx,
	}
}

// HandleMessage 处理消息业务逻辑：保存消息 → 更新会话 → 更新用户最近会话链
// TODO: 消费失败 则需要投递到消息队列 重新消费
func (s *convMsgConsumerServiceImpl) HandleMessage(ctx context.Context, msg *sdkws.MessageData) error {
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
