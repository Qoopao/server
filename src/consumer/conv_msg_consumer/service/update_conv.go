package service

import (
	"context"

	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
)

// updateConversation 修改/创建会话详情（保存完整的 ConversationData，更新 last_message_seq 和 updated_at）
func (s *convMsgConsumerServiceImpl) updateConversation(ctx context.Context, msg *sdkws.MessageData) error {
	return s.storage.UpdateConvWithMessage(ctx, msg)
}
