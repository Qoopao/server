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
	backbonservice "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/backbon/backbonservice"
	foundationmq "github.com/rhp-QE/roc-foundation-util-go/mq"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	sequence "github.com/rhp-QE/roc-im-server/kitex_gen/sequence/sequenceservice"
	consts "github.com/rhp-QE/roc-im-server/src/const"
)

// ServiceContext 管理 conv_msg_consumer 的全局共享资源。
// 下游 RPC 的服务发现和负载均衡统一使用 Kitex resolver/client。
type ServiceContext interface {
	// GetStorage 获取底层存储（MongoDB）
	GetStorage() foundationstorage.Storage

	// GetConsumer 获取 MQ Consumer
	GetConsumer() foundationmq.Consumer

	// GetBackbonServiceClient 获取 backbon 服务客户端（懒加载）
	GetBackbonServiceClient(ctx context.Context) (backbonservice.Client, error)

	// GetSequenceServiceClient 获取 sequence-service 客户端（懒加载）
	GetSequenceServiceClient(ctx context.Context) (sequence.Client, error)

	// Close 关闭所有资源
	Close() error
}

type serviceContextImpl struct {
	storage  foundationstorage.Storage
	consumer foundationmq.Consumer
	resolver kitexdiscovery.Resolver

	backbonClient sync.Map // serviceName -> *clientEntry
	seqClient     sequence.Client
	seqClientOnce sync.Once
	seqClientErr  error
}

// NewServiceContext 创建 ServiceContext 实例
func NewServiceContext(storage foundationstorage.Storage, consumer foundationmq.Consumer, resolver kitexdiscovery.Resolver) ServiceContext {
	return &serviceContextImpl{
		storage:  storage,
		consumer: consumer,
		resolver: resolver,
	}
}

func (s *serviceContextImpl) GetStorage() foundationstorage.Storage {
	return s.storage
}

func (s *serviceContextImpl) GetConsumer() foundationmq.Consumer {
	return s.consumer
}

// GetBackbonServiceClient 获取 backbon 服务客户端（懒加载）
func (s *serviceContextImpl) GetBackbonServiceClient(_ context.Context) (backbonservice.Client, error) {
	serviceName := consts.BackbonServiceName

	entryInterface, _ := s.backbonClient.LoadOrStore(serviceName, &clientEntry{})
	entry := entryInterface.(*clientEntry)

	var err error
	entry.once.Do(func() {
		clientImpl, createErr := backbonservice.NewClient(serviceName, s.clientOptions()...)
		if createErr != nil {
			err = fmt.Errorf("failed to create backbon-service kitex client: %w", createErr)
			return
		}

		entry.client = clientImpl
		klog.Infof("conv_msg_consumer: cached kitex client, service=%s", serviceName)
	})

	if err != nil {
		s.backbonClient.Delete(serviceName)
		return nil, err
	}

	return entry.client.(backbonservice.Client), nil
}

// GetSequenceServiceClient 懒加载创建 SequenceService 客户端
func (s *serviceContextImpl) GetSequenceServiceClient(_ context.Context) (sequence.Client, error) {
	s.seqClientOnce.Do(func() {
		clientImpl, err := sequence.NewClient(
			consts.SequenceServiceName,
			s.clientOptions()...,
		)
		if err != nil {
			s.seqClientErr = fmt.Errorf("failed to create sequence-service kitex client: %w", err)
			return
		}

		klog.Infof("conv_msg_consumer: cached kitex client, service=%s", consts.SequenceServiceName)
		s.seqClient = clientImpl
	})

	return s.seqClient, s.seqClientErr
}

func (s *serviceContextImpl) clientOptions() []client.Option {
	opts := []client.Option{
		client.WithSuite(tracing.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: consts.ConvMsgConsumerName}),
		client.WithRPCTimeout(10 * time.Second),
	}
	if s.resolver != nil {
		opts = append(opts, client.WithResolver(s.resolver))
	}
	return opts
}

// clientEntry 客户端条目，包含客户端和创建同步原语
type clientEntry struct {
	client interface{}
	once   sync.Once
}

// Close 按顺序关闭所有资源
func (s *serviceContextImpl) Close() error {
	var errs []error

	s.backbonClient.Range(func(key, value interface{}) bool {
		serviceName := key.(string)
		entry := value.(*clientEntry)
		if entry.client != nil {
			if closer, ok := entry.client.(interface{ Close() error }); ok {
				if err := closer.Close(); err != nil {
					errs = append(errs, fmt.Errorf("failed to close client for service %s: %w", serviceName, err))
				}
			}
		}
		return true
	})

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
