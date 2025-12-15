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

// GetNextSeq 获取下一个序列号
func (h *SequenceAPI) GetNextSeq(ctx context.Context, req *sequencepb.GetNextSeqRequest) (resp *sequencepb.GetNextSeqResponse, err error) {
	resp = &sequencepb.GetNextSeqResponse{}

	if req == nil || req.ConversationId == "" {
		resp.ErrorCode = "INVALID_REQUEST"
		resp.ErrorMsg = "conversation_id is required"
		return resp, nil
	}

	seq, err := h.service.GetNextSeq(ctx, req.ConversationId)
	if err != nil {
		klog.CtxErrorf(ctx, "GetNextSeq failed",
			"conversation_id", req.ConversationId,
			"error", err.Error())
		resp.ErrorCode = "INTERNAL_ERROR"
		resp.ErrorMsg = err.Error()
		return resp, nil
	}

	resp.Seq = seq
	klog.CtxDebugf(ctx, "GetNextSeq success",
		"conversation_id", req.ConversationId,
		"seq", seq)

	return resp, nil
}

// BatchGetNextSeq 批量获取序列号
func (h *SequenceAPI) BatchGetNextSeq(ctx context.Context, req *sequencepb.BatchGetNextSeqRequest) (resp *sequencepb.BatchGetNextSeqResponse, err error) {
	resp = &sequencepb.BatchGetNextSeqResponse{
		Results: make([]*sequencepb.SeqResult, 0),
	}

	if req == nil || len(req.ConversationIds) == 0 {
		return resp, nil
	}

	results, err := h.service.BatchGetNextSeq(ctx, req.ConversationIds)
	if err != nil {
		klog.CtxErrorf(ctx, "BatchGetNextSeq failed",
			"error", err.Error())
		// 即使有错误，也返回部分结果
	}

	// 构建响应
	for _, convID := range req.ConversationIds {
		result := &sequencepb.SeqResult{
			ConversationId: convID,
		}

		if seq, ok := results[convID]; ok {
			result.Seq = seq
		} else {
			result.ErrorCode = "FAILED"
			result.ErrorMsg = "failed to get seq"
		}

		resp.Results = append(resp.Results, result)
	}

	klog.CtxDebugf(ctx, "BatchGetNextSeq success",
		"count", len(resp.Results))

	return resp, nil
}

// GetMaxSeq 获取当前最大序列号
func (h *SequenceAPI) GetMaxSeq(ctx context.Context, req *sequencepb.GetMaxSeqRequest) (resp *sequencepb.GetMaxSeqResponse, err error) {
	resp = &sequencepb.GetMaxSeqResponse{}

	if req == nil || req.ConversationId == "" {
		resp.ErrorCode = "INVALID_REQUEST"
		resp.ErrorMsg = "conversation_id is required"
		return resp, nil
	}

	maxSeq, err := h.service.GetMaxSeq(ctx, req.ConversationId)
	if err != nil {
		klog.CtxErrorf(ctx, "GetMaxSeq failed",
			"conversation_id", req.ConversationId,
			"error", err.Error())
		resp.ErrorCode = "INTERNAL_ERROR"
		resp.ErrorMsg = err.Error()
		return resp, nil
	}

	resp.MaxSeq = maxSeq
	klog.CtxDebugf(ctx, "GetMaxSeq success",
		"conversation_id", req.ConversationId,
		"max_seq", maxSeq)

	return resp, nil
}

// 编译期检查，确保实现了 sequencepb.SequenceService 接口
var _ sequencepb.SequenceService = (*SequenceAPI)(nil)
