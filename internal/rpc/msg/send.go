//
// Author : Ruanhuipeng
// Date   : 06/08/2025

package msg

import (
	"context"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/google/uuid"
	"github.com/openimsdk/tools/errs"
	"github.com/rhp-QE/roc-im-server/internal/kitex_gen/conversation"
	"github.com/rhp-QE/roc-im-server/internal/kitex_gen/sdkws"
)

// SendMsg implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) sendMessages(ctx context.Context, req *sdkws.SendMessageReq) (resp *sdkws.SendMessageResp, err error) {
	// 1、check
	if len(req.Msgs) == 0 {
		klog.CtxErrorf(ctx, "sendMessages: msgs is empty")
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
			klog.CtxErrorf(ctx, "sendMessages: conv_id is empty",
				"client_msg_id", msg.ClientMsgID,
				"send_id", msg.SendID,
				"msg_index", index)

			respInfos[index].ErrorCode = "1"
			respInfos[index].ErrorMsg = "conv_id is empty"
			continue
		}

		// 生成server_msg_id
		msg.ServerMsgID = uuid.New().String()

		// 生成orderIndex
		orderIndex, err := s.MsgDatabase.AppendMsgToConvMsgList(ctx, msg.ConvID, msg.ServerMsgID)
		if err != nil {
			klog.CtxErrorf(ctx, "sendMessages: failed to append msg to conv msg list",
				"conv_id", msg.ConvID,
				"server_msg_id", msg.ServerMsgID,
				"client_msg_id", msg.ClientMsgID,
				"send_id", msg.SendID,
				"msg_index", index,
				"error", err.Error())

			respInfos[index].ErrorCode = "1"
			respInfos[index].ErrorMsg = err.Error()
			continue
		}
		msg.Seq = orderIndex

		// 存储消息
		if err := s.MsgDatabase.SaveMsgInfo(ctx, msg); err != nil {
			klog.CtxErrorf(ctx, "sendMessages: failed to save msg info",
				"conv_id", msg.ConvID,
				"server_msg_id", msg.ServerMsgID,
				"client_msg_id", msg.ClientMsgID,
				"send_id", msg.SendID,
				"msg_index", index,
				"error", err.Error())

			respInfos[index].ErrorCode = "1"
			respInfos[index].ErrorMsg = err.Error()
			continue
		}

		// 如果没有会话则创建会话
		if err := s.createConversationIfNeed(ctx, msg); err != nil {
			klog.CtxErrorf(ctx, "sendMessages: failed to create conversation",
				"conv_id", msg.ConvID,
				"send_id", msg.SendID,
				"client_msg_id", msg.ClientMsgID,
				"server_msg_id", msg.ServerMsgID,
				"msg_index", index,
				"error", err.Error())

			respInfos[index].ErrorCode = "1"
			respInfos[index].ErrorMsg = err.Error()
			continue
		}

		// 转发消息到MQ
		if err := s.MsgDatabase.MsgToMQ(ctx, msg.SendID, msg.ServerMsgID); err != nil {
			klog.CtxErrorf(ctx, "sendMessages: failed to send msg to MQ",
				"conv_id", msg.ConvID,
				"server_msg_id", msg.ServerMsgID,
				"send_id", msg.SendID,
				"client_msg_id", msg.ClientMsgID,
				"msg_index", index,
				"error", err.Error())

			respInfos[index].ErrorCode = "1"
			respInfos[index].ErrorMsg = err.Error()
			continue
		}

		// 设置返回信息
		respInfos[index].Msg = msg

		klog.CtxDebugf(ctx, "sendMessages: message processed successfully",
			"conv_id", msg.ConvID,
			"server_msg_id", msg.ServerMsgID,
			"client_msg_id", msg.ClientMsgID,
			"send_id", msg.SendID,
			"seq", msg.Seq,
			"msg_index", index)

	}

	resp = &sdkws.SendMessageResp{
		Infos: respInfos,
	}

	return resp, nil
}

func (s *MessageServiceImpl) createConversationIfNeed(ctx context.Context, msg *sdkws.MsgData) error {
	conv, _ := s.MsgDatabase.GetConversationInfo(ctx, msg.ConvID)

	if conv == nil {
		klog.CtxInfof(ctx, "createConversationIfNeed: creating new conversation",
			"conv_id", msg.ConvID,
			"owner_user_id", msg.SendID,
			"client_msg_id", msg.ClientMsgID,
			"server_msg_id", msg.ServerMsgID)

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
