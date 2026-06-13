package rpc

import (
	"context"

	"github.com/cloudwego/kitex/pkg/klog"
	back "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/back"
	"github.com/rhp-QE/roc-im-server/src/rpc/backservice/application"
	deps "github.com/rhp-QE/roc-im-server/src/rpc/backservice/infrastructure/deps"
)

// BackHandler 实现 backservice Kitex 接口。
// 这一层只做 RPC 参数兜底、日志和错误转换，不承载消息路由规则。
type BackHandler struct {
	usecase application.BackUsecase
}

// NewBackHandler 创建 Kitex handler。
func NewBackHandler(serviceCtx deps.ServiceContext) back.BackService {
	return &BackHandler{
		usecase: application.NewBackUsecase(serviceCtx),
	}
}

// Call 处理来自长链网关的调用请求。
func (h *BackHandler) Call(ctx context.Context, req *back.CallRequest) (resp *back.CallResponse, err error) {
	if req == nil {
		klog.CtxErrorf(ctx, "[BackHandler] Call request is nil")
		return nil, err
	}

	klog.CtxInfof(ctx, "[BackHandler] Call",
		"request_id", req.GetRequestID(),
		"service", req.GetService(),
		"method", req.GetMethod(),
		"user_id", req.GetUserID())

	return h.usecase.Call(ctx, req)
}

var _ back.BackService = (*BackHandler)(nil)
