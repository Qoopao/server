package service

import (
	"context"

	back "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/back"
	servicecontext "github.com/rhp-QE/roc-im-server/src/rpc/backservice/servicecontext"
)

// BackService 定义 BackService 的业务接口（逻辑层）
type BackService interface {
	// Call 处理来自网关的调用请求，根据 req.Service 和 req.Method 路由到对应的业务服务
	Call(ctx context.Context, req *back.CallRequest) (resp *back.CallResponse, err error)
}

// backServiceImpl 是 BackService 的具体实现
type backServiceImpl struct {
	serviceCtx servicecontext.ServiceContext
}

// NewBackService 创建 BackService 实例
func NewBackService(serviceCtx servicecontext.ServiceContext) BackService {
	return &backServiceImpl{
		serviceCtx: serviceCtx,
	}
}
