package storage

import (
	"context"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
	"go.mongodb.org/mongo-driver/bson"
)

// UpdateUserRecentConversation 更新用户最近会话链（将会话移到最前面）
func (s *convMsgStorageImpl) UpdateUserRecentConversation(ctx context.Context, userID string, convID string, lastSeq int64) error {
	if userID == "" || convID == "" {
		return nil
	}

	store := s.getStore()
	filter := bson.M{"_id": userID}
	var doc orm.UserRecentConversationsDocument
	err := store.FindOne(ctx, collectionUserRecentConversations, filter, &doc)

	// 检查是否是文档不存在的错误
	if err != nil {
		if err == foundationstorage.ErrNotFound {
			// 文档不存在，创建新文档
			return s.createUserRecentConversations(ctx, userID, convID, lastSeq)
		}
		klog.CtxErrorf(ctx, "[ConvMsgStorage] find user recent conversations failed",
			"user_id", userID,
			"error", err.Error())
		return err
	}

	// 文档存在，更新数组
	return s.updateUserRecentConversations(ctx, userID, convID, lastSeq)
}

// createUserRecentConversations 创建用户最近会话链文档
func (s *convMsgStorageImpl) createUserRecentConversations(ctx context.Context, userID string, convID string, lastSeq int64) error {
	now := time.Now()
	store := s.getStore()

	newItem := orm.ConversationItem{
		ConvID:    convID,
		LastSeq:   lastSeq,
		UpdatedAt: now,
	}
	newDoc := &orm.UserRecentConversationsDocument{
		ID:            userID,
		Conversations: []orm.ConversationItem{newItem},
	}

	_, err := store.InsertOne(ctx, collectionUserRecentConversations, newDoc)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConvMsgStorage] create user recent conversations failed",
			"user_id", userID,
			"conv_id", convID,
			"error", err.Error())
		return err
	}

	klog.CtxDebugf(ctx, "[ConvMsgStorage] create user recent conversations success",
		"user_id", userID,
		"conv_id", convID)
	return nil
}

// updateUserRecentConversations 更新用户最近会话链（先删除再添加到最前面）
func (s *convMsgStorageImpl) updateUserRecentConversations(ctx context.Context, userID string, convID string, lastSeq int64) error {
	filter := bson.M{"_id": userID}

	// 1. 先删除已存在的该会话（如果存在）
	if err := s.removeConversationFromList(ctx, filter, convID); err != nil {
		return err
	}

	// 2. 再将该会话添加到数组最前面
	if err := s.addConversationToList(ctx, filter, convID, lastSeq); err != nil {
		return err
	}

	klog.CtxDebugf(ctx, "[ConvMsgStorage] update user recent conversations success",
		"user_id", userID,
		"conv_id", convID)
	return nil
}

// removeConversationFromList 从用户最近会话链中删除指定会话
func (s *convMsgStorageImpl) removeConversationFromList(ctx context.Context, filter bson.M, convID string) error {
	store := s.getStore()
	update := bson.M{
		"$pull": bson.M{
			"conversations": bson.M{"conv_id": convID},
		},
	}
	err := store.UpdateOne(ctx, collectionUserRecentConversations, filter, update)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConvMsgStorage] remove conversation from user list failed",
			"conv_id", convID,
			"error", err.Error())
		return err
	}
	return nil
}

// addConversationToList 将会话添加到用户最近会话链的最前面
func (s *convMsgStorageImpl) addConversationToList(ctx context.Context, filter bson.M, convID string, lastSeq int64) error {
	now := time.Now()
	store := s.getStore()

	newItem := orm.ConversationItem{
		ConvID:    convID,
		LastSeq:   lastSeq,
		UpdatedAt: now,
	}
	update := bson.M{
		"$push": bson.M{
			"conversations": bson.M{
				"$each":     []orm.ConversationItem{newItem},
				"$position": 0,
			},
		},
	}
	err := store.UpdateOne(ctx, collectionUserRecentConversations, filter, update)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConvMsgStorage] add conversation to user list failed",
			"conv_id", convID,
			"error", err.Error())
		return err
	}
	return nil
}
