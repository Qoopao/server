package service

import (
	"context"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
)

// FetchUserRecentConvList 混链拉取：获取用户最近的会话列表（包含会话和消息）
func (s *conversationServiceImpl) FetchUserRecentConvList(ctx context.Context, req *sdkws.FetchUserRecentConvListRequest) (*sdkws.FetchUserRecentConvListResponse, error) {
	userID := req.GetUserID()
	if userID == "" {
		return &sdkws.FetchUserRecentConvListResponse{
			Conversations: []*sdkws.ConversationData{},
			Left:          0,
			Right:         0,
			HasMore:       false,
		}, nil
	}

	cursor := req.GetCursor()
	limit := req.GetLimit()
	if limit <= 0 {
		limit = 20
	}
	forward := req.GetForward()
	news := req.GetNews()

	// 如果 news 为 true，表示拉取新消息，cursor 应该设为 -1 或 0
	if news {
		cursor = -1
	}

	// 调用 storage 层查询
	conversations, left, right, hasMore, err := s.storage.FetchUserRecentConvList(ctx, userID, cursor, limit, forward)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationService] fetch user recent conv list failed",
			"user_id", userID,
			"cursor", cursor,
			"limit", limit,
			"error", err.Error())
		return &sdkws.FetchUserRecentConvListResponse{
			Conversations: []*sdkws.ConversationData{},
			Left:          0,
			Right:         0,
			HasMore:       false,
		}, err
	}

	// TODO: 为每个会话获取最新消息（如果需要）
	// 这里可以根据业务需求，为每个会话获取最新 N 条消息
	// 暂时不实现，由上层业务决定

	klog.CtxDebugf(ctx, "[ConversationService] fetch user recent conv list success",
		"user_id", userID,
		"cursor", cursor,
		"limit", limit,
		"count", len(conversations),
		"has_more", hasMore,
		"left", left,
		"right", right)

	return &sdkws.FetchUserRecentConvListResponse{
		Conversations: conversations,
		Left:          left,
		Right:         right,
		HasMore:       hasMore,
	}, nil
}

