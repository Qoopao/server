package service

import (
	"context"
	"fmt"

	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	sequencepb "github.com/rhp-QE/roc-im-server/kitex_gen/sequence"
	sequence "github.com/rhp-QE/roc-im-server/kitex_gen/sequence/sequenceservice"
	"github.com/rhp-QE/roc-im-server/src/common/util"
	convstorage "github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/storage"
)

// updateUserRecentConvsIfSmallConv 如果是单聊则更新用户最近会话链（更新发送者和接收者的最近会话链）
func (s *convMsgConsumerServiceImpl) updateUserRecentConvsIfSmallConv(ctx context.Context, msg *sdkws.MessageData) error {
	// 只处理单聊
	if !util.IsSingleChat(msg) {
		return nil
	}

	// 提取会话成员（发送者和接收者）
	members := util.GetMessageMembers(msg)

	// 获取 sequence-service 客户端（用户会话版本号服务）
	seqClient, err := s.serviceCtx.GetSequenceServiceClient(ctx)
	if err != nil {
		return err
	}

	// 循环更新每个成员的最近会话链
	for _, userID := range members {
		var version int64

		// 获取版本号
		if version, err = getNextVersion(ctx, seqClient, userID, msg.ConvID); err != nil {
			break
		}

		// 更新最近会话链
		if err := s.storage.UpdateUserRecentConversation(ctx, &convstorage.UpdateUserRecentConversationRequest{
			UserID:         userID,
			ConvID:         msg.ConvID,
			LastMessageSeq: msg.Seq,
			Version:        version,
		}); err != nil {
			break
		}
	}
	return nil
}

// getNextVersion 获取单个用户在指定会话下的最近会话版本号（用户会话序列号）
// 将 RPC 调用与错误检查封装在一个函数内，避免主业务逻辑被大量错误处理代码打断。
func getNextVersion(ctx context.Context, seqClient sequence.Client, userID, convID string) (int64, error) {
	req := &sequencepb.GetNextSeqIncRequest{
		Id: fmt.Sprintf("%s:%s", userID, convID),
	}

	seqResp, err := seqClient.GetNextSeqInc(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("call sequence-service GetNextSeqInc failed for user_id=%s conv_id=%s: %w", userID, convID, err)
	}
	if seqResp == nil {
		return 0, fmt.Errorf("sequence-service returned empty response for user_id=%s conv_id=%s", userID, convID)
	}
	if seqResp.ErrorCode != "" {
		return 0, fmt.Errorf("sequence-service business error for user_id=%s conv_id=%s code=%s msg=%s",
			userID, convID, seqResp.ErrorCode, seqResp.ErrorMsg)
	}

	return seqResp.Seq, nil
}
