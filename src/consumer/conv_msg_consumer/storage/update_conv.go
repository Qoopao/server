package storage

import (
	"context"
	"errors"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
	"go.mongodb.org/mongo-driver/bson"
)

// UpdateConvWithMessage 更新/创建会话详情（根据消息数据保存完整的 ConversationData）
func (s *convMsgStorageImpl) UpdateConvWithMessage(ctx context.Context, msg *sdkws.MessageData) error {
	if msg == nil || msg.ConvID == "" {
		return nil
	}

	// 根据消息数据构造 ConversationData
	conv := &sdkws.ConversationData{
		ConvID:        msg.ConvID,
		ConvType:      msg.ConvType,
		LastMessageID: msg.SMessageID,
	}

	convID := msg.ConvID
	lastMessageSeq := msg.Seq
	store := s.getStore()
	filter := bson.M{"_id": convID}
	var existing orm.ConversationDocument
	err := store.FindOne(ctx, collectionConversations, filter, &existing)

	if err != nil {
		// 文档不存在，创建新文档
		if err == foundationstorage.ErrNotFound {
			return s.createConversation(ctx, conv, lastMessageSeq)
		}
		klog.CtxErrorf(ctx, "[ConvMsgStorage] find conversation failed",
			"conv_id", convID,
			"error", err.Error())
		return err
	}

	// 文档存在，更新
	return s.updateConversationFields(ctx, conv, lastMessageSeq)
}

// createConversation 创建新会话文档
func (s *convMsgStorageImpl) createConversation(ctx context.Context, conv *sdkws.ConversationData, lastMessageSeq int64) error {
	now := time.Now()
	store := s.getStore()

	// 序列化 ConversationData 为 protobuf 二进制
	data, err := conv.Marshal(nil)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConvMsgStorage] marshal conversation failed",
			"conv_id", conv.ConvID,
			"error", err.Error())
		return err
	}

	doc := &orm.ConversationDocument{
		ID:             conv.ConvID,
		Data:           data,
		LastMessageSeq: lastMessageSeq,
		UpdatedAt:      now,
	}
	_, err = store.InsertOne(ctx, collectionConversations, doc)
	if err != nil {
		// 如果插入时已存在（并发情况），则更新
		if errors.Is(err, foundationstorage.ErrDuplicateKey) {
			return s.updateConversationFields(ctx, conv, lastMessageSeq)
		}
		klog.CtxErrorf(ctx, "[ConvMsgStorage] create conversation failed",
			"conv_id", conv.ConvID,
			"error", err.Error())
		return err
	}

	klog.CtxDebugf(ctx, "[ConvMsgStorage] create conversation success",
		"conv_id", conv.ConvID,
		"last_message_seq", lastMessageSeq)
	return nil
}

// updateConversationFields 更新会话字段（保存完整的 ConversationData）
func (s *convMsgStorageImpl) updateConversationFields(ctx context.Context, conv *sdkws.ConversationData, lastMessageSeq int64) error {
	now := time.Now()
	store := s.getStore()

	// 序列化 ConversationData 为 protobuf 二进制
	data, err := conv.Marshal(nil)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConvMsgStorage] marshal conversation failed",
			"conv_id", conv.ConvID,
			"error", err.Error())
		return err
	}

	filter := bson.M{"_id": conv.ConvID}
	update := bson.M{
		"$set": bson.M{
			"data":             data,
			"last_message_seq": lastMessageSeq,
			"updated_at":       now,
		},
	}
	err = store.UpdateOne(ctx, collectionConversations, filter, update)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConvMsgStorage] update conversation failed",
			"conv_id", conv.ConvID,
			"error", err.Error())
		return err
	}

	klog.CtxDebugf(ctx, "[ConvMsgStorage] update conversation success",
		"conv_id", conv.ConvID,
		"last_message_seq", lastMessageSeq)
	return nil
}
