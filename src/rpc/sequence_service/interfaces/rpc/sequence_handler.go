package rpc

import (
	"context"

	"github.com/cloudwego/kitex/pkg/klog"
	sequencepb "github.com/rhp-QE/roc-im-server/kitex_gen/sequence"
	"github.com/rhp-QE/roc-im-server/src/rpc/sequence_service/application"
)

// SequenceHandler 实现 SequenceService Kitex 接口。
type SequenceHandler struct {
	usecase application.SequenceUsecase
}

// NewSequenceHandler 创建 SequenceHandler 实例
func NewSequenceHandler(seqUsecase application.SequenceUsecase) *SequenceHandler {
	return &SequenceHandler{
		usecase: seqUsecase,
	}
}

// GetNextSeqInc 获取下一个序列号（简单递增）
func (h *SequenceHandler) GetNextSeqInc(ctx context.Context, req *sequencepb.GetNextSeqIncRequest) (resp *sequencepb.GetNextSeqResponse, err error) {
	resp = &sequencepb.GetNextSeqResponse{}

	if req == nil || req.Id == "" {
		resp.ErrorCode = "INVALID_REQUEST"
		resp.ErrorMsg = "id is required"
		return resp, nil
	}

	seq, err := h.usecase.GetNextSeqInc(ctx, req.Id)
	if err != nil {
		klog.CtxErrorf(ctx, "GetNextSeqInc failed",
			"id", req.Id,
			"error", err.Error())
		resp.ErrorCode = "INTERNAL_ERROR"
		resp.ErrorMsg = err.Error()
		return resp, nil
	}

	resp.Seq = seq
	klog.CtxDebugf(ctx, "GetNextSeqInc success",
		"id", req.Id,
		"seq", seq)

	return resp, nil
}

// GetNextSeqConsecutive 获取下一个序列号（连续递增）
func (h *SequenceHandler) GetNextSeqConsecutive(ctx context.Context, req *sequencepb.GetNextSeqConsecutiveRequest) (resp *sequencepb.GetNextSeqResponse, err error) {
	resp = &sequencepb.GetNextSeqResponse{}

	if req == nil || req.Id == "" {
		resp.ErrorCode = "INVALID_REQUEST"
		resp.ErrorMsg = "id is required"
		return resp, nil
	}

	seq, err := h.usecase.GetNextSeqConsecutive(ctx, req.Id)
	if err != nil {
		klog.CtxErrorf(ctx, "GetNextSeqConsecutive failed",
			"id", req.Id,
			"error", err.Error())
		resp.ErrorCode = "INTERNAL_ERROR"
		resp.ErrorMsg = err.Error()
		return resp, nil
	}

	resp.Seq = seq
	klog.CtxDebugf(ctx, "GetNextSeqConsecutive success",
		"id", req.Id,
		"seq", seq)

	return resp, nil
}

// 编译期检查，确保实现了 sequencepb.SequenceService 接口
var _ sequencepb.SequenceService = (*SequenceHandler)(nil)
