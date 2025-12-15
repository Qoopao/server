package servicecontext

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rhp-QE/roc-foundation-util-go/mq"
	"github.com/rhp-QE/roc-foundation-util-go/service_registry/discovery"
	"github.com/rhp-QE/roc-foundation-util-go/service_registry/loadbalancer"
	foundationregistry "github.com/rhp-QE/roc-foundation-util-go/service_registry/registry"
	sequence "github.com/rhp-QE/roc-im-server/kitex_gen/sequence/sequenceservice"
)

// ServiceContext 管理 message_service 的全局共享资源
//
// 职责：
// 1. 管理服务注册中心（Registry）
// 2. 管理服务发现客户端（Discovery）
// 3. 懒加载 SequenceService 客户端
// 4. 管理 MQ Producer
type ServiceContext interface {
	// GetRegistry 获取服务注册中心
	GetRegistry() foundationregistry.Registry

	// GetDiscovery 获取服务发现客户端
	GetDiscovery() discovery.Discovery

	// GetSequenceServiceClient 懒加载获取 SequenceService 客户端
	GetSequenceServiceClient(ctx context.Context) (sequence.Client, error)

	// GetMQProducer 获取 MQ Producer
	GetMQProducer() mq.Producer

	// Close 关闭所有全局资源
	Close() error
}

type serviceContextImpl struct {
	registry  foundationregistry.Registry
	discovery discovery.Discovery
	producer  mq.Producer

	seqClientOnce sync.Once
	seqClient     sequence.Client
	seqClientErr  error
}

// NewServiceContext 创建 ServiceContext 实例
// registry: 服务注册中心，用于服务发现
// producer: MQ 生产者，用于发送消息
func NewServiceContext(registry foundationregistry.Registry, producer mq.Producer) ServiceContext {
	// 创建服务发现客户端，默认使用轮询负载均衡
	lb := loadbalancer.NewRoundRobinLoadBalancer()
	disc := discovery.NewDiscovery(registry, lb)

	return &serviceContextImpl{
		registry:  registry,
		discovery: disc,
		producer:  producer,
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

// GetSequenceServiceClient 懒加载创建 SequenceService 客户端
func (s *serviceContextImpl) GetSequenceServiceClient(ctx context.Context) (sequence.Client, error) {
	s.seqClientOnce.Do(func() {
		// 通过服务发现获取 sequence-service 的实例
		// 这里使用 3 秒超时，避免长时间阻塞
		discoveryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		instance, err := s.discovery.GetInstance(discoveryCtx, "sequence-service")
		if err != nil {
			s.seqClientErr = fmt.Errorf("failed to discover sequence-service: %w", err)
			return
		}

		target := fmt.Sprintf("%s:%d", instance.Host, instance.Port)
		clientImpl, err := sequence.NewClient(
			"sequence-service",
			client.WithHostPorts(target),
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
