package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/google/uuid"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
	consts "github.com/rhp-QE/roc-im-server/src/const"
)

// CreateGroup 创建群聊（对齐客户端 CreateGroupContext：owner_user_id, member_user_ids, group_name）
func (s *conversationServiceImpl) CreateGroup(ctx context.Context, req *sdkws.CreateGroupRequest) (*sdkws.CreateGroupResponse, error) {
	// 1. 数据检查
	if invalidResp, ok := s.validateCreateGroupRequest(req); ok {
		return invalidResp, nil
	}
	// 2. 会话信息构造
	conv, err := s.buildGroupConversationData(ctx, req)
	if err != nil {
		return &sdkws.CreateGroupResponse{ErrorCode: 500, ErrorMsg: err.Error()}, nil
	}
	// 3. 数据库写入
	if err := s.createConversationStorage(ctx, conv); err != nil {
		return &sdkws.CreateGroupResponse{ErrorCode: 500, ErrorMsg: err.Error()}, nil
	}
	// 4. 被邀请用户命令消息下推（由本业务构造「会话创建」cmd，通用推送层不关心 cmd 含义）
	userIDs := s.collectGroupPushTargets(req)
	cmdMsg := s.buildConversationCreatedCmd(ctx, conv)
	if cmdMsg != nil {
		s.pushCmdToUsers(ctx, userIDs, cmdMsg, conv.ConvID)
	}

	return &sdkws.CreateGroupResponse{
		ErrorCode:    0,
		Conversation: conv,
	}, nil
}

// validateCreateGroupRequest 校验创建群聊请求；无效时返回 (错误响应, true)，调用方应直接 return。
func (s *conversationServiceImpl) validateCreateGroupRequest(req *sdkws.CreateGroupRequest) (*sdkws.CreateGroupResponse, bool) {
	if req == nil || req.GetOwnerID() == "" {
		return &sdkws.CreateGroupResponse{
			ErrorCode: 400,
			ErrorMsg:  "owner_id required",
		}, true
	}
	return nil, false
}

// buildGroupConversationData 构造群聊会话数据（convID、成员 JSON、Version 等）；失败时在内部打错误日志并返回 error。
func (s *conversationServiceImpl) buildGroupConversationData(ctx context.Context, req *sdkws.CreateGroupRequest) (*sdkws.ConversationData, error) {
	newID, err := uuid.NewRandom()
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationService] CreateGroup uuid.NewRandom failed: %v", err)
		return nil, err
	}
	convID := fmt.Sprintf("0:%d:%s", orm.ConvTypeGroupChat, newID.String())
	membersJSON, _ := json.Marshal(req.GetMemberUIDs())
	return &sdkws.ConversationData{
		ConvID:   convID,
		OwnerID:  req.GetOwnerID(),
		ConvType: int32(orm.ConvTypeGroupChat),
		Name:     req.GetName(),
		Members:  string(membersJSON),
		Version:  time.Now().UnixNano(),
	}, nil
}

// createConversationStorage 将会话写入数据库；失败时在内部打错误日志并返回 error。
func (s *conversationServiceImpl) createConversationStorage(ctx context.Context, conv *sdkws.ConversationData) error {
	if err := s.storage.CreateConversation(ctx, conv); err != nil {
		klog.CtxErrorf(ctx, "[ConversationService] CreateGroup storage.CreateConversation failed: conv_id=%s err=%v", conv.ConvID, err)
		return err
	}
	return nil
}

// buildConversationCreatedCmd 构造「会话创建」cmd 消息体；序列化失败时打日志并返回 nil。
func (s *conversationServiceImpl) buildConversationCreatedCmd(ctx context.Context, conv *sdkws.ConversationData) *sdkws.CmdMessage {
	if conv == nil {
		return nil
	}
	data, err := conv.Marshal(nil)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConversationService] buildConversationCreatedCmd marshal conv failed: conv_id=%s err=%v", conv.ConvID, err)
		return nil
	}
	return &sdkws.CmdMessage{
		Id:      consts.CmdIDInvitedToGroup,
		Data:    data,
		Version: conv.Version,
		Cmd:     consts.CmdIDInvitedToGroup,
	}
}

// collectGroupPushTargets 收集创建群聊的推送目标：群主 + 去重后的成员（排除空与群主）。
func (s *conversationServiceImpl) collectGroupPushTargets(req *sdkws.CreateGroupRequest) []string {
	userIDs := make([]string, 0, 1+len(req.GetMemberUIDs()))
	userIDs = append(userIDs, req.GetOwnerID())
	for _, u := range req.GetMemberUIDs() {
		if u != "" && u != req.GetOwnerID() {
			userIDs = append(userIDs, u)
		}
	}
	return userIDs
}
