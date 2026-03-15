package consts

// CmdMessage 命令 ID（与 idl/sdkws.proto CmdMessage 注释对齐）
const (
	// 会话类：2xxx
	CmdIDConversationCreated = 2000 // 会话创建（创建群聊后下推）
	CmdIDConversationDeleted = 2001 // 会话状态发生改变（删除）
	CmdIDInvitedToGroup      = 3001 // 被邀请进群
)
