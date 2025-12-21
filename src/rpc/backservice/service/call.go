package service

import (
	"context"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	back "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/back"
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

	// TODO: 根据 req.Service 和 req.Method 路由到对应的业务服务
	// 这里可以根据业务需求实现具体的路由逻辑
	// 例如：
	// 1. 根据 service 名称路由到对应的服务（msg-service, conversation-service 等）
	// 2. 根据 method 调用对应的方法
	// 3. 处理响应并返回

	// 临时实现：返回错误提示
	resp.Error = "BackService.Call not implemented yet"
	klog.CtxWarnf(ctx, "[BackService] Call not implemented - service: %s, method: %s", req.GetService(), req.GetMethod())

	return resp, nil
}
