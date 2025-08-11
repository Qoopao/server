package conversation

import (
	"context"

	conversation "github.com/roc/roc-im-server/internal/kitex_gen/conversation"
	sdkws "github.com/roc/roc-im-server/internal/kitex_gen/sdkws"
)

func (s *ConversationServiceImpl) fetchConvMessageList(ctx context.Context, req *sdkws.FetchConvMessageListReq) (resp *sdkws.FetchConvMessageListResp, err error) {
	msgList, haveMore, err := s.MessageDB.GetConvMessageList(ctx, req.ConvID, req.Cursor, req.Limit, req.Forward)
	if err != nil {
		return nil, err
	}

	resp = &sdkws.FetchConvMessageListResp{
		Messages: msgList,
		HaveMore: haveMore,
	}

	return resp, nil
}

func (s *ConversationServiceImpl) fetchUserMessageList(ctx context.Context, req *sdkws.FetchUserMessageListReq) (resp *sdkws.FetchUserMessageListResp, err error) {
	//1、获取有哪些会话
	//2、获取会话消息
	//3、获取会话信息
	var (
		convList  []string
		haveMore  bool
		convsInfo []*sdkws.ConversationInfo
	)

	// 获取有哪些会话
	convList, haveMore, err = s.MessageDB.GetUserConvList(ctx, req.UserID, req.Cursor, req.Limit, req.Forward)

	for _, convID := range convList {
		// 获取会话信息
		serverConv, _ := s.MessageDB.GetConversationInfo(ctx, convID)
		sdkConv, _ := convertServerConvToSDKConv(serverConv)

		// 获取会话消息
		msgList, _, _ := s.MessageDB.GetConvMessageList(ctx, convID, -1, 10, true)
		setConvMessages(msgList, sdkConv)

		if sdkConv != nil {
			convsInfo = append(convsInfo, sdkConv)
		}
	}

	if err != nil {
		return nil, err
	}

	resp = &sdkws.FetchUserMessageListResp{
		HasMore:   haveMore,
		ConvsInfo: convsInfo,
	}

	return resp, nil
}

func convertServerConvToSDKConv(conv *conversation.ConversationInfo) (*sdkws.ConversationInfo, error) {
	return &sdkws.ConversationInfo{
		ConvID:      conv.ConversationID,
		OwnerUserID: conv.OwnerUserID,
		ConvType:    conv.ConversationType,
		ConvName:    conv.ConversationName,
		ConvAvatar:  conv.ConversationAvatar,
	}, nil
}

func setConvMessages(messages []*sdkws.MsgData, sdkConv *sdkws.ConversationInfo) {
	if len(messages) == 0 || sdkConv == nil {
		return
	}

	sdkConv.Msgs = make([]*sdkws.MessageUnion, 0)
	for _, message := range messages {
		sdkConv.Msgs = append(sdkConv.Msgs, &sdkws.MessageUnion{
			IsCmd:  false,
			Msg:    message,
			CmdMsg: nil,
		})
	}

	sdkConv.LastMsg = messages[len(messages)-1]
}
