package service

import (
	"context"

	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
)

// saveMessage 保存消息详情（一条消息一条文档，写入单链）
func (s *convMsgConsumerServiceImpl) saveMessage(ctx context.Context, msg *sdkws.MessageData) error {
	return s.storage.SaveMessage(ctx, msg.SMessageID, msg)
}

