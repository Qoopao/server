package servicecontext

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	"github.com/rhp-QE/roc-foundation-util-go/mq"
	"github.com/rhp-QE/roc-foundation-util-go/service_registry/discovery"
	"github.com/rhp-QE/roc-foundation-util-go/service_registry/loadbalancer"
	foundationregistry "github.com/rhp-QE/roc-foundation-util-go/service_registry/registry"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	sequence "github.com/rhp-QE/roc-im-server/kitex_gen/sequence/sequenceservice"
	consts "github.com/rhp-QE/roc-im-server/src/const"
)

// ServiceContext 管理 message_service 的全局共享资源
//
// 职责：
// 1. 管理服务注册中心（Registry）
// 2. 管理服务发现客户端（Discovery）
// 3. 懒加载 SequenceService 客户端
// 4. 管理 MQ Producer
// 5. 管理 MongoDB 存储
type ServiceContext interface {
	// GetRegistry 获取服务注册中心
	GetRegistry() foundationregistry.Registry

	// GetDiscovery 获取服务发现客户端
	GetDiscovery() discovery.Discovery

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
	registry  foundationregistry.Registry
	discovery discovery.Discovery
	producer  mq.Producer
	storage   foundationstorage.Storage

	seqClientOnce sync.Once
	seqClient     sequence.Client
	seqClientErr  error
}

// NewServiceContext 创建 ServiceContext 实例
// registry: 服务注册中心，用于服务发现
// producer: MQ 生产者，用于发送消息
// storage: MongoDB 存储，用于查询消息
func NewServiceContext(registry foundationregistry.Registry, producer mq.Producer, storage foundationstorage.Storage) ServiceContext {
	// 创建服务发现客户端，默认使用轮询负载均衡
	lb := loadbalancer.NewRoundRobinLoadBalancer()
	disc := discovery.NewDiscovery(registry, lb)

	return &serviceContextImpl{
		registry:  registry,
		discovery: disc,
		producer:  producer,
		storage:   storage,
	}
}

func (s *serviceContextImpl) GetRegistry() foundationregistry.Registry {
	return s.registry
}

func (s *serviceContextImpl) GetDiscovery() discovery.Discovery {
	return s.discovery
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
		// 这里使用 3 秒超时，避免长时间阻塞
		discoveryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		instance, err := s.discovery.GetInstance(discoveryCtx, consts.SequenceServiceName)
		if err != nil {
			s.seqClientErr = fmt.Errorf("failed to discover sequence-service: %w", err)
			return
		}

		target := fmt.Sprintf("%s:%d", instance.Host, instance.Port)
		clientImpl, err := sequence.NewClient(
			consts.SequenceServiceName, // 与服务端 WithServerBasicInfo 中设置的服务名称保持一致
			client.WithHostPorts(target),
			client.WithSuite(tracing.NewClientSuite()),
			client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: consts.SequenceServiceName}),
			client.WithRPCTimeout(10*time.Second), // 设置 RPC 超时时间为 10 秒
		)
		if err != nil {
			s.seqClientErr = fmt.Errorf("failed to create sequence-service client: %w", err)
			return
		}

		klog.Infof("message_service: created sequence-service client, target=%s", target)
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

	if s.discovery != nil {
		if err := s.discovery.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close discovery: %w", err))
		}
	}

	if s.storage != nil {
		if err := s.storage.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close storage: %w", err))
		}
	}

	if s.registry != nil {
		if err := s.registry.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close registry: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("close service context error: %v", errs)
	}

	return nil
}
