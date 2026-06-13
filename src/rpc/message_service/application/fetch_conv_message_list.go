package application

import (
	"context"

	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
)

// FetchConvMessageList 查询会话消息列表
// 新版协议：使用 left/right 指定区间；left/right 都为 0 时表示按 limit 拉取最新消息
func (s *messageServiceImpl) FetchConvMessageList(ctx context.Context, req *sdkws.FetchConvMessageListRequest) (*sdkws.FetchConvMessageListResponse, error) {
	var (
		messages []*sdkws.MessageData
		err      error
	)

	// left/right 都为 0：按 limit 拉取最新消息
	if req.GetMode() == 0 {
		messages, err = s.messageRepo.FetchConvLatestMessageList(ctx, req.GetConvID(), req.GetLimit())
	} else {
		messages, err = s.messageRepo.FetchConvMessageListWithRange(ctx, req.GetConvID(), req.GetLeft(), req.GetRight())
	}
	if err != nil {
		return nil, err
	}

	var respLeft, respRight int64
	if len(messages) > 0 {
		respLeft = messages[0].GetSeq()
		respRight = messages[len(messages)-1].GetSeq()
	}

	return &sdkws.FetchConvMessageListResponse{
		Messages: messages,
		Left:     respLeft,
		Right:    respRight,
		HaveMore: false,
	}, nil
}
