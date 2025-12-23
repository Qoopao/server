package storage

import (
	"context"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
	"go.mongodb.org/mongo-driver/bson"
)

// UpdateUserRecentConversation 更新用户最近会话链（一个 user+conv 一条文档，按 Version 排序）
// Version 字段作用：1) 防止旧状态覆盖新状态（并发安全） 2) 客户端增量同步（通过 version 判断会话是否有更新）
// TODO: Version 当前使用时间戳，后续改为用户维度的序列号服务生成
func (s *convMsgStorageImpl) UpdateUserRecentConversation(ctx context.Context, userID string, convID string, lastSeq int64) error {
	if userID == "" || convID == "" {
		return nil
	}

	now := time.Now()
	doc := &orm.UserRecentConversationsDocument{
		ID:             userID + ":" + convID,
		UserID:         userID,
		ConvID:         convID,
		LastMessageSeq: lastSeq,
		Version:        now.UnixNano(), // TODO: 后续改为用户维度的序列号服务生成
		UpdatedAt:      now,
	}

	// 1. 先尝试更新已有文档（正常路径：绝大部分都是已经存在的会话）
	if updated, err := s.tryUpdateExisting(ctx, doc); err != nil {
		return err
	} else if updated {
		return nil
	}

	// 2. 更新失败（文档不存在或版本已更新），尝试插入新文档
	return s.tryInsertNew(ctx, doc)
}

// tryUpdateExisting 尝试更新已有文档（只有当旧 Version < 新 Version 时才允许被覆盖）
func (s *convMsgStorageImpl) tryUpdateExisting(ctx context.Context, doc *orm.UserRecentConversationsDocument) (bool, error) {
	store := s.getStore()

	filter := bson.M{
		"_id":     doc.ID,
		"version": bson.M{"$lt": doc.Version}, // 只有旧版本才允许被更新
	}
	update := bson.M{
		"$set": bson.M{
			"user_id":          doc.UserID,
			"conv_id":          doc.ConvID,
			"last_message_seq": doc.LastMessageSeq,
			"version":          doc.Version,
			"updated_at":       doc.UpdatedAt,
		},
	}

	if err := store.UpdateOne(ctx, collectionUserRecentConversations, filter, update); err == nil {
		// 更新成功（命中了旧版本）
		klog.CtxDebugf(ctx, "[ConvMsgStorage] update existing user recent conversation doc success",
			"user_id", doc.UserID,
			"conv_id", doc.ConvID,
			"last_message_seq", doc.LastMessageSeq,
			"version", doc.Version)
		return true, nil
	} else if err != foundationstorage.ErrNotFound {
		// 真实错误，直接返回
		klog.CtxErrorf(ctx, "[ConvMsgStorage] update user recent conversation doc failed",
			"user_id", doc.UserID,
			"conv_id", doc.ConvID,
			"error", err.Error())
		return false, err
	}

	// ErrNotFound：文档不存在或版本已 >= 本次版本
	return false, nil
}

// tryInsertNew 尝试插入新文档，如果遇到重复键则尝试条件更新
func (s *convMsgStorageImpl) tryInsertNew(ctx context.Context, doc *orm.UserRecentConversationsDocument) error {
	store := s.getStore()

	if _, err := store.InsertOne(ctx, collectionUserRecentConversations, doc); err != nil {
		if err != foundationstorage.ErrDuplicateKey {
			// 非重复键错误，直接返回
			klog.CtxErrorf(ctx, "[ConvMsgStorage] insert user recent conversation doc failed",
				"user_id", doc.UserID,
				"conv_id", doc.ConvID,
				"error", err.Error())
			return err
		}

		// 重复键：已有其他并发写入，尝试条件更新
		return s.tryUpdateUserRecentConvAfterDuplicateInsert(ctx, doc)
	}

	// 插入成功（第一次出现这个 user+conv 的会话记录）
	klog.CtxDebugf(ctx, "[ConvMsgStorage] insert new user recent conversation doc success",
		"user_id", doc.UserID,
		"conv_id", doc.ConvID,
		"last_message_seq", doc.LastMessageSeq,
		"version", doc.Version)
	return nil
}

// tryUpdateUserRecentConvAfterDuplicateInsert 插入遇到重复键后，尝试条件更新（可能覆盖旧版本或被新版本拒绝）
func (s *convMsgStorageImpl) tryUpdateUserRecentConvAfterDuplicateInsert(ctx context.Context, doc *orm.UserRecentConversationsDocument) error {
	store := s.getStore()

	filter := bson.M{
		"_id":     doc.ID,
		"version": bson.M{"$lt": doc.Version}, // 只有旧版本才允许被更新
	}
	update := bson.M{
		"$set": bson.M{
			"user_id":          doc.UserID,
			"conv_id":          doc.ConvID,
			"last_message_seq": doc.LastMessageSeq,
			"version":          doc.Version,
			"updated_at":       doc.UpdatedAt,
		},
	}

	if err := store.UpdateOne(ctx, collectionUserRecentConversations, filter, update); err != nil {
		// 如果这里 ErrNotFound，说明库里的 version 已经 >= 本次 version，本次写是旧写，跳过即可
		if err == foundationstorage.ErrNotFound {
			klog.CtxDebugf(ctx, "[ConvMsgStorage] skip outdated user recent conversation update after duplicate insert",
				"user_id", doc.UserID,
				"conv_id", doc.ConvID,
				"last_message_seq", doc.LastMessageSeq,
				"version", doc.Version)
			return nil
		}

		klog.CtxErrorf(ctx, "[ConvMsgStorage] update user recent conversation doc failed after duplicate insert",
			"user_id", doc.UserID,
			"conv_id", doc.ConvID,
			"error", err.Error())
		return err
	}

	klog.CtxDebugf(ctx, "[ConvMsgStorage] update existing user recent conversation doc success after duplicate insert",
		"user_id", doc.UserID,
		"conv_id", doc.ConvID,
		"last_message_seq", doc.LastMessageSeq,
		"version", doc.Version)
	return nil
}
