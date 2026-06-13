package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	consts "github.com/rhp-QE/roc-im-server/src/const"
)

// InviteGroupMembers 邀请进群（对齐客户端 InviteGroupMembersContext：conv_id, member_user_ids）
func (s *conversationUsecase) InviteGroupMembers(ctx context.Context, req *sdkws.InviteGroupMembersRequest) (*sdkws.InviteGroupMembersResponse, error) {
	if req == nil || req.GetConvID() == "" {
		return &sdkws.InviteGroupMembersResponse{
			ErrorCode: 400,
			ErrorMsg:  "conv_id required",
		}, nil
	}
	if len(req.GetMemberUIDs()) == 0 {
		return &sdkws.InviteGroupMembersResponse{
			ErrorCode: 400,
			ErrorMsg:  "member_uids required",
		}, nil
	}

	convID := req.GetConvID()
	convMap, err := s.repo.BatchGetConversations(ctx, []string{convID}, "")
	if err != nil {
		return &sdkws.InviteGroupMembersResponse{ErrorCode: 500, ErrorMsg: err.Error()}, nil
	}
	conv, ok := convMap[convID]
	if !ok || conv == nil {
		return &sdkws.InviteGroupMembersResponse{ErrorCode: 404, ErrorMsg: "conversation not found"}, nil
	}

	var existing []string
	if conv.Members != "" {
		_ = json.Unmarshal([]byte(conv.Members), &existing)
	}
	seen := make(map[string]bool)
	for _, u := range existing {
		seen[u] = true
	}
	for _, u := range req.GetMemberUIDs() {
		if !seen[u] {
			existing = append(existing, u)
			seen[u] = true
		}
	}
	newMembersJSON, _ := json.Marshal(existing)
	conv.Members = string(newMembersJSON)
	conv.Version = time.Now().UnixNano()

	if err = s.repo.UpdateConversation(ctx, conv); err != nil {
		return &sdkws.InviteGroupMembersResponse{ErrorCode: 500, ErrorMsg: err.Error()}, nil
	}

	// 向被邀请用户下推「被邀请进群」cmd（由本业务构造 cmd，通用推送层不关心含义）
	cmdMsg := s.buildInvitedToGroupCmd(ctx, conv)
	if cmdMsg != nil {
		s.pushCmdToUsers(ctx, req.GetMemberUIDs(), cmdMsg, conv.ConvID)
	}

	return &sdkws.InviteGroupMembersResponse{ErrorCode: 0}, nil
}

// buildInvitedToGroupCmd 构造「被邀请进群」cmd 消息体；序列化失败时打日志并返回 nil。
func (s *conversationUsecase) buildInvitedToGroupCmd(ctx context.Context, conv *sdkws.ConversationData) *sdkws.CmdMessage {
	if conv == nil {
		return nil
	}
	data, err := conv.Marshal(nil)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationUsecase] buildInvitedToGroupCmd marshal conv failed: conv_id=%s err=%v", conv.ConvID, err)
		return nil
	}
	return &sdkws.CmdMessage{
		Id:      consts.CmdIDInvitedToGroup,
		Data:    data,
		Version: conv.Version,
	}
}
