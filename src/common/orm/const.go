package orm

const (
	// DBMessage 消息 & 会话相关数据所在的数据库（message_service / conv_msg_consumer / conversation_service）
	DBMessage = "im_message"

	// CollectionConversations 会话详情集合名称
	CollectionConversations = "conversations"
	// CollectionUserRecentConversations 用户最近会话链集合名称
	CollectionUserRecentConversations = "user_recent_conversations"
	// CollectionMessages 消息集合名称
	CollectionMessages = "messages"
)
