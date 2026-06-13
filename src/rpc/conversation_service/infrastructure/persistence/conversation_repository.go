package persistence

import (
	"context"

	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	deps "github.com/rhp-QE/roc-im-server/src/rpc/conversation_service/infrastructure/deps"
)

// ConversationRepository 定义会话相关持久化接口。
type ConversationRepository interface {
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

	// CreateConversation 写入会话文档（仅做 conv 转 doc + 入库，不区分会话类型）
	CreateConversation(ctx context.Context, conv *sdkws.ConversationData) error

	// UpdateConversation 更新会话文档（仅做 conv 转 doc + UpdateOne）
	UpdateConversation(ctx context.Context, conv *sdkws.ConversationData) error
}

type conversationRepository struct {
	serviceCtx deps.ServiceContext
}

// NewConversationRepository 创建 ConversationRepository 实现
func NewConversationRepository(serviceCtx deps.ServiceContext) ConversationRepository {
	return &conversationRepository{
		serviceCtx: serviceCtx,
	}
}
