package application

import (
	"context"

	"github.com/rhp-QE/roc-im-server/kitex_gen/sdkws"
	consts "github.com/rhp-QE/roc-im-server/src/const"
	deps "github.com/rhp-QE/roc-im-server/src/rpc/message_service/infrastructure/deps"
	persistence "github.com/rhp-QE/roc-im-server/src/rpc/message_service/infrastructure/persistence"
)

// MessageService 定义消息发送链路的应用用例接口。
type MessageService interface {
	// BatchSendMessage 批量发送消息，完成幂等、seq 分配、消息事实持久化和 MQ 投递。
	BatchSendMessage(ctx context.Context, req *sdkws.BatchSendMessageRequest) (*sdkws.BatchSendMessageResponse, error)

	// FetchConvMessageList 查询会话消息列表
	FetchConvMessageList(ctx context.Context, req *sdkws.FetchConvMessageListRequest) (*sdkws.FetchConvMessageListResponse, error)

	// BatchGetMessages 根据消息ID列表批量获取消息详情
	BatchGetMessages(ctx context.Context, req *sdkws.BatchGetMessagesRequest) (*sdkws.BatchGetMessagesResponse, error)
}

// messageServiceImpl 是 MessageService 的具体实现
type messageServiceImpl struct {
	messageRepo     persistence.MessageRepository
	serviceCtx      deps.ServiceContext
	sequenceSvcName string
}

// NewMessageService 创建 MessageService 实例
func NewMessageService(messageRepo persistence.MessageRepository, serviceCtx deps.ServiceContext) MessageService {
	return &messageServiceImpl{
		messageRepo:     messageRepo,
		serviceCtx:      serviceCtx,
		sequenceSvcName: consts.SequenceServiceName,
	}
}
