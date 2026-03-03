package consts

// 全局服务端口常量
// 所有 RPC 服务的默认监听端口统一在此定义，避免在各个 start.go 中硬编码魔法数字。
const (
	BackservicePortDefault         = 10300
	MessageServicePortDefault      = 10400
	ConversationServicePortDefault = 10500
	SequenceServicePortDefault     = 10600
)
