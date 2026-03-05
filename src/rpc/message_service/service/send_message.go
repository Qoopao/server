package service

import (
	"context"
	"fmt"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/google/uuid"
	"github.com/openimsdk/tools/errs"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	sequencepb "github.com/rhp-QE/roc-im-server/kitex_gen/sequence"
)

// BatchSendMessage 实现批量发送消息的核心流程：
// 1. 为每条消息生成 seq（调用 sequence-service）
// 2. 将成功分配 seq 的消息批量投递到 MQ
// 3. 返回逐条结果（是否成功、失败原因）
func (s *messageServiceImpl) BatchSendMessage(ctx context.Context, req *sdkws.BatchSendMessageRequest) (*sdkws.BatchSendMessageResponse, error) {
	if req == nil || len(req.Msgs) == 0 {
		return nil, errs.ErrArgs.WrapMsg("msgs is empty")
	}

	results := make([]*sdkws.SendMessageResult, len(req.Msgs))

	// 获取 sequence-service 客户端
	seqClient, err := s.serviceCtx.GetSequenceServiceClient(ctx)
	if err != nil {
		klog.CtxErrorf(ctx, "BatchSendMessage: failed to get sequence client", "error", err.Error())
		return nil, errs.ErrInternalServer.WrapMsg(fmt.Sprintf("failed to get sequence service client: %v", err))
	}

	// 为每条消息生成 seq
	successMsgs := make([]*sdkws.MessageData, 0, len(req.Msgs))
	for idx, msg := range req.Msgs {
		result := &sdkws.SendMessageResult{
			ErrorCode: "",
			ErrorMsg:  "",
			Msg:       msg,
		}
		results[idx] = result

		if msg == nil {
			result.ErrorCode = "INVALID_MESSAGE"
			result.ErrorMsg = "message is nil"
			continue
		}

		if msg.ConvID == "" {
			result.ErrorCode = "INVALID_CONVERSATION"
			result.ErrorMsg = "convID is required"
			continue
		}

		// 调用 sequence-service 获取下一个 消息序列号（连续递增）
		seqResp, err := seqClient.GetNextSeqConsecutive(ctx, &sequencepb.GetNextSeqConsecutiveRequest{
			Id: msg.ConvID,
		})
		if err != nil {
			klog.CtxErrorf(ctx, "BatchSendMessage: GetNextSeq failed",
				"conv_id", msg.ConvID,
				"error", err.Error())
			result.ErrorCode = "SEQUENCE_SERVICE_ERROR"
			result.ErrorMsg = err.Error()
			continue
		}

		if seqResp.ErrorCode != "" {
			klog.CtxErrorf(ctx, "BatchSendMessage: GetNextSeq response error",
				"conv_id", msg.ConvID,
				"error_code", seqResp.ErrorCode,
				"error_msg", seqResp.ErrorMsg)
			result.ErrorCode = seqResp.ErrorCode
			result.ErrorMsg = seqResp.ErrorMsg
			continue
		}

		// 设置消息 seq
		msg.Seq = seqResp.Seq

		// 生成消息ID
		if msg.SMessageID == "" {
			msg.SMessageID = uuid.New().String()
		}

		successMsgs = append(successMsgs, msg)

		klog.CtxDebugf(ctx, "BatchSendMessage: seq and msgID assigned",
			"conv_id", msg.ConvID,
			"seq", msg.Seq,
			"msg_id", msg.SMessageID)
	}

	// 将成功分配 seq 的消息批量投递到 MQ
	if len(successMsgs) > 0 {
		if err := s.storage.PublishMessages(ctx, successMsgs); err != nil {
			klog.CtxErrorf(ctx, "BatchSendMessage: publish to mq failed", "error", err.Error())
			// 将发布失败标记到所有成功分配 seq 的消息上
			for _, res := range results {
				if res != nil && res.ErrorCode == "" {
					res.ErrorCode = "MQ_PUBLISH_FAILED"
					res.ErrorMsg = err.Error()
				}
			}
		}
	}

	return &sdkws.BatchSendMessageResponse{
		Results: results,
	}, nil
}
