package persistence

import (
	"context"
	"errors"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-foundation-util-go/stringutil"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
	"go.mongodb.org/mongo-driver/bson"
)

// BatchGetConversations 根据会话ID列表批量获取会话详情
func (s *conversationRepository) BatchGetConversations(ctx context.Context, convIDs []string, ownerID string) (map[string]*sdkws.ConversationData, error) {
	if len(convIDs) == 0 {
		return make(map[string]*sdkws.ConversationData), nil
	}

	store := s.serviceCtx.GetStorage()
	if store == nil {
		return nil, errors.New("storage is nil")
	}

	// 构建查询条件
	filter := bson.M{
		"_id": bson.M{"$in": convIDs},
	}

	// 如果提供了 ownerID，添加过滤条件（注意：ConversationDocument 中没有 ownerID 字段，这里可能需要根据实际数据结构调整）
	// 暂时不添加 ownerID 过滤，因为 ConversationDocument 结构中没有此字段

	// 查询文档
	var docs []orm.ConversationDocument
	err := store.Find(ctx, orm.CollectionConversations, filter, &docs)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationRepository] batch get conversations failed",
			"conv_ids", convIDs,
			"owner_id", ownerID,
			"error", err.Error())
		return nil, err
	}

	// 转换为 ConversationData 并构建映射
	result := make(map[string]*sdkws.ConversationData, len(docs))
	for _, doc := range docs {
		convData := &sdkws.ConversationData{}
		if err := convData.Unmarshal(doc.Data); err != nil {
			klog.CtxErrorf(ctx, "[ConversationRepository] unmarshal conversation data failed",
				"conv_id", doc.ID,
				"error", err.Error())
			continue
		}

		// 如果提供了 ownerID，进行过滤
		if stringutil.IsNotEmpty(ownerID) && convData.OwnerID != ownerID {
			continue
		}

		result[doc.ID] = convData
	}

	klog.CtxDebugf(ctx, "[ConversationRepository] batch get conversations success",
		"request_count", len(convIDs),
		"found_count", len(result),
		"owner_id", ownerID)

	return result, nil
}
