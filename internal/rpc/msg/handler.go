package msg

import (
	"context"
	msg "github.com/roc/roc-im-server/internal/kitex_gen/msg"
	sdkws "github.com/roc/roc-im-server/internal/kitex_gen/sdkws"
)

// MessageServiceImpl implements the last service interface defined in the IDL.
type MessageServiceImpl struct{}

// GetMaxSeq implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetMaxSeq(ctx context.Context, req *sdkws.GetMaxSeqReq) (resp *sdkws.GetMaxSeqResp, err error) {
	// TODO: Your code here...
	return
}

// GetMaxSeqs implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetMaxSeqs(ctx context.Context, req *msg.GetMaxSeqsReq) (resp *msg.SeqsInfoResp, err error) {
	// TODO: Your code here...
	return
}

// GetHasReadSeqs implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetHasReadSeqs(ctx context.Context, req *msg.GetHasReadSeqsReq) (resp *msg.SeqsInfoResp, err error) {
	// TODO: Your code here...
	return
}

// GetMsgByConversationIDs implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetMsgByConversationIDs(ctx context.Context, req *msg.GetMsgByConversationIDsReq) (resp *msg.GetMsgByConversationIDsResp, err error) {
	// TODO: Your code here...
	return
}

// GetConversationMaxSeq implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetConversationMaxSeq(ctx context.Context, req *msg.GetConversationMaxSeqReq) (resp *msg.GetConversationMaxSeqResp, err error) {
	// TODO: Your code here...
	return
}

// PullMessageBySeqs implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) PullMessageBySeqs(ctx context.Context, req *sdkws.PullMessageBySeqsReq) (resp *sdkws.PullMessageBySeqsResp, err error) {
	// TODO: Your code here...
	return
}

// GetSeqMessage implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetSeqMessage(ctx context.Context, req *msg.GetSeqMessageReq) (resp *msg.GetSeqMessageResp, err error) {
	// TODO: Your code here...
	return
}

// SearchMessage implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) SearchMessage(ctx context.Context, req *msg.SearchMessageReq) (resp *msg.SearchMessageResp, err error) {
	// TODO: Your code here...
	return
}

// SendMsg implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) SendMsg(ctx context.Context, req *msg.SendMsgReq) (resp *msg.SendMsgResp, err error) {
	// TODO: Your code here...
	return
}

// SendSimpleMsg implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) SendSimpleMsg(ctx context.Context, req *msg.SendSimpleMsgReq) (resp *msg.SendSimpleMsgResp, err error) {
	// TODO: Your code here...
	return
}

// SetUserConversationsMinSeq implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) SetUserConversationsMinSeq(ctx context.Context, req *msg.SetUserConversationsMinSeqReq) (resp *msg.SetUserConversationsMinSeqResp, err error) {
	// TODO: Your code here...
	return
}

// ClearConversationsMsg implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) ClearConversationsMsg(ctx context.Context, req *msg.ClearConversationsMsgReq) (resp *msg.ClearConversationsMsgResp, err error) {
	// TODO: Your code here...
	return
}

// UserClearAllMsg implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) UserClearAllMsg(ctx context.Context, req *msg.UserClearAllMsgReq) (resp *msg.UserClearAllMsgResp, err error) {
	// TODO: Your code here...
	return
}

// DeleteMsgs implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) DeleteMsgs(ctx context.Context, req *msg.DeleteMsgsReq) (resp *msg.DeleteMsgsResp, err error) {
	// TODO: Your code here...
	return
}

// DeleteMsgPhysicalBySeq implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) DeleteMsgPhysicalBySeq(ctx context.Context, req *msg.DeleteMsgPhysicalBySeqReq) (resp *msg.DeleteMsgPhysicalBySeqResp, err error) {
	// TODO: Your code here...
	return
}

// DeleteMsgPhysical implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) DeleteMsgPhysical(ctx context.Context, req *msg.DeleteMsgPhysicalReq) (resp *msg.DeleteMsgPhysicalResp, err error) {
	// TODO: Your code here...
	return
}

