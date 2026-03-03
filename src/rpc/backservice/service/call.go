package service

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	back "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/back"
	"github.com/rhp-QE/roc-foundation-util-go/stringutil"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	consts "github.com/rhp-QE/roc-im-server/src/const"
)

// Call 处理来自网关的调用请求
// 作为网关层，根据客户端的 method（SDKWSMethod）自动路由到对应的后端服务
func (s *backServiceImpl) Call(ctx context.Context, req *back.CallRequest) (resp *back.CallResponse, err error) {
	// 创建响应
	resp = &back.CallResponse{
		RequestID: req.GetRequestID(),
		Type:      "response",
		Success:   false,
		Timestamp: time.Now().Unix(),
	}

	// 参数校验
	serviceName := req.GetService()
	methodName := req.GetMethod()
	payload := req.GetPayload()

	if stringutil.IsEmpty(serviceName) {
		resp.Error = "service name is empty"
		klog.CtxErrorf(ctx, "[BackService] service name is empty")
		return resp, nil
	}

	if stringutil.IsEmpty(methodName) {
		resp.Error = "method name is empty"
		klog.CtxErrorf(ctx, "[BackService] method name is empty")
		return resp, nil
	}

	// 网关层：根据客户端的 method 自动路由到对应的后端服务
	// 将 SDKWSMethod 枚举映射到对应的后端服务
	var responsePayload []byte
	var callErr error
	responsePayload, callErr = s.callBackserviceIM(ctx, methodName, payload)

	// 处理调用结果
	if callErr != nil {
		resp.Error = callErr.Error()
		klog.CtxErrorf(ctx, "[BackService] call service failed - service: %s, method: %s, error: %v",
			serviceName, methodName, callErr)
		return resp, nil
	}

	resp.Success = true
	resp.Payload = responsePayload

	klog.CtxDebugf(ctx, "[BackService] call service success - service: %s, method: %s",
		serviceName, methodName)

	return resp, nil
}

// callBackserviceIM 网关层路由：将客户端的 SDKWSMethod 映射到对应的后端服务
func (s *backServiceImpl) callBackserviceIM(ctx context.Context, methodName string, payload []byte) ([]byte, error) {
	switch methodName {
	case consts.SDKWSMethodSendMessage:
		// 101: 发送消息 -> message-service.BatchSendMessage
		client, err := s.serviceCtx.GetMessageServiceClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get message service client: %w", err)
		}
		req := &sdkws.BatchSendMessageRequest{}
		if err := req.Unmarshal(payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal BatchSendMessageRequest: %w", err)
		}
		resp, err := client.BatchSendMessage(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.Marshal(nil)

	case consts.SDKWSMethodPullSingleList:
		// 102: 拉取单链 -> message-service.FetchConvMessageList
		client, err := s.serviceCtx.GetMessageServiceClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get message service client: %w", err)
		}
		req := &sdkws.FetchConvMessageListRequest{}
		if err := req.Unmarshal(payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal FetchConvMessageListRequest: %w", err)
		}
		resp, err := client.FetchConvMessageList(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.Marshal(nil)

	case consts.SDKWSMethodPullMixList:
		// 103: 拉取混链 -> conversation-service.FetchUserRecentConvList
		client, err := s.serviceCtx.GetConversationServiceClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get conversation service client: %w", err)
		}
		req := &sdkws.FetchUserRecentConvListRequest{}
		if err := req.Unmarshal(payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal FetchUserRecentConvListRequest: %w", err)
		}
		resp, err := client.FetchUserRecentConvList(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.Marshal(nil)

	case consts.SDKWSMethodPushUserMessage:
		// 104: 下推用户消息（推送消息，不需要后端服务调用）
		// 返回空响应，推送由网关层处理
		return nil, fmt.Errorf("PUSH_USER_MESSAGE is a push operation, should be handled by gateway")

	case consts.SDKWSMethodPushCmdMessage:
		// 105: 下推命令消息（推送消息，不需要后端服务调用）
		// 返回空响应，推送由网关层处理
		return nil, fmt.Errorf("PUSH_CMD_MESSAGE is a push operation, should be handled by gateway")

	case consts.SDKWSMethodUserMessageIntegrityCheck:
		// 106: 混链拉取会话完整性校验 -> conversation-service.UserMessageIntegrityCheck
		client, err := s.serviceCtx.GetConversationServiceClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get conversation service client: %w", err)
		}
		req := &sdkws.UserMessageIntegrityCheckRequest{}
		if err := req.Unmarshal(payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal UserMessageIntegrityCheckRequest: %w", err)
		}
		resp, err := client.UserMessageIntegrityCheck(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.Marshal(nil)

	case consts.SDKWSMethodMessageChange:
		// 107: 消息改变 -> message-service.BatchChangeMessages
		client, err := s.serviceCtx.GetMessageServiceClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get message service client: %w", err)
		}
		req := &sdkws.BatchChangeMessagesRequest{}
		if err := req.Unmarshal(payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal BatchChangeMessagesRequest: %w", err)
		}
		resp, err := client.BatchChangeMessages(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.Marshal(nil)

	case consts.SDKWSMethodConversationChange:
		// 108: 会话改变 -> conversation-service.BatchChangeConversations
		client, err := s.serviceCtx.GetConversationServiceClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get conversation service client: %w", err)
		}
		req := &sdkws.BatchChangeConversationsRequest{}
		if err := req.Unmarshal(payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal BatchChangeConversationsRequest: %w", err)
		}
		resp, err := client.BatchChangeConversations(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.Marshal(nil)

	default:
		return nil, fmt.Errorf("unsupported SDKWSMethod: %s", methodName)
	}
}
