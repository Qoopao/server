package util

import (
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
)

// IsSingleChat 判断是否为单聊
func IsSingleChat(msg *sdkws.MessageData) bool {
	return msg != nil && msg.ConvType == int32(orm.ConvTypeSingleChat)
}

// IsGroupChat 判断是否为群聊
func IsGroupChat(msg *sdkws.MessageData) bool {
	return msg != nil && msg.ConvType == int32(orm.ConvTypeGroupChat)
}

// IsSystemNotification 判断是否为系统通知
func IsSystemNotification(msg *sdkws.MessageData) bool {
	return msg != nil && msg.ConvType == int32(orm.ConvTypeSystemNotification)
}

// GetMessageMembers 获取消息的会话成员（单聊时返回发送者和接收者）
func GetMessageMembers(msg *sdkws.MessageData) []string {
	if msg == nil {
		return nil
	}
	members := make([]string, 0, 2)
	if msg.SendID != "" {
		members = append(members, msg.SendID)
	}
	if msg.RecvID != "" {
		members = append(members, msg.RecvID)
	}
	return members
}

