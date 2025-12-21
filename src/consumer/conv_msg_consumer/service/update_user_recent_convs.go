package service

import (
	"context"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/common/util"
)

// updateUserRecentConvsIfSmallConv 如果是单聊则更新用户最近会话链（更新发送者和接收者的最近会话链）
func (s *convMsgConsumerServiceImpl) updateUserRecentConvsIfSmallConv(ctx context.Context, msg *sdkws.MessageData) error {
	// 只处理单聊
	if !util.IsSingleChat(msg) {
		return nil
	}

	// 提取会话成员（发送者和接收者）
	members := util.GetMessageMembers(msg)

	// 循环更新每个成员的最近会话链
	var lastErr error
	for _, userID := range members {
		if err := s.storage.UpdateUserRecentConversation(ctx, userID, msg.ConvID, msg.Seq); err != nil {
			klog.CtxErrorf(ctx, "[ConvMsgConsumer] update user recent conversation failed",
				"user_id", userID,
				"conv_id", msg.ConvID,
				"error", err.Error())
			lastErr = err
			// 继续处理其他成员，但记录最后一个错误
		}
	}
	return lastErr
}
