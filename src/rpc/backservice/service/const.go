package service

// 消息服务方法名常量
const (
	MethodBatchSendMessage     = "BatchSendMessage"
	MethodBatchChangeMessages  = "BatchChangeMessages"
	MethodFetchConvMessageList = "FetchConvMessageList"
	MethodBatchGetMessages     = "BatchGetMessages"
)

// 会话服务方法名常量
const (
	MethodBatchChangeConversations  = "BatchChangeConversations"
	MethodFetchUserRecentConvList   = "FetchUserRecentConvList"
	MethodUserMessageIntegrityCheck = "UserMessageIntegrityCheck"
	MethodBatchGetConversations     = "BatchGetConversations"
)
