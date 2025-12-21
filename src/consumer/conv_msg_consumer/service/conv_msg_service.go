package service

import (
	"context"
	"fmt"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/google/uuid"
	"github.com/openimsdk/tools/errs"
	foundationmq "github.com/rhp-QE/roc-foundation-util-go/mq"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	convstorage "github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/storage"
)

// ConvMsgConsumerService 会话消息消费者业务接口
type ConvMsgConsumerService interface {
	// HandleMessage 处理从 MQ 拉取到的一条消息
	HandleMessage(ctx context.Context, msg *foundationmq.Message) error
}

type convMsgConsumerServiceImpl struct {
	storage convstorage.ConvMsgStorage
}

// NewConvMsgConsumerService 创建 ConvMsgConsumerService 实例
func NewConvMsgConsumerService(storage convstorage.ConvMsgStorage) ConvMsgConsumerService {
	return &convMsgConsumerServiceImpl{
		storage: storage,
	}
}

// HandleMessage 处理 MQ 消息：解析消息体 → 写会话单链 → 写消息详情 & 会话详情
func (s *convMsgConsumerServiceImpl) HandleMessage(ctx context.Context, mqMsg *foundationmq.Message) error {
	if mqMsg == nil || len(mqMsg.Value) == 0 {
		return nil
	}

	var msg sdkws.MessageData
	if err := msg.Unmarshal(mqMsg.Value); err != nil {
		klog.CtxErrorf(ctx, "[ConvMsgConsumer] unmarshal message failed",
			"error", err.Error())
		return errs.ErrInternalServer.WrapMsg(fmt.Sprintf("unmarshal message failed: %v", err))
	}

	if msg.ConvID == "" {
		klog.CtxErrorf(ctx, "[ConvMsgConsumer] empty convID, skip message",
			"raw_key", mqMsg.Key)
		return nil
	}

	// 生成/选择消息ID：优先 server_msg_id，其次 client_msg_id，最后 convID:seq
	msgID := msg.SMessageID
	if msgID == "" {
		msgID = msg.CMessaegID
	}
	if msgID == "" {
		if msg.Seq > 0 {
			msgID = fmt.Sprintf("%s:%d", msg.ConvID, msg.Seq)
		} else {
			msgID = uuid.New().String()
		}
	}

	// 1. 保存消息详情（一条消息一条文档，包含 conv_id/seq）
	if err := s.storage.SaveMessage(ctx, msgID, &msg); err != nil {
		return err
	}

	klog.CtxDebugf(ctx, "[ConvMsgConsumer] message processed successfully",
		"conv_id", msg.ConvID,
		"msg_id", msgID,
		"seq", msg.Seq)
	return nil
}
