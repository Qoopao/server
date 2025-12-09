package main

import (
	"context"
	sdkws "github.com/roc/roc-im-server/kitex_gen/sdkws"
)

// MessageServiceImpl implements the last service interface defined in the IDL.
type MessageServiceImpl struct{}

// BatchSendMessage implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) BatchSendMessage(ctx context.Context, req *sdkws.BatchSendMessageRequest) (resp *sdkws.BatchSendMessageResponse, err error) {
	// TODO: Your code here...
	return
}

// BatchChangeMessages implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) BatchChangeMessages(ctx context.Context, req *sdkws.BatchChangeMessagesRequest) (resp *sdkws.BatchChangeMessagesResponse, err error) {
	// TODO: Your code here...
	return
}

// FetchConvMessageList implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) FetchConvMessageList(ctx context.Context, req *sdkws.FetchConvMessageListRequest) (resp *sdkws.FetchConvMessageListResponse, err error) {
	// TODO: Your code here...
	return
}

// BatchGetMessages implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) BatchGetMessages(ctx context.Context, req *sdkws.BatchGetMessagesRequest) (resp *sdkws.BatchGetMessagesResponse, err error) {
	// TODO: Your code here...
	return
}

// BatchChangeConversations implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) BatchChangeConversations(ctx context.Context, req *sdkws.BatchChangeConversationsRequest) (resp *sdkws.BatchChangeConversationsResponse, err error) {
	// TODO: Your code here...
	return
}

// FetchUserRecentConvList implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) FetchUserRecentConvList(ctx context.Context, req *sdkws.FetchUserRecentConvListRequest) (resp *sdkws.FetchUserRecentConvListResponse, err error) {
	// TODO: Your code here...
	return
}

// UserMessageIntegrityCheck implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) UserMessageIntegrityCheck(ctx context.Context, req *sdkws.UserMessageIntegrityCheckRequest) (resp *sdkws.UserMessageIntegrityCheckResponse, err error) {
	// TODO: Your code here...
	return
}

// BatchGetConversations implements the ConversationServiceImpl interface.
func (s *ConversationServiceImpl) BatchGetConversations(ctx context.Context, req *sdkws.BatchGetConversationsRequest) (resp *sdkws.BatchGetConversationsResponse, err error) {
	// TODO: Your code here...
	return
}
