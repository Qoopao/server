package service

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	back "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/back"
	"github.com/rhp-QE/roc-foundation-util-go/stringutil"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	consts "github.com/rhp-QE/roc-im-server/src/rpc/const"
)

// Call 处理来自网关的调用请求
// 根据 req.Service 和 req.Method 路由到对应的业务服务
func (s *backServiceImpl) Call(ctx context.Context, req *back.CallRequest) (resp *back.CallResponse, err error) {
	// 创建响应
	resp = &back.CallResponse{
		RequestID: req.GetRequestID(),
		Type:      req.GetType(),
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

	// 根据 service 名称路由到对应的服务
	var responsePayload []byte
	var callErr error

	switch serviceName {
	case consts.MessageServiceName:
		responsePayload, callErr = s.callMessageService(ctx, methodName, payload)
	case consts.ConversationServiceName:
		responsePayload, callErr = s.callConversationService(ctx, methodName, payload)
	default:
		resp.Error = fmt.Sprintf("unknown service: %s", serviceName)
		klog.CtxErrorf(ctx, "[BackService] unknown service: %s", serviceName)
		return resp, nil
	}

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

// callMessageService 调用消息服务
func (s *backServiceImpl) callMessageService(ctx context.Context, methodName string, payload []byte) ([]byte, error) {
	client, err := s.serviceCtx.GetMessageServiceClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get message service client: %w", err)
	}

	switch methodName {
	case MethodBatchSendMessage:
		req := &sdkws.BatchSendMessageRequest{}
		if err := req.Unmarshal(payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal BatchSendMessageRequest: %w", err)
		}
		resp, err := client.BatchSendMessage(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.Marshal(nil)

	case MethodBatchChangeMessages:
		req := &sdkws.BatchChangeMessagesRequest{}
		if err := req.Unmarshal(payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal BatchChangeMessagesRequest: %w", err)
		}
		resp, err := client.BatchChangeMessages(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.Marshal(nil)

	case MethodFetchConvMessageList:
		req := &sdkws.FetchConvMessageListRequest{}
		if err := req.Unmarshal(payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal FetchConvMessageListRequest: %w", err)
		}
		resp, err := client.FetchConvMessageList(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.Marshal(nil)

	case MethodBatchGetMessages:
		req := &sdkws.BatchGetMessagesRequest{}
		if err := req.Unmarshal(payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal BatchGetMessagesRequest: %w", err)
		}
		resp, err := client.BatchGetMessages(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.Marshal(nil)

	default:
		return nil, fmt.Errorf("unknown method: %s for message-service", methodName)
	}
}

// callConversationService 调用会话服务
func (s *backServiceImpl) callConversationService(ctx context.Context, methodName string, payload []byte) ([]byte, error) {
	client, err := s.serviceCtx.GetConversationServiceClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation service client: %w", err)
	}

	switch methodName {
	case MethodBatchChangeConversations:
		req := &sdkws.BatchChangeConversationsRequest{}
		if err := req.Unmarshal(payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal BatchChangeConversationsRequest: %w", err)
		}
		resp, err := client.BatchChangeConversations(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.Marshal(nil)

	case MethodFetchUserRecentConvList:
		req := &sdkws.FetchUserRecentConvListRequest{}
		if err := req.Unmarshal(payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal FetchUserRecentConvListRequest: %w", err)
		}
		resp, err := client.FetchUserRecentConvList(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.Marshal(nil)

	case MethodUserMessageIntegrityCheck:
		req := &sdkws.UserMessageIntegrityCheckRequest{}
		if err := req.Unmarshal(payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal UserMessageIntegrityCheckRequest: %w", err)
		}
		resp, err := client.UserMessageIntegrityCheck(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.Marshal(nil)

	case MethodBatchGetConversations:
		req := &sdkws.BatchGetConversationsRequest{}
		if err := req.Unmarshal(payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal BatchGetConversationsRequest: %w", err)
		}
		resp, err := client.BatchGetConversations(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.Marshal(nil)

	default:
		return nil, fmt.Errorf("unknown method: %s for conversation-service", methodName)
	}
}
