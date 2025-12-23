package orm

import "time"

// UserMessageStatusDocument 用户消息状态文档
// 存储每个用户对每条消息的状态（已读/未读、是否收藏、是否点赞等）
// 注意：如果消息量很大，可以考虑使用 Redis 存储，MongoDB 只存储重要状态
type UserMessageStatusDocument struct {
	ID        string    `bson:"_id"`        // 复合主键：user_id:msg_id
	UserID    string    `bson:"user_id"`    // 用户ID（用于索引和查询）
	MsgID     string    `bson:"msg_id"`     // 消息ID（用于索引和查询）
	ConvID    string    `bson:"conv_id"`    // 会话ID（用于索引和查询）
	IsRead    bool      `bson:"is_read"`    // 是否已读
	IsStarred bool      `bson:"is_starred"` // 是否收藏
	IsLiked   bool      `bson:"is_liked"`   // 是否点赞
	ReadTime  time.Time `bson:"read_time"`  // 阅读时间
	UpdatedAt time.Time `bson:"updated_at"` // 最近更新时间
	// 可以扩展其他状态：是否转发、是否删除等
}
