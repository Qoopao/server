package application

import (
	"context"

	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	deps "github.com/rhp-QE/roc-im-server/src/rpc/conversation_service/infrastructure/deps"
	"github.com/rhp-QE/roc-im-server/src/rpc/conversation_service/infrastructure/persistence"
)

// ConversationUsecase 定义会话链路应用用例接口。
type ConversationUsecase interface {
	// FetchUserRecentConvList 混链拉取：获取用户最近的会话列表（包含会话和消息）
	FetchUserRecentConvList(ctx context.Context, req *sdkws.FetchUserRecentConvListRequest) (*sdkws.FetchUserRecentConvListResponse, error)

	// UserMessageIntegrityCheck 混链连续性检查：检查用户消息的完整性，并补齐缺失的会话信息
	UserMessageIntegrityCheck(ctx context.Context, req *sdkws.UserMessageIntegrityCheckRequest) (*sdkws.UserMessageIntegrityCheckResponse, error)

	// BatchGetConversations 批量获取会话：根据会话ID列表批量获取会话详情
	BatchGetConversations(ctx context.Context, req *sdkws.BatchGetConversationsRequest) (*sdkws.BatchGetConversationsResponse, error)

	// CreateGroup 创建群聊（对齐客户端 CreateGroupContext：owner_user_id, member_user_ids, group_name）
	CreateGroup(ctx context.Context, req *sdkws.CreateGroupRequest) (*sdkws.CreateGroupResponse, error)

	// InviteGroupMembers 邀请进群（对齐客户端 InviteGroupMembersContext：conv_id, member_user_ids）
	InviteGroupMembers(ctx context.Context, req *sdkws.InviteGroupMembersRequest) (*sdkws.InviteGroupMembersResponse, error)
}

// conversationUsecase 是 ConversationUsecase 的具体实现
type conversationUsecase struct {
	repo       persistence.ConversationRepository
	serviceCtx deps.ServiceContext
}

// NewConversationUsecase 创建 ConversationUsecase 实例
func NewConversationUsecase(repo persistence.ConversationRepository, serviceCtx deps.ServiceContext) ConversationUsecase {
	return &conversationUsecase{
		repo:       repo,
		serviceCtx: serviceCtx,
	}
}
