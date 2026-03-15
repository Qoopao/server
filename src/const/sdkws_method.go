package consts

// SDK WebSocket 方法「字符串编码」（与 C++ 端 SDKWSMethod 数值一致，但在 Go 内统一使用字符串）
// 注意：这里维护的是数字的字符串形式，例如 "104"，便于直接用于网络 method 字段。
const (
	SDKWSMethodSendMessage               = "101" // 发送消息
	SDKWSMethodPullSingleList            = "102" // 拉取单链
	SDKWSMethodPullMixList               = "103" // 拉取混链
	SDKWSMethodPushUserMessage           = "104" // 下推用户消息
	SDKWSMethodPushCmdMessage            = "105" // 下推命令消息
	SDKWSMethodUserMessageIntegrityCheck = "106" // 混链拉取会话完整性校验
	SDKWSMethodMessageChange             = "107" // 消息改变 请求
	SDKWSMethodConversationChange        = "108" // 会话改变 请求
	SDKWSMethodCreateGroup               = "109" // 创建群聊
	SDKWSMethodInviteGroupMembers        = "110" // 邀请进群
)
