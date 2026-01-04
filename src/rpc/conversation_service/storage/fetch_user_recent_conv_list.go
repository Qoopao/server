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

// FetchUserRecentConvList 获取用户最近的会话列表（包含会话和消息）
func (s *conversationStorageImpl) FetchUserRecentConvList(ctx context.Context, userID string, cursor int64, limit int64, forward bool) ([]*sdkws.ConversationData, int64, int64, bool, error) {
	if userID == "" {
		return nil, 0, 0, false, errors.New("user_id is empty")
	}

	if limit <= 0 {
		limit = 20
	}

	store := s.serviceCtx.GetStorage()
	if store == nil {
		return nil, 0, 0, false, errors.New("storage is nil")
	}

	// 1. 从 user_recent_conversations 集合查询用户的最近会话列表
	filter := bson.M{"user_id": userID}

	// cursor 是 updated_at 的时间戳（毫秒），如果 cursor > 0，则添加时间过滤条件
	var cursorTime time.Time
	if cursor > 0 {
		cursorTime = time.Unix(cursor/1000, (cursor%1000)*1000000)
		if forward {
			// 向前查询（新会话）：updated_at > cursorTime
			filter["updated_at"] = bson.M{"$gt": cursorTime}
		} else {
			// 向后查询（旧会话）：updated_at < cursorTime
			filter["updated_at"] = bson.M{"$lt": cursorTime}
		}
	}

	// 按 updated_at 降序排序（最新的在前）
	sort := bson.D{{Key: "updated_at", Value: -1}}

	var userConvDocs []orm.UserRecentConversationsDocument
	err := store.Find(ctx, collectionUserRecentConversations, filter, &userConvDocs,
		foundationstorage.WithLimit(limit+1), // 多查一条用于判断是否有更多
		foundationstorage.WithSort(sort),
	)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationStorage] fetch user recent conv list failed",
			"user_id", userID,
			"cursor", cursor,
			"limit", limit,
			"error", err.Error())
		return nil, 0, 0, false, err
	}

	// 判断是否有更多数据
	hasMore := len(userConvDocs) > int(limit)
	if hasMore {
		userConvDocs = userConvDocs[:limit]
	}

	// 计算 left 和 right（时间戳范围）
	var left, right int64
	if len(userConvDocs) > 0 {
		left = userConvDocs[len(userConvDocs)-1].UpdatedAt.UnixMilli()
		right = userConvDocs[0].UpdatedAt.UnixMilli()
	}

	// 2. 收集所有会话ID
	convIDs := make([]string, 0, len(userConvDocs))
	for _, doc := range userConvDocs {
		convIDs = append(convIDs, doc.ConvID)
	}

	if len(convIDs) == 0 {
		return []*sdkws.ConversationData{}, left, right, hasMore, nil
	}

	// 3. 批量获取会话详情
	convMap, err := s.BatchGetConversations(ctx, convIDs, "")
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationStorage] batch get conversations failed",
			"user_id", userID,
			"conv_ids", convIDs,
			"error", err.Error())
		return nil, left, right, hasMore, err
	}

	// 4. 为每个会话获取最新消息（可选，这里获取最新1条）
	// 注意：根据业务需求，可能需要获取更多消息或跳过此步骤
	convDataList := make([]*sdkws.ConversationData, 0, len(userConvDocs))
	for _, userConvDoc := range userConvDocs {
		convData, exists := convMap[userConvDoc.ConvID]
		if !exists {
			klog.CtxWarnf(ctx, "[ConversationStorage] conversation not found",
				"user_id", userID,
				"conv_id", userConvDoc.ConvID)
			continue
		}

		// 获取会话的最新消息（可选）
		// 这里可以根据业务需求决定是否获取消息
		// 暂时不获取消息，由上层业务决定

		convDataList = append(convDataList, convData)
	}

	klog.CtxDebugf(ctx, "[ConversationStorage] fetch user recent conv list success",
		"user_id", userID,
		"cursor", cursor,
		"limit", limit,
		"count", len(convDataList),
		"has_more", hasMore,
		"left", left,
		"right", right)

	return convDataList, left, right, hasMore, nil
}

