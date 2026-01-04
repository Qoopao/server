package service

import (
	"context"

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

	// TODO: 实现消息完整性检查逻辑
	// 1. 检查用户在 [left, right] 时间范围内的会话列表是否完整
	// 2. 如果 convIDs 不为空，检查这些会话是否存在
	// 3. 如果发现缺失的会话，补齐会话信息
	// 目前先返回未实现的结果

	klog.CtxWarnf(ctx, "[ConversationService] UserMessageIntegrityCheck not fully implemented",
		"user_id", userID,
		"left", left,
		"right", right,
		"conv_ids", convIDs)

	// 如果提供了 convIDs，批量获取这些会话
	var conversations []*sdkws.ConversationData
	if len(convIDs) > 0 {
		convMap, err := s.storage.BatchGetConversations(ctx, convIDs, "")
		if err != nil {
			klog.CtxErrorf(ctx, "[ConversationService] batch get conversations failed in integrity check",
				"user_id", userID,
				"conv_ids", convIDs,
				"error", err.Error())
		} else {
			// 转换为列表
			conversations = make([]*sdkws.ConversationData, 0, len(convMap))
			for _, convData := range convMap {
				conversations = append(conversations, convData)
			}
		}
	}

	// 暂时假设完整性检查通过（实际需要实现具体的检查逻辑）
	isIntegrity := true

	return &sdkws.UserMessageIntegrityCheckResponse{
		IsIntegrity:   isIntegrity,
		Left:          left,
		Right:         right,
		Conversations: conversations,
	}, nil
}

