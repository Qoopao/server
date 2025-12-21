package orm

// ConvType 会话类型枚举
type ConvType int32

const (
	// ConvTypeSingleChat 单聊
	ConvTypeSingleChat ConvType = 1
	// ConvTypeGroupChat 群聊
	ConvTypeGroupChat ConvType = 2
	// ConvTypeSystemNotification 系统通知
	ConvTypeSystemNotification ConvType = 3
)

