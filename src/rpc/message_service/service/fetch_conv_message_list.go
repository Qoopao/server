package service

import (
	"context"

	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
)

// FetchConvMessageList 查询会话消息列表
func (s *messageServiceImpl) FetchConvMessageList(ctx context.Context, req *sdkws.FetchConvMessageListRequest) (*sdkws.FetchConvMessageListResponse, error) {

	convID := req.GetConvID()
	cursor := req.GetCursor()
	limit := req.GetLimit()
	if limit < 1 {
		limit = 20
	}

	var (
		messages []*sdkws.MessageData
		err      error
	)

	if cursor < 0 {
		messages, err = s.storage.FetchConvLatestMessageList(ctx, convID, limit)
	} else {
		left, right := modifyFetchRange(cursor, limit, req.GetForward())
		messages, err = s.storage.FetchConvMessageListWithRange(ctx, convID, left, right)
	}

	if err != nil {
		return nil, err
	}

	return &sdkws.FetchConvMessageListResponse{
		Messages: messages,
	}, nil
}

func modifyFetchRange(cursor int64, limit int64, forward bool) (int64, int64) {
	if cursor < 0 { // 意味直接拉取最新数据
		return 0, 0
	}

	if forward {
		return min(0, cursor-limit+1), cursor
	} else {
		return cursor, cursor + limit - 1
	}
}
