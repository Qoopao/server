package api

import (
	"context"

	"github.com/cloudwego/kitex/pkg/klog"
	messagepb "github.com/rhp-QE/roc-im-server/kitex_gen/message"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/rpc/message_service/service"
)

// MessageAPI 实现 MessageService 接口（接口层）
type MessageAPI struct {
	svc service.MessageService
}

// NewMessageAPI 创建 MessageAPI 实例
func NewMessageAPI(svc service.MessageService) *MessageAPI {
	return &MessageAPI{
		svc: svc,
	}
}

// BatchSendMessage 批量发送消息
func (h *MessageAPI) BatchSendMessage(ctx context.Context, req *sdkws.BatchSendMessageRequest) (resp *sdkws.BatchSendMessageResponse, err error) {
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

// BatchChangeMessages 暂未实现
func (h *MessageAPI) BatchChangeMessages(ctx context.Context, req *sdkws.BatchChangeMessagesRequest) (resp *sdkws.BatchChangeMessagesResponse, err error) {
	klog.CtxWarnf(ctx, "BatchChangeMessages not implemented")
	return &sdkws.BatchChangeMessagesResponse{
		Results: []*sdkws.CmdMessageOptResult{},
	}, nil
}

// FetchConvMessageList 暂未实现
func (h *MessageAPI) FetchConvMessageList(ctx context.Context, req *sdkws.FetchConvMessageListRequest) (resp *sdkws.FetchConvMessageListResponse, err error) {
	klog.CtxWarnf(ctx, "FetchConvMessageList not implemented")
	return &sdkws.FetchConvMessageListResponse{}, nil
}

// BatchGetMessages 暂未实现
func (h *MessageAPI) BatchGetMessages(ctx context.Context, req *sdkws.BatchGetMessagesRequest) (resp *sdkws.BatchGetMessagesResponse, err error) {
	klog.CtxWarnf(ctx, "BatchGetMessages not implemented")
	return &sdkws.BatchGetMessagesResponse{
		Results: []*sdkws.GetMessageResult{},
	}, nil
}

// 编译期检查，确保实现了 messagepb.MessageService 接口
var _ messagepb.MessageService = (*MessageAPI)(nil)
