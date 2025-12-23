package orm

import "time"

// ConversationDocument MongoDB 中的会话详情文档（按 conv 维度，一条会话一条文档）
type ConversationDocument struct {
	ID             string    `bson:"_id"`              // conv_id（主键）
	Data           []byte    `bson:"data"`             // ConversationData 的 protobuf 二进制数据
	LastMessageSeq int64     `bson:"last_message_seq"` // 最后一条消息的 seq（用于索引和查询）
	UpdatedAt      time.Time `bson:"updated_at"`       // 最近更新时间
}

// UserRecentConversationsDocument MongoDB 中的用户最近会话链文档
// 一条文档只表示「一个用户在某个会话上的最近活跃信息」
// 通过 user_id + updated_at 建索引，客户端按时间范围增量拉取
type UserRecentConversationsDocument struct {
	ID             string    `bson:"_id"`              // 复合主键：user_id:conv_id
	UserID         string    `bson:"user_id"`          // 用户ID
	ConvID         string    `bson:"conv_id"`          // 会话ID
	LastMessageSeq int64     `bson:"last_message_seq"` // 该用户在该会话上的最后一条消息 seq（用于未读数计算）
	Version        int64     `bson:"version"`          // 版本号（单调递增，用于并发保护）
	UpdatedAt      time.Time `bson:"updated_at"`       // 最近活跃时间（用于增量拉取和排序）
}
