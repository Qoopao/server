package orm

import "time"

// ConversationDocument MongoDB 中的会话详情文档
type ConversationDocument struct {
	ID             string    `bson:"_id"`              // conv_id（主键）
	Data           []byte    `bson:"data"`             // ConversationData 的 protobuf 二进制数据
	LastMessageSeq int64     `bson:"last_message_seq"` // 最后一条消息的 seq（用于索引和查询）
	UpdatedAt      time.Time `bson:"updated_at"`       // 最近更新时间
}

// ConversationItem 用户最近会话链中的会话项
type ConversationItem struct {
	ConvID    string    `bson:"conv_id"`
	LastSeq   int64     `bson:"last_seq"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// UserRecentConversationsDocument MongoDB 中的用户最近会话链文档
type UserRecentConversationsDocument struct {
	ID            string             `bson:"_id"`           // user_id（主键）
	Conversations []ConversationItem `bson:"conversations"` // 有序数组，最前面的是最近活跃的
}

