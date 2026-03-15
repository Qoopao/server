package service

import (
	"context"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	backbon "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/backbon"
	frontier "github.com/rhp-QE/roc-foundation-service/long_connection_service/src/frontier"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	consts "github.com/rhp-QE/roc-im-server/src/const"
)

// pushCmdToUsers 向指定用户下推命令消息（通用：仅负责序列化 CmdMessage 并调用长连 PushData，不关心 cmd 类型与 payload 含义）
// 调用方负责构造 cmdMsg（Id、Data、Version 等），避免在通用逻辑中耦合具体业务（如会话创建、邀请进群等）。
func (s *conversationServiceImpl) pushCmdToUsers(ctx context.Context, userIDs []string, cmdMsg *sdkws.CmdMessage, requestID string) {
	if len(userIDs) == 0 || cmdMsg == nil {
		return
	}

	payload, err := cmdMsg.Marshal(nil)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationService] push cmd marshal CmdMessage failed: request_id=%s err=%v", requestID, err)
		return
	}

	client, err := s.serviceCtx.GetBackbonServiceClient(ctx)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationService] push cmd get backbon client failed: request_id=%s err=%v", requestID, err)
		return
	}

	req := &backbon.PushDataReq{
		UserIDs: userIDs,
		Message: &backbon.PushMessage{
			RequestID: requestID,
			Type:      frontier.MessageTypePush,
			Service:   consts.BackserviceIMName,
			Method:    consts.SDKWSMethodPushCmdMessage,
			Payload:   payload,
			Timestamp: time.Now().Unix(),
			Metadata:  map[string]string{"request_id": requestID},
		},
		Broadcast: false,
	}

	resp, err := client.PushData(ctx, req)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationService] push cmd failed: request_id=%s err=%v", requestID, err)
		return
	}
	if resp.GetFailCount() > 0 {
		klog.CtxWarnf(ctx, "[ConversationService] push cmd partial fail: request_id=%s success=%d fail=%d",
			requestID, resp.GetSuccessCount(), resp.GetFailCount())
	} else {
		klog.CtxDebugf(ctx, "[ConversationService] push cmd success: request_id=%s count=%d", requestID, len(userIDs))
	}
}
