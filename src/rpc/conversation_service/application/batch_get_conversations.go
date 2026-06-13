package application

import (
	"context"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
)

// BatchGetConversations 批量获取会话：根据会话ID列表批量获取会话详情
func (s *conversationUsecase) BatchGetConversations(ctx context.Context, req *sdkws.BatchGetConversationsRequest) (*sdkws.BatchGetConversationsResponse, error) {
	convIDs := req.GetConvIDs()
	ownerID := req.GetOwnerID()

	if len(convIDs) == 0 {
		return &sdkws.BatchGetConversationsResponse{
			Results: []*sdkws.GetConversationResult{},
		}, nil
	}

	// 调用 storage 层批量查询
	convMap, err := s.repo.BatchGetConversations(ctx, convIDs, ownerID)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationUsecase] batch get conversations failed",
			"conv_ids", convIDs,
			"owner_id", ownerID,
			"error", err.Error())
		return &sdkws.BatchGetConversationsResponse{
			Results: []*sdkws.GetConversationResult{},
		}, err
	}

	// 构建结果列表
	results := make([]*sdkws.GetConversationResult, 0, len(convIDs))
	for _, convID := range convIDs {
		result := &sdkws.GetConversationResult{
			ConvID: convID,
		}

		if convData, found := convMap[convID]; found {
			result.Conversation = convData
		} else {
			// 会话未找到
			result.ErrorCode = "CONV_NOT_FOUND"
			result.ErrorMsg = "conversation not found"
		}

		results = append(results, result)
	}

	klog.CtxDebugf(ctx, "[ConversationUsecase] batch get conversations success",
		"request_count", len(convIDs),
		"found_count", len(convMap),
		"owner_id", ownerID)

	return &sdkws.BatchGetConversationsResponse{
		Results: results,
	}, nil
}
