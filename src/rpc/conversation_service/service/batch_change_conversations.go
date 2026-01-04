package service

import (
	"context"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
)

// BatchChangeConversations 批量更改会话（会话状态、已读状态、置顶状态、属性等）
func (s *conversationServiceImpl) BatchChangeConversations(ctx context.Context, req *sdkws.BatchChangeConversationsRequest) (*sdkws.BatchChangeConversationsResponse, error) {
	cmdMessages := req.GetCmdMessages()

	if len(cmdMessages) == 0 {
		return &sdkws.BatchChangeConversationsResponse{
			Results: []*sdkws.CmdMessageOptResult{},
		}, nil
	}

	// TODO: 实现批量更改会话的逻辑
	// 这里需要根据 cmdMessages 中的命令类型（cmd）执行不同的操作：
	// - 2001: 会话状态发生改变（删除）
	// - 2002: 会话已读状态发生改变
	// - 2003: 置顶状态发生改变
	// - 2005: 会话syncExt发生改变
	// 目前先返回未实现的结果

	klog.CtxWarnf(ctx, "[ConversationService] BatchChangeConversations not fully implemented",
		"cmd_count", len(cmdMessages))

	results := make([]*sdkws.CmdMessageOptResult, 0, len(cmdMessages))
	for _, cmdMsg := range cmdMessages {
		result := &sdkws.CmdMessageOptResult{
			Cmd:      cmdMsg.GetCmd(),
			Id:       cmdMsg.GetId(),
			ErrorCode: 0,
		}

		// TODO: 根据 cmd 执行相应的更新操作
		// 暂时标记为成功（实际需要实现具体的更新逻辑）

		results = append(results, result)
	}

	return &sdkws.BatchChangeConversationsResponse{
		Results: results,
	}, nil
}

