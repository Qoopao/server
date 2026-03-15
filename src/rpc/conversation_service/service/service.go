package service

import (
	"context"

	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	servicecontext "github.com/rhp-QE/roc-im-server/src/rpc/conversation_service/service_context"
	"github.com/rhp-QE/roc-im-server/src/rpc/conversation_service/storage"
)

// ConversationService 定义会话服务的业务接口（逻辑层）
type ConversationService interface {
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

// conversationServiceImpl 是 ConversationService 的具体实现
type conversationServiceImpl struct {
	storage    storage.ConversationStorage
	serviceCtx servicecontext.ServiceContext
}

// NewConversationService 创建 ConversationService 实例
func NewConversationService(storage storage.ConversationStorage, serviceCtx servicecontext.ServiceContext) ConversationService {
	return &conversationServiceImpl{
		storage:    storage,
		serviceCtx: serviceCtx,
	}
}
