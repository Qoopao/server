package storage

import (
	"context"

	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	servicecontext "github.com/rhp-QE/roc-im-server/src/rpc/conversation_service/service_context"
)

// ConversationStorage 定义会话相关的持久化接口（存储层）
type ConversationStorage interface {
	// FetchUserRecentConvListByVersion 根据版本号查询大于指定 version 的所有最新数据
	// userID: 用户ID
	// version: 版本号，查询 version > 指定值的所有数据
	// 返回：会话列表（按 version 升序，最旧的在前）
	FetchUserRecentConvListByVersion(ctx context.Context, userID string, version int64) ([]*sdkws.ConversationData, error)

	// FetchUserRecentConvListByVersionRange 根据版本号区间查询数据
	// userID: 用户ID
	// lowVersion: 最小版本号（包含）
	// upVersion: 最大版本号（包含）
	// 返回：会话列表（按 version 升序）
	FetchUserRecentConvListByVersionRange(ctx context.Context, userID string, lowVersion int64, upVersion int64) ([]*sdkws.ConversationData, error)

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
