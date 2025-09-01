//
// Author : Ruanhuipeng
// Date   : 06/08/2025

package msg

import (
	"context"

	"github.com/google/uuid"
	"github.com/openimsdk/tools/errs"
	"github.com/roc/roc-im-server/internal/kitex_gen/conversation"
	"github.com/roc/roc-im-server/internal/kitex_gen/sdkws"
	"go.uber.org/zap"
)

// SendMsg implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) sendMessages(ctx context.Context, req *sdkws.SendMessageReq) (resp *sdkws.SendMessageResp, err error) {
	// 1、check
	if len(req.Msgs) == 0 {
		Logger.Warn("sendMessages: msgs is empty")
		return nil, errs.ErrArgs.WrapMsg("msgs is empty")
	}

	// 构造返回信息
	respInfos := make([]*sdkws.SendMessageRespInfo, len(req.Msgs))
	for index := range req.Msgs {
		respInfos[index] = &sdkws.SendMessageRespInfo{
			ErrorCode: "0",
			ErrorMsg:  "",
		}
	}

	// 生成消息ID
	for index, msg := range req.Msgs {
		// 如果conv_id为空，则失败
		if msg.ConvID == "" {
			Logger.Error("sendMessages: conv_id is empty",
				zap.String("clientMsgID", msg.ClientMsgID),
				zap.String("sendID", msg.SendID))
			respInfos[index].ErrorCode = "1"
			respInfos[index].ErrorMsg = "conv_id is empty"
			continue
		}

		// 生成server_msg_id
		msg.ServerMsgID = uuid.New().String()

		// 生成orderIndex
		orderIndex, err := s.MsgDatabase.AppendMsgToConvMsgList(ctx, msg.ConvID, msg.ServerMsgID)
		if err != nil {
			Logger.Error("sendMessages: failed to append msg to conv msg list",
				zap.String("convID", msg.ConvID),
				zap.String("serverMsgID", msg.ServerMsgID),
				zap.Error(err))
			respInfos[index].ErrorCode = "1"
			respInfos[index].ErrorMsg = err.Error()
			continue
		}
		msg.Seq = orderIndex

		// 存储消息
		if err := s.MsgDatabase.SaveMsgInfo(ctx, msg); err != nil {
			Logger.Error("sendMessages: failed to save msg info",
				zap.String("convID", msg.ConvID),
				zap.String("serverMsgID", msg.ServerMsgID),
				zap.Error(err))
			respInfos[index].ErrorCode = "1"
			respInfos[index].ErrorMsg = err.Error()
			continue
		}

		// 如果没有会话则创建会话
		if err := s.createConversationIfNeed(ctx, msg); err != nil {
			Logger.Error("sendMessages: failed to create conversation",
				zap.String("convID", msg.ConvID),
				zap.String("sendID", msg.SendID),
				zap.Error(err))
			respInfos[index].ErrorCode = "1"
			respInfos[index].ErrorMsg = err.Error()
			continue
		}

		// 转发消息到MQ
		if err := s.MsgDatabase.MsgToMQ(ctx, msg.SendID, msg.ServerMsgID); err != nil {
			Logger.Error("sendMessages: failed to send msg to MQ",
				zap.String("convID", msg.ConvID),
				zap.String("serverMsgID", msg.ServerMsgID),
				zap.String("sendID", msg.SendID),
				zap.Error(err))
			respInfos[index].ErrorCode = "1"
			respInfos[index].ErrorMsg = err.Error()
			continue
		}

		// 设置返回信息
		respInfos[index].Msg = msg

		Logger.Debug("sendMessages: message processed successfully",
			zap.String("convID", msg.ConvID),
			zap.String("serverMsgID", msg.ServerMsgID),
			zap.Int64("seq", msg.Seq))
	}

	resp = &sdkws.SendMessageResp{
		Infos: respInfos,
	}
	return resp, nil
}

func (s *MessageServiceImpl) createConversationIfNeed(ctx context.Context, msg *sdkws.MsgData) error {
	conv, _ := s.MsgDatabase.GetConversationInfo(ctx, msg.ConvID)

	if conv == nil {
		Logger.Info("createConversationIfNeed: creating new conversation",
			zap.String("convID", msg.ConvID),
			zap.String("ownerUserID", msg.SendID))

		conv = &conversation.ConversationInfo{
			ConversationID:     msg.ConvID,
			OwnerUserID:        msg.SendID,
			ConversationType:   0,
			ConversationName:   "",
			ConversationAvatar: "",
		}
	}

	// TODO: 修改会话的最后一条消息

	// TODO: 扩展到群聊
	return s.MsgDatabase.SaveConversationInfo(ctx, msg.ConvID, conv)
}
