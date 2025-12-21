package storage

import (
	"context"
	"errors"

	"github.com/cloudwego/kitex/pkg/klog"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
)

// SaveMessage 保存消息详情（如果已存在则忽略重复错误）
// 说明：
//   - 每条消息一条文档，包含 conv_id / seq 等字段
//   - 会话单链通过 {conv_id, seq} 索引高效查询
//   - 索引在启动时创建，见 start.go:createMongoStorage()
func (s *convMsgStorageImpl) SaveMessage(ctx context.Context, msgID string, msg *sdkws.MessageData) error {
	if msgID == "" || msg == nil {
		return nil
	}

	data, err := msg.Marshal(nil)
	if err != nil {
		klog.CtxErrorf(ctx, "[ConvMsgStorage] marshal msg failed",
			"msg_id", msgID,
			"conv_id", msg.ConvID,
			"error", err.Error())
		return err
	}

	doc := &orm.MessageDocument{
		ID:     msgID,
		ConvID: msg.ConvID,
		Seq:    msg.Seq,
		Data:   data,
	}

	_, err = s.getStore().InsertOne(ctx, collectionMessages, doc)
	if err != nil {
		// 已存在则视为成功（幂等）
		if errors.Is(err, foundationstorage.ErrDuplicateKey) {
			return nil
		}
		klog.CtxErrorf(ctx, "[ConvMsgStorage] save message detail failed",
			"msg_id", msgID,
			"conv_id", msg.ConvID,
			"error", err.Error())
		return err
	}

	klog.CtxDebugf(ctx, "[ConvMsgStorage] save message detail success",
		"msg_id", msgID,
		"conv_id", msg.ConvID)
	return nil
}
