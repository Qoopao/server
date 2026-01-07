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

// FetchUserRecentConvListByVersion 根据版本号查询大于指定 version 的所有最新数据
func (s *conversationStorageImpl) FetchUserRecentConvListByVersion(ctx context.Context, userID string, version int64) ([]*sdkws.ConversationData, error) {
	if userID == "" {
		return nil, errors.New("user_id is empty")
	}

	store := s.serviceCtx.GetStorage()
	if store == nil {
		return nil, errors.New("storage is nil")
	}

	// 1. 从 user_recent_conversations 集合查询 version > 指定值的所有数据
	filter := bson.M{
		"user_id": userID,
		"version": bson.M{"$gt": version},
	}

	// 按 version 升序排序（最旧的在前）
	sort := bson.D{{Key: "version", Value: 1}}

	var userConvDocs []orm.UserRecentConversationsDocument
	err := store.Find(ctx, orm.CollectionUserRecentConversations, filter, &userConvDocs,
		foundationstorage.WithSort(sort),
	)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationStorage] fetch user recent conv list by version failed",
			"user_id", userID,
			"version", version,
			"error", err.Error())
		return nil, err
	}

	// 2. 收集所有会话ID
	convIDs := make([]string, 0, len(userConvDocs))
	for _, doc := range userConvDocs {
		convIDs = append(convIDs, doc.ConvID)
	}

	if len(convIDs) == 0 {
		return []*sdkws.ConversationData{}, nil
	}

	// 3. 批量获取会话详情
	convMap, err := s.BatchGetConversations(ctx, convIDs, "")
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationStorage] batch get conversations failed",
			"user_id", userID,
			"conv_ids", convIDs,
			"error", err.Error())
		return nil, err
	}

	// 4. 组装会话数据列表（保持 version 顺序）
	convDataList := make([]*sdkws.ConversationData, 0, len(userConvDocs))
	for _, userConvDoc := range userConvDocs {
		convData, exists := convMap[userConvDoc.ConvID]
		if !exists {
			klog.CtxWarnf(ctx, "[ConversationStorage] conversation not found",
				"user_id", userID,
				"conv_id", userConvDoc.ConvID)
			continue
		}

		convDataList = append(convDataList, convData)
	}

	klog.CtxDebugf(ctx, "[ConversationStorage] fetch user recent conv list by version success",
		"user_id", userID,
		"version", version,
		"count", len(convDataList))

	return convDataList, nil
}

// FetchUserRecentConvListByVersionRange 根据版本号区间查询数据
func (s *conversationStorageImpl) FetchUserRecentConvListByVersionRange(ctx context.Context, userID string, lowVersion int64, upVersion int64) ([]*sdkws.ConversationData, error) {
	if userID == "" {
		return nil, errors.New("user_id is empty")
	}

	store := s.serviceCtx.GetStorage()
	if store == nil {
		return nil, errors.New("storage is nil")
	}

	// 1. 从 user_recent_conversations 集合查询版本号在区间内的数据
	filter := bson.M{
		"user_id": userID,
	}

	// 构建版本号区间过滤条件
	versionFilter := bson.M{}
	if lowVersion > 0 {
		versionFilter["$gte"] = lowVersion
	}
	if upVersion > 0 {
		versionFilter["$lte"] = upVersion
	}
	if len(versionFilter) > 0 {
		filter["version"] = versionFilter
	}

	// 按 version 升序排序
	sort := bson.D{{Key: "version", Value: 1}}

	var userConvDocs []orm.UserRecentConversationsDocument
	err := store.Find(ctx, orm.CollectionUserRecentConversations, filter, &userConvDocs,
		foundationstorage.WithSort(sort),
	)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationStorage] fetch user recent conv list by version range failed",
			"user_id", userID,
			"low_version", lowVersion,
			"up_version", upVersion,
			"error", err.Error())
		return nil, err
	}

	// 2. 收集所有会话ID
	convIDs := make([]string, 0, len(userConvDocs))
	for _, doc := range userConvDocs {
		convIDs = append(convIDs, doc.ConvID)
	}

	if len(convIDs) == 0 {
		return []*sdkws.ConversationData{}, nil
	}

	// 3. 批量获取会话详情
	convMap, err := s.BatchGetConversations(ctx, convIDs, "")
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationStorage] batch get conversations failed",
			"user_id", userID,
			"conv_ids", convIDs,
			"error", err.Error())
		return nil, err
	}

	// 4. 组装会话数据列表（保持 version 顺序）
	convDataList := make([]*sdkws.ConversationData, 0, len(userConvDocs))
	for _, userConvDoc := range userConvDocs {
		convData, exists := convMap[userConvDoc.ConvID]
		if !exists {
			klog.CtxWarnf(ctx, "[ConversationStorage] conversation not found",
				"user_id", userID,
				"conv_id", userConvDoc.ConvID)
			continue
		}

		convDataList = append(convDataList, convData)
	}

	klog.CtxDebugf(ctx, "[ConversationStorage] fetch user recent conv list by version range success",
		"user_id", userID,
		"low_version", lowVersion,
		"up_version", upVersion,
		"count", len(convDataList))

	return convDataList, nil
}
