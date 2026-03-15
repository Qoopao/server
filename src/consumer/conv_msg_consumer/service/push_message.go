package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	backbon "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/backbon"
	backbonservice "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/backbon/backbonservice"
	frontier "github.com/rhp-QE/roc-foundation-service/long_connection_service/src/frontier"
	"github.com/rhp-QE/roc-foundation-util-go/array"
	"github.com/rhp-QE/roc-foundation-util-go/stringutil"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
	consts "github.com/rhp-QE/roc-im-server/src/const"
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
	if msg == nil || msg.ConvID == "" {
		return nil
	}

	convID := msg.ConvID
	convMembers, err := s.findMembersForConv(context.Background(), convID)
	if err != nil {
		klog.CtxErrorf(context.Background(), "[ConvMsgConsumer] findMembersForConv failed",
			"conv_id", convID,
			"error", err.Error())
		return nil
	}

	if !stringutil.IsEmpty(msg.SendID) {
		convMembers = stringutil.FilterExclude(convMembers, msg.SendID)
	}

	return convMembers
}

// findMembersForConv 在 service 层按 convID 解析成员：
// - 单聊：0:1:uid1:uid2 直接从 convID 解析成员
// - 群聊/其他：调用 storage.GetConversationDoc 查询会话文档，再从 ConversationData.Members + OwnerID 构造成员列表
func (s *convMsgConsumerServiceImpl) findMembersForConv(ctx context.Context, convID string) ([]string, error) {
	if convID == "" {
		return nil, nil
	}

	parts := strings.Split(convID, ":")
	// 单聊：严格匹配 0:1:uid1:uid2（至少 4 段），才走 convID 解析逻辑
	if len(parts) >= 4 && parts[0] == "0" && parts[1] == fmt.Sprintf("%d", orm.ConvTypeSingleChat) {
		userPart := strings.Join(parts[2:], ":")
		userIDs := strings.Split(userPart, ":")
		members := make([]string, 0, len(userIDs))
		seen := make(map[string]struct{}, len(userIDs))
		for _, uid := range userIDs {
			if uid == "" {
				continue
			}
			if _, ok := seen[uid]; ok {
				continue
			}
			seen[uid] = struct{}{}
			members = append(members, uid)
		}
		return members, nil
	}

	conv, err := s.storage.GetConversationData(ctx, convID)
	if err != nil || conv == nil {
		return nil, err
	}

	members := make([]string, 0)
	if conv.Members != "" {
		if err := json.Unmarshal([]byte(conv.Members), &members); err != nil {
			klog.CtxErrorf(ctx, "[ConvMsgConsumer] findMembersForConv unmarshal Members failed",
				"conv_id", convID,
				"error", err.Error())
			return nil, err
		}
	}

	if !stringutil.IsEmpty(conv.OwnerID) {
		members = append(members, conv.OwnerID)
	}

	// 对成员列表去重
	return array.Distinct(members), nil
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
		Type:      frontier.MessageTypePush,
		Service:   consts.BackserviceIMName,
		Method:    consts.SDKWSMethodPushUserMessage,
		Payload:   payload,
		Timestamp: time.Now().Unix(),
		Metadata:  metadata,
	}
}

// getBackbonServiceClient 获取 backbon 服务客户端
func (s *convMsgConsumerServiceImpl) getBackbonServiceClient(ctx context.Context) (backbonservice.Client, error) {
	return s.serviceCtx.GetBackbonServiceClient(ctx)
}
