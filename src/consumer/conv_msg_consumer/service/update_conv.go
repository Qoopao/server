package service

import (
	"context"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
)

// updateConversation 修改/创建会话详情（保存完整的 ConversationData，更新 last_message_seq 和 updated_at）
func (s *convMsgConsumerServiceImpl) updateConversation(ctx context.Context, msg *sdkws.MessageData) error {
	if err := s.storage.UpdateConvWithMessage(ctx, msg); err != nil {
		klog.CtxErrorf(ctx, "[ConvMsgConsumer] update conversation failed",
			"conv_id", msg.ConvID,
			"seq", msg.Seq,
			"error", err.Error())
		return err
	}
	return nil
}

