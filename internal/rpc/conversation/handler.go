package conversation

import (
	"context"
	conversation "github.com/roc/roc-im-server/internal/kitex_gen/conversation"
)

// ConversationServiceImpl implements the last service interface defined in the IDL.
type ConversationServiceImpl struct{}

// GetConversation implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) GetConversation(ctx context.Context, req *conversation.GetConversationReq) (resp *conversation.GetConversationResp, err error) {
	// TODO: Your code here...
	return
}

// GetSortedConversationList implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) GetSortedConversationList(ctx context.Context, req *conversation.GetSortedConversationListReq) (resp *conversation.GetSortedConversationListResp, err error) {
	// TODO: Your code here...
	return
}

// GetAllConversations implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) GetAllConversations(ctx context.Context, req *conversation.GetAllConversationsReq) (resp *conversation.GetAllConversationsResp, err error) {
	// TODO: Your code here...
	return
}

// GetConversations implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) GetConversations(ctx context.Context, req *conversation.GetConversationsReq) (resp *conversation.GetConversationsResp, err error) {
	// TODO: Your code here...
	return
}

// SetConversation implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) SetConversation(ctx context.Context, req *conversation.SetConversationReq) (resp *conversation.SetConversationResp, err error) {
	// TODO: Your code here...
	return
}

// GetRecvMsgNotNotifyUserIDs implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) GetRecvMsgNotNotifyUserIDs(ctx context.Context, req *conversation.GetRecvMsgNotNotifyUserIDsReq) (resp *conversation.GetRecvMsgNotNotifyUserIDsResp, err error) {
	// TODO: Your code here...
	return
}

// CreateSingleChatConversations implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) CreateSingleChatConversations(ctx context.Context, req *conversation.CreateSingleChatConversationsReq) (resp *conversation.CreateSingleChatConversationsResp, err error) {
	// TODO: Your code here...
	return
}

// CreateGroupChatConversations implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) CreateGroupChatConversations(ctx context.Context, req *conversation.CreateGroupChatConversationsReq) (resp *conversation.CreateGroupChatConversationsResp, err error) {
	// TODO: Your code here...
	return
}

// SetConversationMaxSeq implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) SetConversationMaxSeq(ctx context.Context, req *conversation.SetConversationMaxSeqReq) (resp *conversation.SetConversationMaxSeqResp, err error) {
	// TODO: Your code here...
	return
}

// SetConversationMinSeq implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) SetConversationMinSeq(ctx context.Context, req *conversation.SetConversationMinSeqReq) (resp *conversation.SetConversationMinSeqResp, err error) {
	// TODO: Your code here...
	return
}

// GetConversationIDs implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) GetConversationIDs(ctx context.Context, req *conversation.GetConversationIDsReq) (resp *conversation.GetConversationIDsResp, err error) {
	// TODO: Your code here...
	return
}

// SetConversations implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) SetConversations(ctx context.Context, req *conversation.SetConversationsReq) (resp *conversation.SetConversationsResp, err error) {
	// TODO: Your code here...
	return
}

// GetUserConversationIDsHash implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) GetUserConversationIDsHash(ctx context.Context, req *conversation.GetUserConversationIDsHashReq) (resp *conversation.GetUserConversationIDsHashResp, err error) {
	// TODO: Your code here...
	return
}

// GetConversationsByConversationID implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) GetConversationsByConversationID(ctx context.Context, req *conversation.GetConversationsByConversationIDReq) (resp *conversation.GetConversationsByConversationIDResp, err error) {
	// TODO: Your code here...
	return
}

// GetConversationOfflinePushUserIDs implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) GetConversationOfflinePushUserIDs(ctx context.Context, req *conversation.GetConversationOfflinePushUserIDsReq) (resp *conversation.GetConversationOfflinePushUserIDsResp, err error) {
	// TODO: Your code here...
	return
}

// GetConversationNotReceiveMessageUserIDs implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) GetConversationNotReceiveMessageUserIDs(ctx context.Context, req *conversation.GetConversationNotReceiveMessageUserIDsReq) (resp *conversation.GetConversationNotReceiveMessageUserIDsResp, err error) {
	// TODO: Your code here...
	return
}

// UpdateConversation implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) UpdateConversation(ctx context.Context, req *conversation.UpdateConversationReq) (resp *conversation.UpdateConversationResp, err error) {
	// TODO: Your code here...
	return
}

// GetFullOwnerConversationIDs implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) GetFullOwnerConversationIDs(ctx context.Context, req *conversation.GetFullOwnerConversationIDsReq) (resp *conversation.GetFullOwnerConversationIDsResp, err error) {
	// TODO: Your code here...
	return
}

// GetIncrementalConversation implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) GetIncrementalConversation(ctx context.Context, req *conversation.GetIncrementalConversationReq) (resp *conversation.GetIncrementalConversationResp, err error) {
	// TODO: Your code here...
	return
}

// GetOwnerConversation implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) GetOwnerConversation(ctx context.Context, req *conversation.GetOwnerConversationReq) (resp *conversation.GetOwnerConversationResp, err error) {
	// TODO: Your code here...
	return
}

// GetConversationsNeedClearMsg implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) GetConversationsNeedClearMsg(ctx context.Context, req *conversation.GetConversationsNeedClearMsgReq) (resp *conversation.GetConversationsNeedClearMsgResp, err error) {
	// TODO: Your code here...
	return
}

// GetNotNotifyConversationIDs implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) GetNotNotifyConversationIDs(ctx context.Context, req *conversation.GetNotNotifyConversationIDsReq) (resp *conversation.GetNotNotifyConversationIDsResp, err error) {
	// TODO: Your code here...
	return
}

// GetPinnedConversationIDs implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) GetPinnedConversationIDs(ctx context.Context, req *conversation.GetPinnedConversationIDsReq) (resp *conversation.GetPinnedConversationIDsResp, err error) {
	// TODO: Your code here...
	return
}

// ClearUserConversationMsg implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) ClearUserConversationMsg(ctx context.Context, req *conversation.ClearUserConversationMsgReq) (resp *conversation.ClearUserConversationMsgResp, err error) {
	// TODO: Your code here...
	return
}

// UpdateConversationsByUser implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) UpdateConversationsByUser(ctx context.Context, req *conversation.UpdateConversationsByUserReq) (resp *conversation.UpdateConversationsByUserResp, err error) {
	// TODO: Your code here...
	return
}