// SetSendMsgStatus implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) SetSendMsgStatus(ctx context.Context, req *msg.SetSendMsgStatusReq) (resp *msg.SetSendMsgStatusResp, err error) {
	// TODO: Your code here...
	return
}

// GetSendMsgStatus implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetSendMsgStatus(ctx context.Context, req *msg.GetSendMsgStatusReq) (resp *msg.GetSendMsgStatusResp, err error) {
	// TODO: Your code here...
	return
}

// RevokeMsg implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) RevokeMsg(ctx context.Context, req *msg.RevokeMsgReq) (resp *msg.RevokeMsgResp, err error) {
	// TODO: Your code here...
	return
}

// MarkMsgsAsRead implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) MarkMsgsAsRead(ctx context.Context, req *msg.MarkMsgsAsReadReq) (resp *msg.MarkMsgsAsReadResp, err error) {
	// TODO: Your code here...
	return
}

// MarkConversationAsRead implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) MarkConversationAsRead(ctx context.Context, req *msg.MarkConversationAsReadReq) (resp *msg.MarkConversationAsReadResp, err error) {
	// TODO: Your code here...
	return
}

// SetConversationHasReadSeq implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) SetConversationHasReadSeq(ctx context.Context, req *msg.SetConversationHasReadSeqReq) (resp *msg.SetConversationHasReadSeqResp, err error) {
	// TODO: Your code here...
	return
}

// GetConversationsHasReadAndMaxSeq implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetConversationsHasReadAndMaxSeq(ctx context.Context, req *msg.GetConversationsHasReadAndMaxSeqReq) (resp *msg.GetConversationsHasReadAndMaxSeqResp, err error) {
	// TODO: Your code here...
	return
}

// GetActiveUser implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetActiveUser(ctx context.Context, req *msg.GetActiveUserReq) (resp *msg.GetActiveUserResp, err error) {
	// TODO: Your code here...
	return
}

// GetActiveGroup implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetActiveGroup(ctx context.Context, req *msg.GetActiveGroupReq) (resp *msg.GetActiveGroupResp, err error) {
	// TODO: Your code here...
	return
}

// GetServerTime implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetServerTime(ctx context.Context, req *msg.GetServerTimeReq) (resp *msg.GetServerTimeResp, err error) {
	// TODO: Your code here...
	return
}

// ClearMsg implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) ClearMsg(ctx context.Context, req *msg.ClearMsgReq) (resp *msg.ClearMsgResp, err error) {
	// TODO: Your code here...
	return
}

// DestructMsgs implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) DestructMsgs(ctx context.Context, req *msg.DestructMsgsReq) (resp *msg.DestructMsgsResp, err error) {
	// TODO: Your code here...
	return
}

// GetActiveConversation implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetActiveConversation(ctx context.Context, req *msg.GetActiveConversationReq) (resp *msg.GetActiveConversationResp, err error) {
	// TODO: Your code here...
	return
}

// SetUserConversationMaxSeq implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) SetUserConversationMaxSeq(ctx context.Context, req *msg.SetUserConversationMaxSeqReq) (resp *msg.SetUserConversationMaxSeqResp, err error) {
	// TODO: Your code here...
	return
}

// SetUserConversationMinSeq implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) SetUserConversationMinSeq(ctx context.Context, req *msg.SetUserConversationMinSeqReq) (resp *msg.SetUserConversationMinSeqResp, err error) {
	// TODO: Your code here...
	return
}

// GetLastMessageSeqByTime implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetLastMessageSeqByTime(ctx context.Context, req *msg.GetLastMessageSeqByTimeReq) (resp *msg.GetLastMessageSeqByTimeResp, err error) {
	// TODO: Your code here...
	return
}

// GetLastMessage implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) GetLastMessage(ctx context.Context, req *msg.GetLastMessageReq) (resp *msg.GetLastMessageResp, err error) {
	// TODO: Your code here...
	return
}
