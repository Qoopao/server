package service

import (
	"context"
	"errors"
	"math"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-foundation-util-go/stringutil"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
)

// FetchUserRecentConvList 混链拉取：获取用户最近的会话列表（包含会话和消息）
func (s *conversationServiceImpl) FetchUserRecentConvList(ctx context.Context, req *sdkws.FetchUserRecentConvListRequest) (*sdkws.FetchUserRecentConvListResponse, error) {
	if err := checkRequestParams(ctx, req); err != nil {
		return &sdkws.FetchUserRecentConvListResponse{
			Error: err.Error(),
		}, nil
	}

	userID := req.GetUserID()
	lowVersion := req.GetLowerVersion()
	upVersion := req.GetUpperVersion()
	first := req.GetFirst()

	var (
		conversations []*sdkws.ConversationData
		err           error
	)

	// 根据 first 标志决定查询方式
	if first {
		// 第一次拉取：查询大于 lowVersion 的所有最新数据
		conversations, err = s.storage.FetchUserRecentConvListByVersion(ctx, userID, upVersion)
	} else {
		// 增量拉取：查询 [lowVersion, upVersion] 区间内的数据
		conversations, err = s.storage.FetchUserRecentConvListByVersionRange(ctx, userID, lowVersion, upVersion)
	}

	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationService] fetch user recent conv list failed",
			"user_id", userID,
			"low_version", lowVersion,
			"up_version", upVersion,
			"first", first,
			"error", err.Error())
		return &sdkws.FetchUserRecentConvListResponse{
			Conversations: []*sdkws.ConversationData{},
			Left:          0,
			Right:         0,
			Error:         err.Error(),
		}, nil
	}

	// 从会话列表中计算最小和最大版本号
	var left, right int64
	if len(conversations) > 0 {
		// 为每个会话附带一小段最近消息 window
		s.attachRecentMessages(ctx, conversations)

		left = int64(math.MaxInt64)
		right = int64(math.MinInt64)
		for _, conv := range conversations {
			version := conv.GetVersion()
			left = min(left, version)
			right = max(right, version)
		}
	}

	klog.CtxDebugf(ctx, "[ConversationService] fetch user recent conv list success",
		"user_id", userID,
		"low_version", lowVersion,
		"up_version", upVersion,
		"first", first,
		"count", len(conversations),
		"left", left,
		"right", right)

	return &sdkws.FetchUserRecentConvListResponse{
		Conversations: conversations,
		Left:          left,
		Right:         right,
	}, nil
}

func checkRequestParams(ctx context.Context, req *sdkws.FetchUserRecentConvListRequest) error {
	if stringutil.IsEmpty(req.UserID) {
		return errors.New("uid is empty")
	}

	return nil
}

// attachRecentMessages 为每个会话附带固定数量的最近消息
func (s *conversationServiceImpl) attachRecentMessages(ctx context.Context, conversations []*sdkws.ConversationData) {
	const perConvMsgLimit int64 = 20

	if len(conversations) == 0 || perConvMsgLimit <= 0 {
		return
	}

	client, err := s.serviceCtx.GetMessageServiceClient()
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationService] get message service client failed: %v", err)
		return
	}

	for _, conv := range conversations {
		if conv == nil {
			continue
		}
		convID := conv.GetConvID()
		if convID == "" {
			continue
		}

		msgReq := &sdkws.FetchConvMessageListRequest{
			ConvID:  convID,
			Cursor:  -1,              // 从最新消息开始
			Limit:   perConvMsgLimit, // 每个会话固定条数
			Forward: false,           // 默认按时间从旧到新返回
		}

		msgResp, err := client.FetchConvMessageList(ctx, msgReq)
		if err != nil {
			klog.CtxErrorf(ctx, "[ConversationService] fetch conv messages failed",
				"conv_id", convID,
				"error", err.Error())
			continue
		}

		conv.Messages = msgResp.GetMessages()
	}
}
