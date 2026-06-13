package application

import (
	"context"

	back "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/back"
	deps "github.com/rhp-QE/roc-im-server/src/rpc/backservice/infrastructure/deps"
)

// BackUsecase 定义网关 RPC 路由应用用例。
type BackUsecase interface {
	// Call 处理来自网关的调用请求，根据 req.Service 和 req.Method 路由到对应的业务服务
	Call(ctx context.Context, req *back.CallRequest) (resp *back.CallResponse, err error)
}

// backUsecase 是 BackUsecase 的具体实现
type backUsecase struct {
	serviceCtx deps.ServiceContext
}

// NewBackUsecase 创建 BackUsecase 实例
func NewBackUsecase(serviceCtx deps.ServiceContext) BackUsecase {
	return &backUsecase{
		serviceCtx: serviceCtx,
	}
}
