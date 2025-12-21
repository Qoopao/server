package api

import (
	"context"

	"github.com/cloudwego/kitex/pkg/klog"
	back "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/back"
	"github.com/rhp-QE/roc-im-server/src/rpc/backservice/service"
	servicecontext "github.com/rhp-QE/roc-im-server/src/rpc/backservice/servicecontext"
)

// BackServiceAPI BackService API 层接口
type BackServiceAPI interface {
	// Call 处理来自网关的调用请求
	Call(ctx context.Context, req *back.CallRequest) (resp *back.CallResponse, err error)
}

type backServiceAPIImpl struct {
	service service.BackService
}

// NewBackServiceAPI 创建 BackServiceAPI 实例
func NewBackServiceAPI(service service.BackService) BackServiceAPI {
	return &backServiceAPIImpl{
		service: service,
	}
}

// Call 处理来自网关的调用请求（RPC 映射层）
func (a *backServiceAPIImpl) Call(ctx context.Context, req *back.CallRequest) (resp *back.CallResponse, err error) {
	// 参数校验
	if req == nil {
		klog.CtxErrorf(ctx, "[BackServiceAPI] Call request is nil")
		return nil, err
	}

	// 记录请求日志
	klog.CtxInfof(ctx, "[BackServiceAPI] Call - requestID: %s, service: %s, method: %s, userID: %s",
		req.GetRequestID(), req.GetService(), req.GetMethod(), req.GetUserID())

	// 调用 service 层处理业务逻辑
	return a.service.Call(ctx, req)
}

// BackServiceImpl 实现 kitex 生成的 BackService 接口
type BackServiceImpl struct {
	api BackServiceAPI
}

// NewBackServiceImpl 创建 BackServiceImpl 实例（用于 kitex server）
func NewBackServiceImpl(serviceCtx servicecontext.ServiceContext) back.BackService {
	// 组装各层：service → api
	backService := service.NewBackService(serviceCtx)
	backAPI := NewBackServiceAPI(backService)

	return &BackServiceImpl{
		api: backAPI,
	}
}

// Call 处理来自网关的调用请求（RPC 映射层）
func (s *BackServiceImpl) Call(ctx context.Context, req *back.CallRequest) (resp *back.CallResponse, err error) {
	return s.api.Call(ctx, req)
}
