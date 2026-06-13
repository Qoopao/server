package rpc

import (
	"context"

	"github.com/cloudwego/kitex/pkg/klog"
	messagepb "github.com/rhp-QE/roc-im-server/kitex_gen/message"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/rpc/message_service/application"
)

// MessageHandler 实现 MessageService Kitex 接口。
// 该层只做 RPC 协议适配、入参兜底和错误转换，不承载消息一致性规则。
type MessageHandler struct {
	svc application.MessageService
}

// NewMessageHandler 创建 MessageHandler 实例
func NewMessageHandler(svc application.MessageService) *MessageHandler {
	return &MessageHandler{
		svc: svc,
	}
}

// BatchSendMessage 批量发送消息
func (h *MessageHandler) BatchSendMessage(ctx context.Context, req *sdkws.BatchSendMessageRequest) (resp *sdkws.BatchSendMessageResponse, err error) {
	if req == nil || len(req.Msgs) == 0 {
		return &sdkws.BatchSendMessageResponse{
			Results: []*sdkws.SendMessageResult{},
		}, nil
	}

	resp, err = h.svc.BatchSendMessage(ctx, req)
	if err != nil {
		klog.CtxErrorf(ctx, "BatchSendMessage failed", "error", err.Error())
	}
	return resp, err
}

// FetchConvMessageList 查询会话消息列表
func (h *MessageHandler) FetchConvMessageList(ctx context.Context, req *sdkws.FetchConvMessageListRequest) (resp *sdkws.FetchConvMessageListResponse, err error) {
	// 参数校验
	if req == nil {
		return &sdkws.FetchConvMessageListResponse{
			Messages: []*sdkws.MessageData{},
			HaveMore: false,
			Error:    "request is nil",
		}, nil
	}

	convID := req.GetConvID()
	if convID == "" {
		return &sdkws.FetchConvMessageListResponse{
			Messages: []*sdkws.MessageData{},
			HaveMore: false,
			Error:    "conv_id is empty",
		}, nil
	}

	// 调用 application 层
	resp, err = h.svc.FetchConvMessageList(ctx, req)
	if err != nil {
		klog.CtxErrorf(ctx, "FetchConvMessageList failed", "error", err.Error())
		return &sdkws.FetchConvMessageListResponse{
			Messages: []*sdkws.MessageData{},
			HaveMore: false,
			Error:    err.Error(),
		}, nil
	}
	return resp, nil
}

// BatchGetMessages 根据消息ID列表批量获取消息详情
func (h *MessageHandler) BatchGetMessages(ctx context.Context, req *sdkws.BatchGetMessagesRequest) (resp *sdkws.BatchGetMessagesResponse, err error) {
	// 参数校验
	if req == nil {
		return &sdkws.BatchGetMessagesResponse{
			Results: []*sdkws.GetMessageResult{},
		}, nil
	}

	messageIDs := req.GetMessageIDs()
	if len(messageIDs) == 0 {
		return &sdkws.BatchGetMessagesResponse{
			Results: []*sdkws.GetMessageResult{},
		}, nil
	}

	// 调用 application 层
	resp, err = h.svc.BatchGetMessages(ctx, req)
	if err != nil {
		klog.CtxErrorf(ctx, "BatchGetMessages failed", "error", err.Error())
		return &sdkws.BatchGetMessagesResponse{
			Results: []*sdkws.GetMessageResult{},
		}, err
	}

	return resp, nil
}

// 编译期检查，确保实现了 messagepb.MessageService 接口
var _ messagepb.MessageService = (*MessageHandler)(nil)
