package api

import (
	"context"

	"github.com/cloudwego/kitex/pkg/klog"
	conversationpb "github.com/rhp-QE/roc-im-server/kitex_gen/conversation"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/rpc/conversation_service/service"
)

// ConversationAPI 实现 ConversationService 接口（接口层）
type ConversationAPI struct {
	svc service.ConversationService
}

// NewConversationAPI 创建 ConversationAPI 实例
func NewConversationAPI(svc service.ConversationService) *ConversationAPI {
	return &ConversationAPI{
		svc: svc,
	}
}

// BatchChangeConversations 批量更改会话（会话状态、已读状态、置顶状态、属性等）
func (h *ConversationAPI) BatchChangeConversations(ctx context.Context, req *sdkws.BatchChangeConversationsRequest) (resp *sdkws.BatchChangeConversationsResponse, err error) {
	// TODO: 实现参数校验和业务逻辑
	klog.CtxWarnf(ctx, "BatchChangeConversations not implemented")
	return &sdkws.BatchChangeConversationsResponse{
		Results: []*sdkws.CmdMessageOptResult{},
	}, nil
}

// FetchUserRecentConvList 混链拉取：获取用户最近的会话列表（包含会话和消息）
func (h *ConversationAPI) FetchUserRecentConvList(ctx context.Context, req *sdkws.FetchUserRecentConvListRequest) (resp *sdkws.FetchUserRecentConvListResponse, err error) {
	// TODO: 实现参数校验和业务逻辑
	klog.CtxWarnf(ctx, "FetchUserRecentConvList not implemented")
	return &sdkws.FetchUserRecentConvListResponse{
		Conversations: []*sdkws.ConversationData{},
		Left:          0,
		Right:         0,
		HasMore:       false,
	}, nil
}

// UserMessageIntegrityCheck 混链连续性检查：检查用户消息的完整性，并补齐缺失的会话信息
func (h *ConversationAPI) UserMessageIntegrityCheck(ctx context.Context, req *sdkws.UserMessageIntegrityCheckRequest) (resp *sdkws.UserMessageIntegrityCheckResponse, err error) {
	// TODO: 实现参数校验和业务逻辑
	klog.CtxWarnf(ctx, "UserMessageIntegrityCheck not implemented")
	return &sdkws.UserMessageIntegrityCheckResponse{
		IsIntegrity:   false,
		Left:          0,
		Right:         0,
		Conversations: []*sdkws.ConversationData{},
	}, nil
}

// BatchGetConversations 批量获取会话：根据会话ID列表批量获取会话详情
func (h *ConversationAPI) BatchGetConversations(ctx context.Context, req *sdkws.BatchGetConversationsRequest) (resp *sdkws.BatchGetConversationsResponse, err error) {
	// TODO: 实现参数校验和业务逻辑
	klog.CtxWarnf(ctx, "BatchGetConversations not implemented")
	return &sdkws.BatchGetConversationsResponse{
		Results: []*sdkws.GetConversationResult{},
	}, nil
}

// 编译期检查，确保实现了 conversationpb.ConversationService 接口
var _ conversationpb.ConversationService = (*ConversationAPI)(nil)
