package backservice

import (
	"context"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen"
)

// BackServiceImpl 实现 BackService 接口
// BackService 用于接收来自网关的调用请求，并路由到对应的业务服务
type BackServiceImpl struct {
	// TODO: 添加需要的依赖，例如数据库、缓存等
}

// NewBackServiceImpl 创建新的 BackServiceImpl
func NewBackServiceImpl() *BackServiceImpl {
	return &BackServiceImpl{}
}

// Call 处理来自网关的调用请求
// 根据 req.Service 和 req.Method 路由到对应的业务服务
func (s *BackServiceImpl) Call(ctx context.Context, req *kitex_gen.CallRequest) (resp *kitex_gen.CallResponse, err error) {
	// 记录请求日志
	klog.Infof("BackService.Call - requestID: %s, service: %s, method: %s, userID: %s",
		req.GetRequestID(), req.GetService(), req.GetMethod(), req.GetUserID())

	// 创建响应
	resp = &kitex_gen.CallResponse{
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
	klog.Warnf("BackService.Call not implemented - service: %s, method: %s", req.GetService(), req.GetMethod())

	return resp, nil
}
