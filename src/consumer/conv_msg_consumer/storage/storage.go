package storage

import (
	"context"
	"errors"

	"github.com/cloudwego/kitex/pkg/klog"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	servicecontext "github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/service_context"
)

const (
	// collectionMessages 存储消息详情（KV：msg_id -> msg.pb），带索引 conv_id/seq/time
	collectionMessages = "messages"
	// collectionConversations 存储会话详情快照（KV：conv_id -> last_msg.pb 等）
	collectionConversations = "conversations"
)

// ConvMsgStorage 会话消息相关的持久化接口
type ConvMsgStorage interface {
	// SaveMessage 保存一条消息（包含 conv_id/seq 信息，支持按 seq 查询单链）
	SaveMessage(ctx context.Context, msgID string, msg *sdkws.MessageData) error
}

type convMsgStorageImpl struct {
	serviceCtx servicecontext.ServiceContext
}

// NewConvMsgStorage 创建 ConvMsgStorage 实例
// 注意：底层使用 MongoDB（通过 foundation storage 接入）
func NewConvMsgStorage(serviceCtx servicecontext.ServiceContext) ConvMsgStorage {
	return &convMsgStorageImpl{
		serviceCtx: serviceCtx,
	}
}

// messageDocument MongoDB 中的消息详情文档
type messageDocument struct {
	ID     string `bson:"_id"`     // msg_id（主键，KV 语义）
	ConvID string `bson:"conv_id"` // 会话ID
	Seq    int64  `bson:"seq"`     // 序列号 / order_index（用于链路排序）
	Data   []byte `bson:"data"`    // protobuf 二进制
}

func (s *convMsgStorageImpl) getStore() foundationstorage.Storage {
	return s.serviceCtx.GetStorage()
}

// SaveMessage 保存消息详情（如果已存在则忽略重复错误）
// 说明：
//   - 每条消息一条文档，包含 conv_id / seq 等字段
//   - 会话单链通过 {conv_id, seq} 索引高效查询
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

	doc := &messageDocument{
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
