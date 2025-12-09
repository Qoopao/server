package conversation

import (
	"context"
	"errors"

	conversation "github.com/rhp-QE/roc-im-server/internal/kitex_gen/conversation"
	sdkws "github.com/rhp-QE/roc-im-server/internal/kitex_gen/sdkws"
)

func (s *ConversationServiceImpl) fetchConvMessaegList(ctx context.Context, req *sdkws.FetchConvMessageListReq) (resp *sdkws.FetchConvMessageListResp, err error) {
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

func (s *ConversationServiceImpl) fetchUserMessaegList(ctx context.Context, req *sdkws.FetchUserMessageListReq) (resp *sdkws.FetchUserMessageListResp, err error) {
	//1、获取有哪些会话
	//2、获取会话消息
	//3、获取会话信息
	var (
		cursor    = req.Cursor
		haveMore  bool
		convList  []string
		convsInfo []*sdkws.ConversationInfo
		start     int64
		stop      int64
	)

	// 获取有哪些会话
	if req.News {
		cursor = -1
	}
	convList, haveMore, start, stop, err = s.MessageDB.GetUserConvList(ctx, req.UserID, cursor, req.Limit, req.Forward)

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
		Start:     start,
		Stop:      stop,
	}

	return resp, nil
}

func convertServerConvToSDKConv(conv *conversation.ConversationInfo) (*sdkws.ConversationInfo, error) {
	if conv == nil {
		return nil, errors.New("conv is nil")
	}

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

	lastMsg := messages[len(messages)-1]

	sdkConv.Msgs = make([]*sdkws.MessageUnion, 0)
	for _, message := range messages {
		sdkConv.Msgs = append(sdkConv.Msgs, &sdkws.MessageUnion{
			IsCmd:  false,
			Msg:    message,
			CmdMsg: nil,
		})

		if message.Seq > lastMsg.Seq {
			lastMsg = message
		}
	}

	sdkConv.LastMsg = lastMsg
}
