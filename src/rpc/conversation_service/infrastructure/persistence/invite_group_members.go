package persistence

import (
	"context"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
	"go.mongodb.org/mongo-driver/bson"
)

// UpdateConversation 更新会话文档（仅做 ConversationData -> 文档字段 + UpdateOne）
func (s *conversationRepository) UpdateConversation(ctx context.Context, conv *sdkws.ConversationData) error {
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

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"data":       data,
			"updated_at": now,
		},
	}
	if err := store.UpdateOne(ctx, orm.CollectionConversations, bson.M{"_id": conv.ConvID}, update); err != nil {
		klog.CtxErrorf(ctx, "[ConversationRepository] update conversation failed", "conv_id", conv.ConvID, "error", err.Error())
		return err
	}

	klog.CtxDebugf(ctx, "[ConversationRepository] update conversation success", "conv_id", conv.ConvID)
	return nil
}
