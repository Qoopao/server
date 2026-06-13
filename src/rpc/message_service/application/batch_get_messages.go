package application

import (
	"context"

	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
)

// BatchGetMessages 根据消息ID列表批量获取消息详情
func (s *messageServiceImpl) BatchGetMessages(ctx context.Context, req *sdkws.BatchGetMessagesRequest) (*sdkws.BatchGetMessagesResponse, error) {
	messageIDs := req.GetMessageIDs()
	convID := req.GetConvID()

	// 通过消息仓储批量读取 message fact。
	messageMap, err := s.messageRepo.BatchGetMessages(ctx, messageIDs, convID)
	if err != nil {
		return nil, err
	}

	results := make([]*sdkws.GetMessageResult, 0, len(messageIDs))
	for _, messageID := range messageIDs {
		result := &sdkws.GetMessageResult{
			MessageID: messageID,
		}

		if msg, found := messageMap[messageID]; found {
			result.Message = msg
		} else {
			// 消息未找到
			result.ErrorCode = "MSG_NOT_FOUND"
			result.ErrorMsg = "message not found"
		}

		results = append(results, result)
	}

	return &sdkws.BatchGetMessagesResponse{
		Results: results,
	}, nil
}
