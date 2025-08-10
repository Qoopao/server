// Copyright © 2023 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"
	"errors"

	"github.com/roc/roc-im-server/internal/kitex_gen/conversation"
	"github.com/roc/roc-im-server/internal/kitex_gen/sdkws"
	"github.com/roc/roc-im-server/tools/kvstore"
	"github.com/roc/roc-im-server/tools/math"
	"github.com/roc/roc-im-server/tools/mq"
)

const (
	updateKeyMsg = iota
	updateKeyRevoke
)

// CommonMsgDatabase defines the interface for message database operations.
type CommonMsgDatabase interface {
	MsgToMQ(ctx context.Context, key string, msgID string) error

	SaveMsgInfo(ctx context.Context, msg *sdkws.MsgData) error
	GetMsgInfo(ctx context.Context, messageID string) (*sdkws.MsgData, error)

	AppendMsgToConvMsgList(ctx context.Context, conversationID string, msgID string) (int64, error)
	GetConvMessageList(ctx context.Context, conversationID string, cursor int64, limit int64, forward bool) ([]*sdkws.MsgData, bool, error)

	UpdateUserConvList(ctx context.Context, userID string, conversationID string) error
	GetUserConvList(ctx context.Context, userID string, cursor int64, limit int64, forward bool) ([]string, bool, error)

	SaveConversationInfo(ctx context.Context, conversationID string, conversationInfo *conversation.ConversationInfo) error
	GetConversationInfo(ctx context.Context, conversationID string) (*conversation.ConversationInfo, error)
}

func NewCommonMsgDatabase(mqi mq.MQ, kvstore kvstore.KVStore) CommonMsgDatabase {
	return &commonMsgDatabase{

		mqi:     mqi,
		kvstore: kvstore,
	}
}

type commonMsgDatabase struct {
	mqi     mq.MQ
	kvstore kvstore.KVStore
}

func (db *commonMsgDatabase) MsgToMQ(ctx context.Context, key string, msgID string) error {
	err := db.mqi.Publish(ctx, &mq.Message{
		Topic: "message_topic",
		Body:  []byte(msgID),
	})

	return err
}

func (db *commonMsgDatabase) SaveMsgInfo(ctx context.Context, msg *sdkws.MsgData) error {
	var (
		data []byte
		err  error
	)

	if data, err = msg.Marshal(nil); err != nil {
		return err
	}

	db.kvstore.Set(ctx, keyForMsgInfo(msg.ServerMsgID), data, 0)
	return nil
}

func (db *commonMsgDatabase) AppendMsgToConvMsgList(ctx context.Context, conversationID string, msgID string) (int64, error) {
	if conversationID == "" {
		return 0, errors.New("conversationID is empty")
	}
	return db.kvstore.RPush(ctx, keyForConvMessageList(conversationID), []byte(msgID))
}

func (db *commonMsgDatabase) GetMsgInfo(ctx context.Context, messageID string) (*sdkws.MsgData, error) {
	var (
		data []byte
		err  error
	)

	if data, err = db.kvstore.Get(ctx, keyForMsgInfo(messageID)); err != nil {
		return nil, err
	}

	message := &sdkws.MsgData{}
	if err = message.Unmarshal(data); err != nil {
		return nil, err
	}

	return message, nil
}

func (db *commonMsgDatabase) UpdateUserConvList(ctx context.Context, userID string, conversationID string) error {
	if userID == "" {
		return errors.New("userID is empty")
	}
	// 删除
	if err := db.kvstore.LRem(ctx, keyForUserConvList(userID), 1, []byte(conversationID)); err != nil {
		return err
	}
	// 添加到最新
	_, err := db.kvstore.RPush(ctx, keyForUserConvList(userID), []byte(conversationID))
	return err
}

func (db *commonMsgDatabase) GetConvMessageList(ctx context.Context, conversationID string, cursor int64, limit int64, forward bool) ([]*sdkws.MsgData, bool, error) {
	// todo 参数合法性校验和修正
	var (
		start   int64
		stop    int64
		msgData *sdkws.MsgData
	)

	length, err := db.kvstore.LLen(ctx, keyForConvMessageList(conversationID))
	if err != nil {
		return nil, false, err
	}

	// 修正区间
	start, stop = modifyRange(cursor, limit, forward, length)

	msgList, err := db.kvstore.LRange(ctx, keyForConvMessageList(conversationID), start, stop)
	if err != nil {
		return nil, false, err
	}

	msgDataList := make([]*sdkws.MsgData, 0)
	for _, msgID := range msgList {
		msgData, err = db.GetMsgInfo(ctx, string(msgID))
		if err != nil {
			msgData = &sdkws.MsgData{
				ServerMsgID: string(msgID),
			}
		}
		msgDataList = append(msgDataList, msgData)
	}

	return msgDataList, true, nil
}

func (db *commonMsgDatabase) GetUserConvList(ctx context.Context, userID string, cursor int64, limit int64, forward bool) ([]string, bool, error) {
	var (
		start int64
		stop  int64
	)

	length, err := db.kvstore.LLen(ctx, keyForUserConvList(userID))
	if err != nil {
		return nil, false, err
	}

	// 修正区间
	start, stop = modifyRange(cursor, limit, forward, length)

	convList, err := db.kvstore.LRange(ctx, keyForUserConvList(userID), start, stop)
	if err != nil {
		return nil, false, err
	}

	convStrList := make([]string, 0)
	for _, conv := range convList {
		convStrList = append(convStrList, string(conv))
	}

	return convStrList, true, nil
}

func (db *commonMsgDatabase) SaveConversationInfo(ctx context.Context, conversationID string, conversationInfo *conversation.ConversationInfo) error {
	data, err := conversationInfo.Marshal(nil)
	if err != nil {
		return err
	}
	db.kvstore.Set(ctx, keyForConvInfo(conversationID), data, 0)
	return nil
}

func (db *commonMsgDatabase) GetConversationInfo(ctx context.Context, conversationID string) (*conversation.ConversationInfo, error) {
	var (
		data []byte
		err  error
	)

	if data, err = db.kvstore.Get(ctx, keyForConvInfo(conversationID)); err != nil {
		return nil, err
	}

	conversationInfo := &conversation.ConversationInfo{}
	if err = conversationInfo.Unmarshal(data); err != nil {
		return nil, err
	}

	return conversationInfo, nil
}

func keyForUserConvList(uid string) string {
	return "user_conv_list:" + uid
}

func keyForConvMessageList(cid string) string {
	return "conv_msg_list:" + cid
}

func keyForConvInfo(cid string) string {
	return "conv_info:" + cid
}

func keyForMsgInfo(msgid string) string {
	return "msg_info:" + msgid
}

func modifyRange(cursor int64, limit int64, forward bool, length int64) (int64, int64) {
	var (
		start int64
		stop  int64
	)

	if cursor < 0 {
		return math.Max(0, length-limit), length - 1
	}

	if forward {
		start = math.Max(0, cursor-limit+1)
		stop = cursor
	} else {
		start = cursor
		stop = math.Min(length-1, cursor+limit-1)
	}

	return start, stop
}
