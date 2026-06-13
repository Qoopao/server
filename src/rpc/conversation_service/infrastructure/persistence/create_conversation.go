package persistence

import (
	"context"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
)

// CreateConversation 将会话数据写入库（仅做 ConversationData -> ConversationDocument 转换与 InsertOne，不区分单聊/群聊）
func (s *conversationRepository) CreateConversation(ctx context.Context, conv *sdkws.ConversationData) error {
	if conv == nil || conv.ConvID == "" {
		return nil
	}

	store := s.serviceCtx.GetStorage()
	if store == nil {
		return nil
	}

	data, err := conv.Marshal(nil)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationRepository] marshal conversation failed", "conv_id", conv.ConvID, "error", err.Error())
		return err
	}

	doc := &orm.ConversationDocument{
		ID:             conv.ConvID,
		Data:           data,
		LastMessageSeq: 0,
		UpdatedAt:      time.Now(),
	}

	if _, err = store.InsertOne(ctx, orm.CollectionConversations, doc); err != nil {
		klog.CtxErrorf(ctx, "[ConversationRepository] insert conversation failed", "conv_id", conv.ConvID, "error", err.Error())
		return err
	}

	klog.CtxDebugf(ctx, "[ConversationRepository] create conversation success", "conv_id", conv.ConvID)
	return nil
}
