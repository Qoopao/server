package util

import (
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
)

// IsSingleChatConv 判断会话是否为单聊
func IsSingleChatConv(conv *sdkws.ConversationData) bool {
	return conv != nil && conv.ConvType == int32(orm.ConvTypeSingleChat)
}

// IsGroupChatConv 判断会话是否为群聊
func IsGroupChatConv(conv *sdkws.ConversationData) bool {
	return conv != nil && conv.ConvType == int32(orm.ConvTypeGroupChat)
}

// IsSystemNotificationConv 判断会话是否为系统通知
func IsSystemNotificationConv(conv *sdkws.ConversationData) bool {
	return conv != nil && conv.ConvType == int32(orm.ConvTypeSystemNotification)
}

