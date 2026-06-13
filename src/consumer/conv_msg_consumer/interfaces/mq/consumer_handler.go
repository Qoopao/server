package mq

import (
	"context"
	"fmt"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/openimsdk/tools/errs"
	foundationmq "github.com/rhp-QE/roc-foundation-util-go/mq"
	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	"github.com/rhp-QE/roc-im-server/src/consumer/conv_msg_consumer/application"
)

// ConsumerHandler 会话消息 MQ 入口。
type ConsumerHandler interface {
	// Subscribe 订阅消息队列并处理消息
	Subscribe(ctx context.Context, consumer foundationmq.Consumer) error
}

type consumerHandler struct {
	usecase application.ConsumerUsecase
}

// NewConsumerHandler 创建 ConsumerHandler 实例
func NewConsumerHandler(usecase application.ConsumerUsecase) ConsumerHandler {
	return &consumerHandler{
		usecase: usecase,
	}
}

// Subscribe 订阅消息队列并处理消息
func (a *consumerHandler) Subscribe(ctx context.Context, consumer foundationmq.Consumer) error {
	handler := func(mqMsg *foundationmq.Message) error {
		// 参数校验：检查消息是否为空
		if mqMsg == nil || len(mqMsg.Value) == 0 {
			return nil
		}

		// 反序列化消息
		var msg sdkws.MessageData
		if err := msg.Unmarshal(mqMsg.Value); err != nil {
			klog.CtxErrorf(ctx, "[ConsumerHandler] unmarshal message failed",
				"error", err.Error())
			return errs.ErrInternalServer.WrapMsg(fmt.Sprintf("unmarshal message failed: %v", err))
		}

		// 参数校验：检查必需字段
		if msg.ConvID == "" {
			klog.CtxErrorf(ctx, "[ConsumerHandler] empty convID, skip message",
				"raw_key", mqMsg.Key)
			return nil
		}

		if msg.SMessageID == "" {
			klog.CtxErrorf(ctx, "[ConsumerHandler] msgID is empty, skip message",
				"conv_id", msg.ConvID,
				"seq", msg.Seq,
				"raw_key", mqMsg.Key)
			return errs.ErrArgs.WrapMsg("msgID is required")
		}

		// 调用 application 层处理消息投影与推送。
		return a.usecase.HandleMessage(ctx, &msg)
	}

	return consumer.Subscribe(ctx, handler)
}
