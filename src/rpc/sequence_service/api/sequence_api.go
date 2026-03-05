package api

import (
	"context"

	"github.com/cloudwego/kitex/pkg/klog"
	sequencepb "github.com/rhp-QE/roc-im-server/kitex_gen/sequence"
	"github.com/rhp-QE/roc-im-server/src/rpc/sequence_service/service"
)

// SequenceAPI 实现 SequenceService 接口（接口层）
type SequenceAPI struct {
	service service.SequenceService
}

// NewSequenceAPI 创建 SequenceAPI 实例
func NewSequenceAPI(seqService service.SequenceService) *SequenceAPI {
	return &SequenceAPI{
		service: seqService,
	}
}

// GetNextSeqInc 获取下一个序列号（简单递增）
func (h *SequenceAPI) GetNextSeqInc(ctx context.Context, req *sequencepb.GetNextSeqIncRequest) (resp *sequencepb.GetNextSeqResponse, err error) {
	resp = &sequencepb.GetNextSeqResponse{}

	if req == nil || req.Id == "" {
		resp.ErrorCode = "INVALID_REQUEST"
		resp.ErrorMsg = "id is required"
		return resp, nil
	}

	seq, err := h.service.GetNextSeqInc(ctx, req.Id)
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
func (h *SequenceAPI) GetNextSeqConsecutive(ctx context.Context, req *sequencepb.GetNextSeqConsecutiveRequest) (resp *sequencepb.GetNextSeqResponse, err error) {
	resp = &sequencepb.GetNextSeqResponse{}

	if req == nil || req.Id == "" {
		resp.ErrorCode = "INVALID_REQUEST"
		resp.ErrorMsg = "id is required"
		return resp, nil
	}

	seq, err := h.service.GetNextSeqConsecutive(ctx, req.Id)
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
var _ sequencepb.SequenceService = (*SequenceAPI)(nil)
