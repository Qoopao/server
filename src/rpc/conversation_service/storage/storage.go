package storage

import (
	"context"

	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	servicecontext "github.com/rhp-QE/roc-im-server/src/rpc/conversation_service/service_context"
)

const (
	// collectionConversations 会话详情集合名称（实际集合名会加上前缀 "im_db_"）
	collectionConversations = "conversations"
	// collectionUserRecentConversations 用户最近会话链集合名称
	collectionUserRecentConversations = "user_recent_conversations"
	// collectionMessages 消息集合名称
	collectionMessages = "messages"
)

// ConversationStorage 定义会话相关的持久化接口（存储层）
type ConversationStorage interface {
	// FetchUserRecentConvList 获取用户最近的会话列表（包含会话和消息）
	// userID: 用户ID
	// cursor: 游标位置（updated_at 的时间戳或索引位置）
	// limit: 返回数量限制
	// forward: true 表示向前查询（新会话），false 表示向后查询（旧会话）
	// 返回：会话列表（按 updated_at 降序，最新的在前）
	FetchUserRecentConvList(ctx context.Context, userID string, cursor int64, limit int64, forward bool) ([]*sdkws.ConversationData, int64, int64, bool, error)

	// BatchGetConversations 根据会话ID列表批量获取会话详情
	// convIDs: 会话ID列表
	// ownerID: 会话拥有者用户ID（可选，用于过滤）
	// 返回：会话ID到会话数据的映射
	BatchGetConversations(ctx context.Context, convIDs []string, ownerID string) (map[string]*sdkws.ConversationData, error)
}

type conversationStorageImpl struct {
	serviceCtx servicecontext.ServiceContext
}

// NewConversationStorage 创建 ConversationStorage 实现
func NewConversationStorage(serviceCtx servicecontext.ServiceContext) ConversationStorage {
	return &conversationStorageImpl{
		serviceCtx: serviceCtx,
	}
}

