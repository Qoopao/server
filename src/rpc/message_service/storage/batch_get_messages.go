package storage

import (
	"context"
	"errors"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-foundation-util-go/stringutil"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
	"go.mongodb.org/mongo-driver/bson"
)

// BatchGetMessages 根据消息ID列表批量获取消息详情（仅做查询和数据转换）
func (s *messageStorageImpl) BatchGetMessages(ctx context.Context, messageIDs []string, convID string) (map[string]*sdkws.MessageData, error) {
	if len(messageIDs) == 0 {
		return make(map[string]*sdkws.MessageData), nil
	}

	// 构建查询条件
	filter := bson.M{
		"_id": bson.M{"$in": messageIDs},
	}
	// 如果提供了 convID，添加过滤条件
	if stringutil.IsNotEmpty(convID) {
		filter["conv_id"] = convID
	}

	// 查询文档
	var docs []orm.MessageDocument
	store := s.serviceCtx.GetStorage()
	if store == nil {
		return nil, errors.New("storage is nil")
	}

	err := store.Find(ctx, orm.CollectionMessages, filter, &docs)
	if err != nil {
		klog.CtxErrorf(ctx, "[MessageStorage] batch get messages failed",
			"message_ids", messageIDs,
			"conv_id", convID,
			"error", err.Error())
		return nil, err
	}

	// 转换为 MessageData 并构建映射
	result := make(map[string]*sdkws.MessageData, len(docs))
	for _, doc := range docs {
		msg := &sdkws.MessageData{}
		if err := msg.Unmarshal(doc.Data); err != nil {
			klog.CtxErrorf(ctx, "[MessageStorage] unmarshal message data failed",
				"msg_id", doc.ID,
				"conv_id", doc.ConvID,
				"seq", doc.Seq,
				"error", err.Error())
			continue
		}
		result[doc.ID] = msg
	}

	klog.CtxDebugf(ctx, "[MessageStorage] batch get messages success",
		"request_count", len(messageIDs),
		"found_count", len(result),
		"conv_id", convID)

	return result, nil
}
