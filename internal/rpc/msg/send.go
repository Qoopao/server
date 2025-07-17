//
// Author : Ruanhuipeng
// Date   : 06/08/2025

package msg

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/openimsdk/tools/errs"
	"github.com/roc/roc-im-server/internal/kitex_gen/sdkws"
)

// SendMsg implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) sendMessages(ctx context.Context, req *sdkws.SendMessageReq) (resp *sdkws.SendMessageResp, err error) {
	// 1、check
	if len(req.Msgs) == 0 {
		return nil, errs.ErrArgs.WrapMsg("msgs is empty")
	}

	// 构造返回信息
	respInfos := make([]*sdkws.SendMessageRespInfo, 0, len(req.Msgs))
	for index := range req.Msgs {
		respInfos[index] = &sdkws.SendMessageRespInfo{
			ServerMsgID: "",
			ClientMsgID: "",
			SendTime:    0,
			IsSuccess:   true,
			ErrorCode:   "",
		}
	}

	// 生成消息ID
	for index, msg := range req.Msgs {
		msg.ServerMsgID = uuid.New().String()

		respInfos[index].ServerMsgID = msg.ServerMsgID
		respInfos[index].ClientMsgID = msg.ClientMsgID
	}

	// 追加到单链（会话链） 并生成orderIndex
	for index, msg := range req.Msgs {
		orderIndex, err := s.MsgDatabase.AppendMsgToConvMsgList(ctx, msg.ConvID, msg.ServerMsgID)
		if err != nil {
			respInfos[index].IsSuccess = false
			respInfos[index].ErrorCode = err.Error()
			continue
		}

		msg.ServerOrdIndex = orderIndex
	}

	// 存储消息
	for _, msg := range req.Msgs {
		s.MsgDatabase.SaveMsgToDB(ctx, msg)
	}

	// 转发消息到MQ
	for _, msg := range req.Msgs {
		s.MsgDatabase.MsgToMQ(ctx, msg.SendID, msg)
	}

	resp = &sdkws.SendMessageResp{
		Infos: respInfos,
	}
	return resp, nil
}
