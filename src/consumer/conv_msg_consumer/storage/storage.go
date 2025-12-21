package storage

import (
	"context"

	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	servicecontext "github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/service_context"
)

const (
	// collectionMessages 存储消息详情（KV：msg_id -> msg.pb），带索引 conv_id/seq/time
	collectionMessages = "messages"
	// collectionConversations 存储会话详情快照（KV：conv_id -> last_msg.pb 等）
	collectionConversations = "conversations"
	// collectionUserRecentConversations 存储用户最近会话链（KV：uid -> conversations[]）
	collectionUserRecentConversations = "user_recent_conversations"
)

// ConvMsgStorage 会话消息相关的持久化接口
type ConvMsgStorage interface {
	// SaveMessage 保存一条消息（包含 conv_id/seq 信息，支持按 seq 查询单链）
	SaveMessage(ctx context.Context, msgID string, msg *sdkws.MessageData) error

	// UpdateConvWithMessage 更新/创建会话详情（根据消息数据保存完整的 ConversationData，更新 last_message_seq 和 updated_at）
	UpdateConvWithMessage(ctx context.Context, msg *sdkws.MessageData) error

	// UpdateUserRecentConversation 更新用户最近会话链（将会话移到最前面）
	UpdateUserRecentConversation(ctx context.Context, userID string, convID string, lastSeq int64) error
}

type convMsgStorageImpl struct {
	serviceCtx servicecontext.ServiceContext
}

// NewConvMsgStorage 创建 ConvMsgStorage 实例
// 注意：底层使用 MongoDB（通过 foundation storage 接入）
func NewConvMsgStorage(serviceCtx servicecontext.ServiceContext) ConvMsgStorage {
	return &convMsgStorageImpl{
		serviceCtx: serviceCtx,
	}
}

func (s *convMsgStorageImpl) getStore() foundationstorage.Storage {
	return s.serviceCtx.GetStorage()
}
