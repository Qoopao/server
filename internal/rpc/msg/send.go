//
// Author : Ruanhuipeng
// Date   : 06/08/2025

package msg

import (
	"context"
	"fmt"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/google/uuid"
	"github.com/openimsdk/tools/errs"
	"github.com/roc/roc-im-server/internal/kitex_gen/conversation"
	"github.com/roc/roc-im-server/internal/kitex_gen/sdkws"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// extractOtelInfo 从context中提取OTEL的trace_id和span_id信息
func extractOtelInfo(ctx context.Context) map[string]string {
	otelInfo := make(map[string]string)

	// 获取当前span
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		spanCtx := span.SpanContext()

		// 提取trace_id
		if spanCtx.HasTraceID() {
			otelInfo["trace_id"] = spanCtx.TraceID().String()
		}

		// 提取span_id
		if spanCtx.HasSpanID() {
			otelInfo["span_id"] = spanCtx.SpanID().String()
		}

		// 提取trace_flags
		otelInfo["trace_flags"] = fmt.Sprintf("%02x", spanCtx.TraceFlags())

		// 提取trace_state (如果有)
		if spanCtx.TraceState().Len() > 0 {
			otelInfo["trace_state"] = spanCtx.TraceState().String()
		}
	}

	return otelInfo
}

// SendMsg implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) sendMessages(ctx context.Context, req *sdkws.SendMessageReq) (resp *sdkws.SendMessageResp, err error) {
	klog.CtxDebugf(ctx, "rhpmark-sendMessages-start")

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
				"clientMsgID", msg.ClientMsgID,
				"sendID", msg.SendID)
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
			klog.CtxErrorf(ctx, "sendMessages: failed to save msg info",
				"convID", msg.ConvID,
				"serverMsgID", msg.ServerMsgID,
				"error", err.Error())
			respInfos[index].ErrorCode = "1"
			respInfos[index].ErrorMsg = err.Error()
			continue
		}

		// 如果没有会话则创建会话
		if err := s.createConversationIfNeed(ctx, msg); err != nil {
			klog.CtxErrorf(ctx, "sendMessages: failed to create conversation",
				"convID", msg.ConvID,
				"sendID", msg.SendID,
				"error", err.Error())
			respInfos[index].ErrorCode = "1"
			respInfos[index].ErrorMsg = err.Error()
			continue
		}

		// 转发消息到MQ
		if err := s.MsgDatabase.MsgToMQ(ctx, msg.SendID, msg.ServerMsgID); err != nil {
			klog.CtxErrorf(ctx, "sendMessages: failed to send msg to MQ",
				"convID", msg.ConvID,
				"serverMsgID", msg.ServerMsgID,
				"sendID", msg.SendID,
				"error", err.Error())
			respInfos[index].ErrorCode = "1"
			respInfos[index].ErrorMsg = err.Error()
			continue
		}

		// 设置返回信息
		respInfos[index].Msg = msg

		klog.CtxDebugf(ctx, "sendMessages: message processed successfully", map[string]interface{}{
			"convID":      msg.ConvID,
			"serverMsgID": msg.ServerMsgID,
			"seq":         msg.Seq,
		})

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
			"convID", msg.ConvID,
			"ownerUserID", msg.SendID)

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
