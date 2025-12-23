package storage

import (
	"context"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
	"go.mongodb.org/mongo-driver/bson"
)

// TODO 更新会话逻辑有问题， double check 数据竞争和一致性

// UpdateConvWithMessage 更新/创建会话详情（只更新 last_message_seq，不更新 data 字段）
// 并发安全：使用 LastMessageSeq 条件更新，防止旧消息覆盖新消息（先 Update 再 Insert）
// 注意：data 字段的更新由后续新增的 RPC 接口完成
func (s *convMsgStorageImpl) UpdateConvWithMessage(ctx context.Context, msg *sdkws.MessageData) error {
	if msg == nil || msg.ConvID == "" {
		return nil
	}

	convID := msg.ConvID
	lastMessageSeq := msg.Seq

	// 1. 先尝试更新已有文档（正常路径：绝大部分都是已经存在的会话）
	if updated, err := s.tryUpdateConv(ctx, convID, lastMessageSeq); err != nil {
		return err
	} else if updated {
		return nil
	}

	// 2. 更新失败（文档不存在或 seq 已更新），尝试插入新文档
	// 根据消息数据构造 ConversationData（首次插入需要完整的 data）
	conv := &sdkws.ConversationData{
		ConvID:        msg.ConvID,
		ConvType:      msg.ConvType,
		LastMessageID: msg.SMessageID,
	}
	return s.tryInsertConv(ctx, conv, lastMessageSeq)
}

// tryUpdateConv 尝试更新已有文档（只更新 last_message_seq，不更新 data 字段）
// 只有当旧 LastMessageSeq < 新 LastMessageSeq 时才允许被覆盖
func (s *convMsgStorageImpl) tryUpdateConv(ctx context.Context, convID string, lastMessageSeq int64) (bool, error) {
	now := time.Now()
	store := s.getStore()

	filter := bson.M{
		"_id":              convID,
		"last_message_seq": bson.M{"$lt": lastMessageSeq}, // 只有旧 seq 才允许被更新
	}
	update := bson.M{
		"$set": bson.M{
			"last_message_seq": lastMessageSeq,
			"updated_at":       now,
		},
	}

	if err := store.UpdateOne(ctx, collectionConversations, filter, update); err == nil {
		// 更新成功（命中了旧 seq）
		klog.CtxDebugf(ctx, "[ConvMsgStorage] update existing conversation doc success",
			"conv_id", convID,
			"last_message_seq", lastMessageSeq)
		return true, nil
	} else if err != foundationstorage.ErrNotFound {
		// 真实错误，直接返回
		klog.CtxErrorf(ctx, "[ConvMsgStorage] update conversation doc failed",
			"conv_id", convID,
			"error", err.Error())
		return false, err
	}

	// ErrNotFound：文档不存在或 seq 已 >= 本次 seq
	return false, nil
}

// tryInsertConv 尝试插入新文档（首次插入需要完整的 data 字段）
// 如果遇到重复键则尝试条件更新（只更新 last_message_seq，不更新 data）
func (s *convMsgStorageImpl) tryInsertConv(ctx context.Context, conv *sdkws.ConversationData, lastMessageSeq int64) error {
	now := time.Now()
	store := s.getStore()

	// 序列化 ConversationData 为 protobuf 二进制（首次插入需要完整的 data）
	data, err := conv.Marshal(nil)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConvMsgStorage] marshal conversation failed",
			"conv_id", conv.ConvID,
			"error", err.Error())
		return err
	}

	doc := &orm.ConversationDocument{
		ID:             conv.ConvID,
		Data:           data, // 首次插入需要完整的 data
		LastMessageSeq: lastMessageSeq,
		UpdatedAt:      now,
	}

	if _, err := store.InsertOne(ctx, collectionConversations, doc); err != nil {
		if err != foundationstorage.ErrDuplicateKey {
			// 非重复键错误，直接返回
			klog.CtxErrorf(ctx, "[ConvMsgStorage] insert conversation doc failed",
				"conv_id", conv.ConvID,
				"error", err.Error())
			return err
		}

		// 重复键：已有其他并发写入，尝试条件更新（只更新 last_message_seq，不更新 data）
		return s.tryUpdateConvAfterDuplicateInsert(ctx, conv.ConvID, lastMessageSeq)
	}

	// 插入成功（第一次出现这个会话记录）
	klog.CtxDebugf(ctx, "[ConvMsgStorage] insert new conversation doc success",
		"conv_id", conv.ConvID,
		"last_message_seq", lastMessageSeq)
	return nil
}

// tryUpdateConvAfterDuplicateInsert 插入遇到重复键后，尝试条件更新（只更新 last_message_seq，不更新 data）
// 可能覆盖旧 seq 或被新 seq 拒绝
func (s *convMsgStorageImpl) tryUpdateConvAfterDuplicateInsert(ctx context.Context, convID string, lastMessageSeq int64) error {
	now := time.Now()
	store := s.getStore()

	filter := bson.M{
		"_id":              convID,
		"last_message_seq": bson.M{"$lt": lastMessageSeq}, // 只有旧 seq 才允许被更新
	}
	update := bson.M{
		"$set": bson.M{
			"last_message_seq": lastMessageSeq,
			"updated_at":       now,
		},
	}

	if err := store.UpdateOne(ctx, collectionConversations, filter, update); err != nil {
		// 如果这里 ErrNotFound，说明库里的 seq 已经 >= 本次 seq，本次写是旧写，跳过即可
		if err == foundationstorage.ErrNotFound {
			klog.CtxDebugf(ctx, "[ConvMsgStorage] skip outdated conversation update after duplicate insert",
				"conv_id", convID,
				"last_message_seq", lastMessageSeq)
			return nil
		}

		klog.CtxErrorf(ctx, "[ConvMsgStorage] update conversation doc failed after duplicate insert",
			"conv_id", convID,
			"error", err.Error())
		return err
	}

	klog.CtxDebugf(ctx, "[ConvMsgStorage] update existing conversation doc success after duplicate insert",
		"conv_id", convID,
		"last_message_seq", lastMessageSeq)
	return nil
}
