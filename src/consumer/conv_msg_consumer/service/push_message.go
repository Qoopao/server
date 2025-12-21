package service

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	backbon "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/backbon"
	backbonservice "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/backbon/backbonservice"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/common/util"
)

// pushMessage 调用长链服务向用户推送消息
func (s *convMsgConsumerServiceImpl) pushMessage(ctx context.Context, msg *sdkws.MessageData) error {
	// 获取接收者列表
	userIDs := s.getPushTargets(msg)
	if len(userIDs) == 0 {
		klog.CtxDebugf(ctx, "[ConvMsgConsumer] no push targets, skip push",
			"conv_id", msg.ConvID)
		return nil
	}

	// 构造 PushMessage
	pushMsg := s.buildPushMessage(msg)

	// 构造 PushDataReq
	req := &backbon.PushDataReq{
		UserIDs:   userIDs,
		Message:   pushMsg,
		Broadcast: false,
	}

	// 获取 backbon 服务客户端并推送
	client, err := s.getBackbonServiceClient(ctx)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConvMsgConsumer] get backbon service client failed",
			"conv_id", msg.ConvID,
			"error", err.Error())
		// 不返回错误，推送失败不影响消息持久化
		return nil
	}

	resp, err := client.PushData(ctx, req)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConvMsgConsumer] push message failed",
			"conv_id", msg.ConvID,
			"error", err.Error())
		// 不返回错误，推送失败不影响消息持久化
		return nil
	}

	if resp.GetFailCount() > 0 {
		klog.CtxWarnf(ctx, "[ConvMsgConsumer] push message partial failed",
			"conv_id", msg.ConvID,
			"success_count", resp.GetSuccessCount(),
			"fail_count", resp.GetFailCount())
	} else {
		klog.CtxDebugf(ctx, "[ConvMsgConsumer] push message success",
			"conv_id", msg.ConvID,
			"success_count", resp.GetSuccessCount())
	}

	return nil
}

// getPushTargets 获取推送目标用户列表
func (s *convMsgConsumerServiceImpl) getPushTargets(msg *sdkws.MessageData) []string {
	// 单聊：推送给接收者（排除发送者自己）
	if util.IsSingleChat(msg) {
		if msg.RecvID != "" && msg.RecvID != msg.SendID {
			return []string{msg.RecvID}
		}
		return nil
	}

	// 群聊：TODO 从会话中获取成员列表
	// 目前暂时返回空，后续可以从会话详情中获取成员列表
	klog.CtxWarnf(context.Background(), "[ConvMsgConsumer] group chat push not implemented",
		"conv_id", msg.ConvID)
	return nil
}

// buildPushMessage 构造 PushMessage
func (s *convMsgConsumerServiceImpl) buildPushMessage(msg *sdkws.MessageData) *backbon.PushMessage {
	// 序列化消息数据
	payload, err := msg.Marshal(nil)
	if err != nil {
		klog.CtxErrorf(context.Background(), "[ConvMsgConsumer] marshal message failed",
			"conv_id", msg.ConvID,
			"error", err.Error())
		// 如果序列化失败，返回空 payload
		payload = nil
	}

	// 构造元数据
	metadata := map[string]string{
		"conv_id":   msg.ConvID,
		"msg_id":    msg.SMessageID,
		"conv_type": fmt.Sprintf("%d", msg.ConvType),
	}

	return &backbon.PushMessage{
		RequestID: msg.SMessageID,
		Type:      "message",
		Service:   "message-service",
		Method:    "receive",
		Payload:   payload,
		Timestamp: time.Now().Unix(),
		Metadata:  metadata,
	}
}

// getBackbonServiceClient 获取 backbon 服务客户端
func (s *convMsgConsumerServiceImpl) getBackbonServiceClient(ctx context.Context) (backbonservice.Client, error) {
	return s.serviceCtx.GetBackbonServiceClient(ctx)
}
