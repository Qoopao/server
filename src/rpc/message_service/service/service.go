package service

import (
	"context"

	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	servicecontext "github.com/rhp-QE/roc-im-server/src/rpc/message_service/service_context"
	"github.com/rhp-QE/roc-im-server/src/rpc/message_service/storage"
)

// MessageService 定义消息发送链路的业务接口（逻辑层）
type MessageService interface {
	// BatchSendMessage 批量发送消息：为每条消息生成顺序号并投递到 MQ
	BatchSendMessage(ctx context.Context, req *sdkws.BatchSendMessageRequest) (*sdkws.BatchSendMessageResponse, error)
}

// messageServiceImpl 是 MessageService 的具体实现
type messageServiceImpl struct {
	storage         storage.MessageStorage
	serviceCtx      servicecontext.ServiceContext
	sequenceSvcName string
}

// NewMessageService 创建 MessageService 实例
func NewMessageService(storage storage.MessageStorage, serviceCtx servicecontext.ServiceContext) MessageService {
	return &messageServiceImpl{
		storage:         storage,
		serviceCtx:      serviceCtx,
		sequenceSvcName: "sequence-service",
	}
}
