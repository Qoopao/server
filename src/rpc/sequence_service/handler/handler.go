package handler

import (
	"context"

	"github.com/cloudwego/kitex/pkg/klog"
	sequence "github.com/rhp-QE/roc-im-server/kitex_gen/sequence"
	"github.com/rhp-QE/roc-im-server/src/rpc/sequence_service/service"
)

// SequenceServiceHandler 实现SequenceService接口（接口层）
type SequenceServiceHandler struct {
	service service.SequenceService
}

// NewSequenceServiceHandler 创建SequenceServiceHandler实例
func NewSequenceServiceHandler(seqService service.SequenceService) *SequenceServiceHandler {
	return &SequenceServiceHandler{
		service: seqService,
	}
}

// GetNextSeq 获取下一个序列号
func (h *SequenceServiceHandler) GetNextSeq(ctx context.Context, req *sequence.GetNextSeqRequest) (resp *sequence.GetNextSeqResponse, err error) {
	resp = &sequence.GetNextSeqResponse{}

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
func (h *SequenceServiceHandler) BatchGetNextSeq(ctx context.Context, req *sequence.BatchGetNextSeqRequest) (resp *sequence.BatchGetNextSeqResponse, err error) {
	resp = &sequence.BatchGetNextSeqResponse{
		Results: make([]*sequence.SeqResult, 0),
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
		result := &sequence.SeqResult{
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
func (h *SequenceServiceHandler) GetMaxSeq(ctx context.Context, req *sequence.GetMaxSeqRequest) (resp *sequence.GetMaxSeqResponse, err error) {
	resp = &sequence.GetMaxSeqResponse{}

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
