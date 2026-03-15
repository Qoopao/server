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

// FetchUserRecentConvList 混链拉取：获取用户最近的会话列表（包含会话和消息）
func (h *ConversationAPI) FetchUserRecentConvList(ctx context.Context, req *sdkws.FetchUserRecentConvListRequest) (resp *sdkws.FetchUserRecentConvListResponse, err error) {
	if req == nil {
		return &sdkws.FetchUserRecentConvListResponse{
			Conversations: []*sdkws.ConversationData{},
		}, nil
	}

	if req.GetUserID() == "" {
		return &sdkws.FetchUserRecentConvListResponse{
			Conversations: []*sdkws.ConversationData{},
			Error:         "user_id is empty",
		}, nil
	}

	resp, err = h.svc.FetchUserRecentConvList(ctx, req)
	if err != nil {
		klog.CtxErrorf(ctx, "FetchUserRecentConvList failed",
			"user_id", req.GetUserID(),
			"error", err.Error())
	}
	return resp, err
}

// UserMessageIntegrityCheck 混链连续性检查：检查用户消息的完整性，并补齐缺失的会话信息
func (h *ConversationAPI) UserMessageIntegrityCheck(ctx context.Context, req *sdkws.UserMessageIntegrityCheckRequest) (resp *sdkws.UserMessageIntegrityCheckResponse, err error) {
	if req == nil {
		return &sdkws.UserMessageIntegrityCheckResponse{
			IsIntegrity:   false,
			Left:          0,
			Right:         0,
			Conversations: []*sdkws.ConversationData{},
		}, nil
	}

	if req.GetUserID() == "" {
		return &sdkws.UserMessageIntegrityCheckResponse{
			IsIntegrity:   false,
			Left:          req.GetLeft(),
			Right:         req.GetRight(),
			Conversations: []*sdkws.ConversationData{},
		}, nil
	}

	resp, err = h.svc.UserMessageIntegrityCheck(ctx, req)
	if err != nil {
		klog.CtxErrorf(ctx, "UserMessageIntegrityCheck failed",
			"user_id", req.GetUserID(),
			"left", req.GetLeft(),
			"right", req.GetRight(),
			"error", err.Error())
	}
	return resp, err
}

// BatchGetConversations 批量获取会话：根据会话ID列表批量获取会话详情
func (h *ConversationAPI) BatchGetConversations(ctx context.Context, req *sdkws.BatchGetConversationsRequest) (resp *sdkws.BatchGetConversationsResponse, err error) {
	if req == nil {
		return &sdkws.BatchGetConversationsResponse{
			Results: []*sdkws.GetConversationResult{},
		}, nil
	}

	resp, err = h.svc.BatchGetConversations(ctx, req)
	if err != nil {
		klog.CtxErrorf(ctx, "BatchGetConversations failed",
			"error", err.Error())
		return &sdkws.BatchGetConversationsResponse{
			Results: []*sdkws.GetConversationResult{},
		}, err
	}

	return resp, nil
}

// CreateGroup 创建群聊（对齐客户端 CreateGroupContext：owner_user_id, member_user_ids, group_name）
func (h *ConversationAPI) CreateGroup(ctx context.Context, req *sdkws.CreateGroupRequest) (resp *sdkws.CreateGroupResponse, err error) {
	if req == nil {
		return &sdkws.CreateGroupResponse{ErrorCode: 400, ErrorMsg: "request is nil"}, nil
	}
	return h.svc.CreateGroup(ctx, req)
}

// InviteGroupMembers 邀请进群（对齐客户端 InviteGroupMembersContext：conv_id, member_user_ids）
func (h *ConversationAPI) InviteGroupMembers(ctx context.Context, req *sdkws.InviteGroupMembersRequest) (resp *sdkws.InviteGroupMembersResponse, err error) {
	if req == nil {
		return &sdkws.InviteGroupMembersResponse{ErrorCode: 400, ErrorMsg: "request is nil"}, nil
	}
	return h.svc.InviteGroupMembers(ctx, req)
}

// 编译期检查，确保实现了 conversationpb.ConversationService 接口
var _ conversationpb.ConversationService = (*ConversationAPI)(nil)
