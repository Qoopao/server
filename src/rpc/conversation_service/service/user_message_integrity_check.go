package service

import (
	"context"
	"math"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
)

// UserMessageIntegrityCheck 混链连续性检查：检查用户消息的完整性，并补齐缺失的会话信息
func (s *conversationServiceImpl) UserMessageIntegrityCheck(ctx context.Context, req *sdkws.UserMessageIntegrityCheckRequest) (*sdkws.UserMessageIntegrityCheckResponse, error) {
	userID := req.GetUserID()
	left := req.GetLeft()
	right := req.GetRight()
	convIDs := req.GetConvIDs()

	if userID == "" {
		return &sdkws.UserMessageIntegrityCheckResponse{
			IsIntegrity:   false,
			Left:          0,
			Right:         0,
			Conversations: []*sdkws.ConversationData{},
		}, nil
	}

	// 1. 从 user_recent_conversations 中拉取 [left, right] 版本区间内的会话视图
	serverConvs, err := s.storage.FetchUserRecentConvListByVersionRange(ctx, userID, left, right)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationService] UserMessageIntegrityCheck fetch recent convs failed",
			"user_id", userID,
			"left", left,
			"right", right,
			"error", err.Error())
		return &sdkws.UserMessageIntegrityCheckResponse{
			IsIntegrity:   false,
			Left:          left,
			Right:         right,
			Conversations: []*sdkws.ConversationData{},
		}, nil
	}

	// 2. 计算服务端视角下该区间的最小/最大版本，以及会话 ID 集合
	serverMin := int64(math.MaxInt64)
	serverMax := int64(math.MinInt64)
	serverConvSet := make(map[string]struct{}, len(serverConvs))
	for _, conv := range serverConvs {
		version := conv.GetVersion()
		if version < serverMin {
			serverMin = version
		}
		if version > serverMax {
			serverMax = version
		}
		serverConvSet[conv.GetConvID()] = struct{}{}
	}

	// 若区间内无会话，则认为是完整的（当前视图为空）
	if len(serverConvs) == 0 {
		return &sdkws.UserMessageIntegrityCheckResponse{
			IsIntegrity:   true,
			Left:          left,
			Right:         right,
			Conversations: []*sdkws.ConversationData{},
		}, nil
	}

	// 3. 检查客户端上报的 convIDs 是否覆盖服务端视角下的会话
	clientConvSet := make(map[string]struct{}, len(convIDs))
	for _, id := range convIDs {
		clientConvSet[id] = struct{}{}
	}

	missingConv := false
	for id := range serverConvSet {
		if _, ok := clientConvSet[id]; !ok && len(convIDs) > 0 {
			missingConv = true
			break
		}
	}

	// 4. 依据版本范围与会话缺失情况给出完整性结论
	isIntegrity := !missingConv && (left <= serverMin) && (right >= serverMax)

	respLeft := left
	respRight := right
	if !isIntegrity {
		respLeft = serverMin
		respRight = serverMax
	}

	klog.CtxDebugf(ctx, "[ConversationService] UserMessageIntegrityCheck finished",
		"user_id", userID,
		"req_left", left,
		"req_right", right,
		"resp_left", respLeft,
		"resp_right", respRight,
		"conv_count", len(serverConvs),
		"is_integrity", isIntegrity)

	return &sdkws.UserMessageIntegrityCheckResponse{
		IsIntegrity:   isIntegrity,
		Left:          respLeft,
		Right:         respRight,
		Conversations: serverConvs,
	}, nil
}
