package service

import (
	"fmt"
	"strconv"
	"strings"
)

// SDKWSMethod SDK WebSocket 方法枚举（对应 C++ SDKWSMethod）
type SDKWSMethod int32

const (
	SDKWSMethodSendMessage               SDKWSMethod = 101 // 发送消息
	SDKWSMethodPullSingleList            SDKWSMethod = 102 // 拉取单链
	SDKWSMethodPullMixList               SDKWSMethod = 103 // 拉取混链
	SDKWSMethodPushUserMessage           SDKWSMethod = 104 // 下推用户消息
	SDKWSMethodPushCmdMessage            SDKWSMethod = 105 // 下推命令消息
	SDKWSMethodUserMessageIntegrityCheck SDKWSMethod = 106 // 混链拉取会话完整性校验
	SDKWSMethodMessageChange             SDKWSMethod = 107 // 消息改变 请求
	SDKWSMethodConversationChange        SDKWSMethod = 108 // 会话改变 请求
)

// ParseSDKWSMethod 将字符串 method 解析为 SDKWSMethod 枚举
// 支持数字字符串（如 "101"）和枚举名称（如 "SEND_MESSAGE"）
func ParseSDKWSMethod(methodStr string) (SDKWSMethod, error) {
	// 尝试解析为数字
	if methodNum, err := strconv.ParseInt(methodStr, 10, 32); err == nil {
		method := SDKWSMethod(methodNum)
		if method >= SDKWSMethodSendMessage && method <= SDKWSMethodConversationChange {
			return method, nil
		}
		return 0, fmt.Errorf("invalid SDKWSMethod number: %d", methodNum)
	}

	// 尝试解析为枚举名称（不区分大小写）
	methodUpper := strings.ToUpper(methodStr)
	switch methodUpper {
	case "SEND_MESSAGE", "101":
		return SDKWSMethodSendMessage, nil
	case "PULL_SINGLE_LIST", "102":
		return SDKWSMethodPullSingleList, nil
	case "PULL_MIX_LIST", "103":
		return SDKWSMethodPullMixList, nil
	case "PUSH_USER_MESSAGE", "104":
		return SDKWSMethodPushUserMessage, nil
	case "PUSH_CMD_MESSAGE", "105":
		return SDKWSMethodPushCmdMessage, nil
	case "USER_MESSGAGE_INTEGRITY_CHECK", "USER_MESSAGE_INTEGRITY_CHECK", "106":
		return SDKWSMethodUserMessageIntegrityCheck, nil
	case "MESSAGE_CHANGE", "107":
		return SDKWSMethodMessageChange, nil
	case "CONVERSATION_CHANGE", "108":
		return SDKWSMethodConversationChange, nil
	default:
		return 0, fmt.Errorf("unknown SDKWSMethod: %s", methodStr)
	}
}
