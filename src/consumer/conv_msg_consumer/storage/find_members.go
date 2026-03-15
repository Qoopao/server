package storage

import (
	"context"
	"fmt"

	"github.com/rhp-QE/roc-foundation-util-go/storage"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
)

// GetConversationData 根据 convID 查询会话的 ConversationData（纯存储操作，不做业务解析）
func (s *convMsgStorageImpl) GetConversationData(ctx context.Context, convID string) (*sdkws.ConversationData, error) {
	if convID == "" {
		return nil, nil
	}

	store := s.getStore()
	if store == nil {
		return nil, fmt.Errorf("storage is nil when GetConversationData conv_id=%s", convID)
	}

	var doc orm.ConversationDocument
	if err := store.FindOne(ctx, orm.CollectionConversations, map[string]interface{}{"_id": convID}, &doc); err != nil {
		if err == storage.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("find conversation doc failed for conv_id=%s: %w", convID, err)
	}
	if len(doc.Data) == 0 {
		return nil, nil
	}

	var conv sdkws.ConversationData
	if err := conv.Unmarshal(doc.Data); err != nil {
		return nil, fmt.Errorf("unmarshal conversation data failed for conv_id=%s: %w", convID, err)
	}
	return &conv, nil
}
