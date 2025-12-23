package orm

import "time"

// UserConversationSettingsDocument 用户会话个性化设置文档
// 存储每个用户对每个会话的个性化设置（置顶、免打扰、提醒等）
type UserConversationSettingsDocument struct {
	ID            string    `bson:"_id"`             // 复合主键：user_id:conv_id
	UserID        string    `bson:"user_id"`          // 用户ID（用于索引和查询）
	ConvID        string    `bson:"conv_id"`          // 会话ID（用于索引和查询）
	IsPinned      bool      `bson:"is_pinned"`        // 是否置顶
	IsMuted       bool      `bson:"is_muted"`         // 是否免打扰
	IsRemind      bool      `bson:"is_remind"`        // 是否提醒
	ReadSeq       int64     `bson:"read_seq"`         // 已读消息的最大 seq（用于判断未读数）
	ReadTime      time.Time `bson:"read_time"`        // 最后阅读时间
	PinnedTime    time.Time `bson:"pinned_time"`      // 置顶时间（用于排序）
	UpdatedAt     time.Time `bson:"updated_at"`       // 最近更新时间
	// 可以扩展其他个性化字段：背景色、备注名、是否隐藏等
}

