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
	backbonservice "github.com/rhp-QE/roc-foundation-service/long_connection_service/kitex_gen/backbon/backbonservice"
	foundationmq "github.com/rhp-QE/roc-foundation-util-go/mq"
	"github.com/rhp-QE/roc-foundation-util-go/service_registry/discovery"
	"github.com/rhp-QE/roc-foundation-util-go/service_registry/loadbalancer"
	foundationregistry "github.com/rhp-QE/roc-foundation-util-go/service_registry/registry"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
)

// ServiceContext 管理 conv_msg_consumer 的全局共享资源
type ServiceContext interface {
	// GetStorage 获取底层存储（MongoDB）
	GetStorage() foundationstorage.Storage

	// GetConsumer 获取 MQ Consumer
	GetConsumer() foundationmq.Consumer

	// GetBackbonServiceClient 获取 backbon 服务客户端（懒加载）
	GetBackbonServiceClient(ctx context.Context) (backbonservice.Client, error)

	// Close 关闭所有资源
	Close() error
}

type serviceContextImpl struct {
	storage       foundationstorage.Storage
	consumer      foundationmq.Consumer
	registry      foundationregistry.Registry
	discovery     discovery.Discovery
	backbonClient sync.Map // 缓存 backbon 服务客户端
}

// NewServiceContext 创建 ServiceContext 实例
func NewServiceContext(storage foundationstorage.Storage, consumer foundationmq.Consumer, registry foundationregistry.Registry) ServiceContext {
	lb := loadbalancer.NewRoundRobinLoadBalancer()
	discovery := discovery.NewDiscovery(registry, lb)
	return &serviceContextImpl{
		storage:   storage,
		consumer:  consumer,
		registry:  registry,
		discovery: discovery,
	}
}

func (s *serviceContextImpl) GetStorage() foundationstorage.Storage {
	return s.storage
}

func (s *serviceContextImpl) GetConsumer() foundationmq.Consumer {
	return s.consumer
}

// GetBackbonServiceClient 获取 backbon 服务客户端（懒加载）
func (s *serviceContextImpl) GetBackbonServiceClient(ctx context.Context) (backbonservice.Client, error) {
	serviceName := "backbon-service"

	// 获取或创建客户端
	entryInterface, _ := s.backbonClient.LoadOrStore(serviceName, &clientEntry{})
	entry := entryInterface.(*clientEntry)

	var err error
	entry.once.Do(func() {
		// 从注册中心获取服务实例
		discoverCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		instance, discoverErr := s.discovery.GetInstance(discoverCtx, serviceName)
		if discoverErr != nil {
			err = fmt.Errorf("failed to discover service instance %s: %w", serviceName, discoverErr)
			return
		}

		// 构建 hostPort
		hostPort := fmt.Sprintf("%s:%d", instance.Host, instance.Port)

		// 创建客户端
		client, createErr := backbonservice.NewClient(
			serviceName,
			client.WithHostPorts(hostPort),
			client.WithSuite(tracing.NewClientSuite()),
			client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: serviceName}),
			client.WithRPCTimeout(10*time.Second), // 设置 RPC 超时时间为 10 秒
		)
		if createErr != nil {
			err = fmt.Errorf("failed to create backbon service client at %s: %w", hostPort, createErr)
			return
		}

		entry.client = client
		klog.Infof("Successfully created and cached backbon service client at %s", hostPort)
	})

	if err != nil {
		s.backbonClient.Delete(serviceName)
		return nil, err
	}

	return entry.client.(backbonservice.Client), nil
}

// clientEntry 客户端条目，包含客户端和创建同步原语
type clientEntry struct {
	client interface{}
	once   sync.Once
}

// Close 按顺序关闭所有资源
func (s *serviceContextImpl) Close() error {
	var errs []error

	// 关闭 backbon 服务客户端
	s.backbonClient.Range(func(key, value interface{}) bool {
		entry := value.(*clientEntry)
		if entry.client != nil {
			if closer, ok := entry.client.(interface{ Close() error }); ok {
				if err := closer.Close(); err != nil {
					errs = append(errs, fmt.Errorf("failed to close backbon client: %w", err))
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
