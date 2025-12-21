package servicecontext

import (
	"fmt"

	foundationmq "github.com/rhp-QE/roc-foundation-util-go/mq"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
)

// ServiceContext 管理 conv_msg_consumer 的全局共享资源
// 仅包含：消息队列 Consumer 和 MongoDB Storage
type ServiceContext interface {
	// GetStorage 获取底层存储（MongoDB）
	GetStorage() foundationstorage.Storage

	// GetConsumer 获取 MQ Consumer
	GetConsumer() foundationmq.Consumer

	// Close 关闭所有资源
	Close() error
}

type serviceContextImpl struct {
	storage  foundationstorage.Storage
	consumer foundationmq.Consumer
}

// NewServiceContext 创建 ServiceContext 实例
func NewServiceContext(storage foundationstorage.Storage, consumer foundationmq.Consumer) ServiceContext {
	return &serviceContextImpl{
		storage:  storage,
		consumer: consumer,
	}
}

func (s *serviceContextImpl) GetStorage() foundationstorage.Storage {
	return s.storage
}

func (s *serviceContextImpl) GetConsumer() foundationmq.Consumer {
	return s.consumer
}

// Close 按顺序关闭 MQ Consumer 和 Storage
func (s *serviceContextImpl) Close() error {
	var errs []error

	if s.consumer != nil {
		if err := s.consumer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close mq consumer: %w", err))
		}
	}

	if s.storage != nil {
		if err := s.storage.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close storage: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("close service context error: %v", errs)
	}
	return nil
}


