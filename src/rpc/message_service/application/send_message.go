package application

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
// 1. 使用 sender + client_msg_id 做发送重试幂等检查
// 2. 为新消息生成 seq（调用 sequence-service）和 server_msg_id
// 3. 先写入 message fact，保证返回前消息事实可查
// 4. 将已持久化消息投递到 MQ，驱动会话投影、混链和在线推送
// 5. 返回逐条结果（是否成功、失败原因）
func (s *messageServiceImpl) BatchSendMessage(ctx context.Context, req *sdkws.BatchSendMessageRequest) (*sdkws.BatchSendMessageResponse, error) {
	if req == nil || len(req.Msgs) == 0 {
		return nil, errs.ErrArgs.WrapMsg("msgs is empty")
	}

	results := make([]*sdkws.SendMessageResult, len(req.Msgs))
	resultByServerMsgID := make(map[string]*sdkws.SendMessageResult, len(req.Msgs))

	// 已经写入 message fact 的消息才允许进入 MQ。MQ 当前承担 outbox 的角色；
	// 后续引入事务 outbox 后，可把这里的 PublishMessages 替换为 outbox relay。
	persistedMsgs := make([]*sdkws.MessageData, 0, len(req.Msgs))
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

		if existing, found, err := s.messageRepo.FindMessageByClientMsgID(ctx, msg.SendID, msg.CMessaegID); err != nil {
			result.ErrorCode = "MESSAGE_IDEMPOTENCY_CHECK_FAILED"
			result.ErrorMsg = err.Error()
			continue
		} else if found {
			// 客户端超时重试时必须收敛到首次提交的 server_msg_id 和 conv_seq。
			// 在正式 outbox 状态落地前，重试会重新发布已存在消息以修复可能的 MQ 发布失败；
			// Consumer 和 SDK 必须按 server_msg_id 幂等处理重复事件。
			result.Msg = existing
			persistedMsgs = append(persistedMsgs, existing)
			resultByServerMsgID[existing.SMessageID] = result
			klog.CtxInfof(ctx, "BatchSendMessage: idempotent message reused",
				"send_id", existing.SendID,
				"client_msg_id", existing.CMessaegID,
				"server_msg_id", existing.SMessageID,
				"conv_id", existing.ConvID,
				"seq", existing.Seq)
			continue
		}

		seqClient, err := s.serviceCtx.GetSequenceServiceClient(ctx)
		if err != nil {
			klog.CtxErrorf(ctx, "BatchSendMessage: failed to get sequence client",
				"conv_id", msg.ConvID,
				"client_msg_id", msg.CMessaegID,
				"error", err.Error())
			result.ErrorCode = "SEQUENCE_SERVICE_CLIENT_ERROR"
			result.ErrorMsg = fmt.Sprintf("failed to get sequence service client: %v", err)
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

		if seqResp == nil {
			result.ErrorCode = "SEQUENCE_SERVICE_EMPTY_RESPONSE"
			result.ErrorMsg = "sequence-service returned empty response"
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

		// 生成消息ID。sender + client_msg_id 存在时使用确定性 UUID，
		// 使客户端超时并发重试最终落到同一条 message fact。
		if msg.SMessageID == "" {
			msg.SMessageID = buildServerMessageID(msg)
		}

		persisted, _, err := s.messageRepo.SaveMessageFact(ctx, msg)
		if err != nil {
			result.ErrorCode = "MESSAGE_FACT_SAVE_FAILED"
			result.ErrorMsg = err.Error()
			continue
		}

		result.Msg = persisted
		persistedMsgs = append(persistedMsgs, persisted)
		resultByServerMsgID[persisted.SMessageID] = result

		klog.CtxDebugf(ctx, "BatchSendMessage: seq and msgID assigned",
			"conv_id", msg.ConvID,
			"seq", msg.Seq,
			"msg_id", msg.SMessageID)
	}

	// 将已持久化消息批量投递到 MQ
	if len(persistedMsgs) > 0 {
		if err := s.messageRepo.PublishMessages(ctx, persistedMsgs); err != nil {
			klog.CtxErrorf(ctx, "BatchSendMessage: publish to mq failed", "error", err.Error())
			// MQ 仍处于发送成功边界内；发布失败不能返回 success，避免客户端误判完整提交。
			for _, msg := range persistedMsgs {
				if msg == nil {
					continue
				}
				if res := resultByServerMsgID[msg.SMessageID]; res != nil && res.ErrorCode == "" {
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

func buildServerMessageID(msg *sdkws.MessageData) string {
	if msg == nil {
		return uuid.New().String()
	}
	if msg.SendID != "" && msg.CMessaegID != "" {
		return uuid.NewSHA1(uuid.NameSpaceOID, []byte(msg.SendID+":"+msg.CMessaegID)).String()
	}
	return uuid.New().String()
}
