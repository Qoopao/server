package storage

import (
	"context"
	"errors"

	"github.com/cloudwego/kitex/pkg/klog"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	// collectionMessages 消息集合名称（实际集合名会加上前缀 "im_db_"）
	collectionMessages = "messages"
)

// FetchConvMessageListWithRange 查询会话消息列表（仅做查询和数据转换）
func (s *messageStorageImpl) FetchConvMessageListWithRange(ctx context.Context, convID string, left int64, right int64) ([]*sdkws.MessageData, error) {
	// 构建查询条件
	filter := bson.M{"conv_id": convID}

	// 如果 left 和 right 都为 0，表示查询所有消息
	// 否则使用范围查询：left <= seq <= right
	if left > 0 || right > 0 {
		seqFilter := bson.M{}
		if left > 0 {
			seqFilter["$gte"] = left
		}
		if right > 0 {
			seqFilter["$lte"] = right
		}
		if len(seqFilter) > 0 {
			filter["seq"] = seqFilter
		}
	}

	// 统一使用升序排序
	sort := bson.D{{Key: "seq", Value: 1}}

	// 查询文档
	var docs []orm.MessageDocument
	store := s.serviceCtx.GetStorage()
	if store == nil {
		return nil, errors.New("storage is nil")
	}

	err := store.Find(ctx, collectionMessages, filter, &docs,
		foundationstorage.WithSort(sort),
	)
	if err != nil {
		klog.CtxErrorf(ctx, "[MessageStorage] fetch conv message list failed",
			"conv_id", convID,
			"left", left,
			"right", right,
			"error", err.Error())
		return nil, err
	}

	// 转换为 MessageData
	messages := make([]*sdkws.MessageData, 0, len(docs))
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
		messages = append(messages, msg)
	}

	klog.CtxDebugf(ctx, "[MessageStorage] fetch conv message list success",
		"conv_id", convID,
		"left", left,
		"right", right,
		"count", len(messages))

	return messages, nil
}

// FetchConvLatestMessageList 查询会话内最新 n 条消息（仅做查询和数据转换）
func (s *messageStorageImpl) FetchConvLatestMessageList(ctx context.Context, convID string, limit int64) ([]*sdkws.MessageData, error) {
	// 构建查询条件
	filter := bson.M{"conv_id": convID}

	// 按 seq 降序排序，获取最新的消息
	sort := bson.D{{Key: "seq", Value: -1}}

	// 查询文档
	var docs []orm.MessageDocument
	store := s.serviceCtx.GetStorage()
	if store == nil {
		return nil, errors.New("storage is nil")
	}

	err := store.Find(ctx, collectionMessages, filter, &docs,
		foundationstorage.WithLimit(limit),
		foundationstorage.WithSort(sort),
	)
	if err != nil {
		klog.CtxErrorf(ctx, "[MessageStorage] fetch conv latest messages failed",
			"conv_id", convID,
			"limit", limit,
			"error", err.Error())
		return nil, err
	}

	// 转换为 MessageData
	messages := make([]*sdkws.MessageData, 0, len(docs))
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
		messages = append(messages, msg)
	}

	// 反转顺序，使结果按 seq 升序排列（最旧的消息在前，最新的消息在后）
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	klog.CtxDebugf(ctx, "[MessageStorage] fetch conv latest messages success",
		"conv_id", convID,
		"limit", limit,
		"count", len(messages))

	return messages, nil
}
