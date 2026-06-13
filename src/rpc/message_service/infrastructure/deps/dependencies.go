package deps

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/kitex/client"
	kitexdiscovery "github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	"github.com/rhp-QE/roc-foundation-util-go/mq"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	sequence "github.com/rhp-QE/roc-im-server/kitex_gen/sequence/sequenceservice"
	consts "github.com/rhp-QE/roc-im-server/src/const"
)

// ServiceContext 管理 message_service 的全局共享资源
//
// 职责：
// 1. 懒加载 SequenceService Kitex 客户端
// 2. 管理 MQ Producer
// 3. 管理 MongoDB 存储
type ServiceContext interface {
	// GetSequenceServiceClient 懒加载获取 SequenceService 客户端
	GetSequenceServiceClient(ctx context.Context) (sequence.Client, error)

	// GetMQProducer 获取 MQ Producer
	GetMQProducer() mq.Producer

	// GetStorage 获取 MongoDB 存储
	GetStorage() foundationstorage.Storage

	// Close 关闭所有全局资源
	Close() error
}

type serviceContextImpl struct {
	producer         mq.Producer
	storage          foundationstorage.Storage
	sequenceResolver kitexdiscovery.Resolver

	seqClientOnce sync.Once
	seqClient     sequence.Client
	seqClientErr  error
}

// NewServiceContext 创建 ServiceContext 实例
// producer: MQ 生产者，用于发送消息
// storage: MongoDB 存储，用于查询消息
// sequenceResolver: Kitex resolver，负责把 sequence-service 解析为实例列表
func NewServiceContext(producer mq.Producer, storage foundationstorage.Storage, sequenceResolver kitexdiscovery.Resolver) ServiceContext {
	return &serviceContextImpl{
		producer:         producer,
		storage:          storage,
		sequenceResolver: sequenceResolver,
	}
}

func (s *serviceContextImpl) GetMQProducer() mq.Producer {
	return s.producer
}

func (s *serviceContextImpl) GetStorage() foundationstorage.Storage {
	return s.storage
}

// GetSequenceServiceClient 懒加载创建 SequenceService 客户端
func (s *serviceContextImpl) GetSequenceServiceClient(ctx context.Context) (sequence.Client, error) {
	s.seqClientOnce.Do(func() {
		opts := []client.Option{
			client.WithSuite(tracing.NewClientSuite()),
			client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: consts.MessageServiceName}),
			client.WithRPCTimeout(10 * time.Second),
		}
		if s.sequenceResolver != nil {
			opts = append(opts, client.WithResolver(s.sequenceResolver))
		}

		clientImpl, err := sequence.NewClient(
			consts.SequenceServiceName,
			opts...,
		)
		if err != nil {
			s.seqClientErr = fmt.Errorf("failed to create sequence-service client: %w", err)
			return
		}

		klog.Infof("message_service: created sequence-service client with kitex resolver")
		s.seqClient = clientImpl
	})

	return s.seqClient, s.seqClientErr
}

// Close 关闭所有全局资源
func (s *serviceContextImpl) Close() error {
	var errs []error

	if s.producer != nil {
		if err := s.producer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close mq producer: %w", err))
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
