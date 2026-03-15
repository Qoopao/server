package service

import (
	"context"
	"fmt"

	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	sequencepb "github.com/rhp-QE/roc-im-server/kitex_gen/sequence"
	sequence "github.com/rhp-QE/roc-im-server/kitex_gen/sequence/sequenceservice"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
	"github.com/rhp-QE/roc-im-server/src/common/util"
	convstorage "github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/storage"
)

const MixUPrefix = "mix2user"

// updateUserRecentConvsIfSmallConv 更新用户最近会话链（单聊 / 群聊都更新，系统通知忽略）
func (s *convMsgConsumerServiceImpl) updateUserRecentConvsIfSmallConv(ctx context.Context, msg *sdkws.MessageData) error {
	if msg == nil || util.IsSystemNotification(msg) {
		return nil
	}

	members, err := s.getConvMembersForMessage(ctx, msg)
	if err != nil || len(members) == 0 {
		return err
	}

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

// getNextVersion 获取单个用户混链版本号 （用户会话序列号）
func getNextVersion(ctx context.Context, seqClient sequence.Client, userID, convID string) (int64, error) {
	req := &sequencepb.GetNextSeqIncRequest{
		Id: fmt.Sprintf("%s:%s", MixUPrefix, userID),
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

// getConvMembersForMessage 根据消息类型获取需要维护最近会话链的成员列表
// - 单聊：发送者 + 接收者
// - 群聊：从会话文档中解析成员 JSON
// - 其他类型：返回空
func (s *convMsgConsumerServiceImpl) getConvMembersForMessage(ctx context.Context, msg *sdkws.MessageData) ([]string, error) {
	switch msg.ConvType {
	case int32(orm.ConvTypeSingleChat):
		// 单聊：通过 convID 在 service 层解析成员（findMembersForConv 会优先走 convID 解析，不查 DB）
		return s.findMembersForConv(ctx, msg.ConvID)
	case int32(orm.ConvTypeGroupChat):
		// 群聊：复用统一的成员查询逻辑（包含群主）
		return s.findMembersForConv(ctx, msg.ConvID)
	default:
		return nil, nil
	}
}
