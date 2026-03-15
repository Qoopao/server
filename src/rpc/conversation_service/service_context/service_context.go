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
	"github.com/rhp-QE/roc-foundation-util-go/service_registry/discovery"
	"github.com/rhp-QE/roc-foundation-util-go/service_registry/loadbalancer"
	foundationregistry "github.com/rhp-QE/roc-foundation-util-go/service_registry/registry"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	messageservice "github.com/rhp-QE/roc-im-server/kitex_gen/message/messageservice"
	consts "github.com/rhp-QE/roc-im-server/src/const"
)

// ServiceContext 管理 conversation_service 的全局共享资源
//
// 职责：
// 1. 管理服务注册中心（Registry）
// 2. 管理 MongoDB 存储
// 3. 懒加载 message_service 客户端
type ServiceContext interface {
	// GetRegistry 获取服务注册中心
	GetRegistry() foundationregistry.Registry

	// GetStorage 获取 MongoDB 存储
	GetStorage() foundationstorage.Storage

	// GetMessageServiceClient 获取 message_service 客户端（懒加载）
	GetMessageServiceClient() (messageservice.Client, error)

	// GetBackbonServiceClient 获取 backbon 服务客户端（懒加载，用于下推 cmd 等）
	GetBackbonServiceClient(ctx context.Context) (backbonservice.Client, error)

	// Close 关闭所有全局资源
	Close() error
}

type clientEntry struct {
	client interface{}
	once   sync.Once
}

type serviceContextImpl struct {
	registry foundationregistry.Registry
	storage  foundationstorage.Storage

	discovery      discovery.Discovery
	serviceClients sync.Map // serviceName -> *clientEntry
}

// NewServiceContext 创建 ServiceContext 实例
// registry: 服务注册中心，用于服务发现
// storage: MongoDB 存储，用于查询会话和消息
func NewServiceContext(registry foundationregistry.Registry, storage foundationstorage.Storage) ServiceContext {
	lb := loadbalancer.NewRoundRobinLoadBalancer()
	dis := discovery.NewDiscovery(registry, lb)

	return &serviceContextImpl{
		registry:  registry,
		storage:   storage,
		discovery: dis,
	}
}

func (s *serviceContextImpl) GetRegistry() foundationregistry.Registry {
	return s.registry
}

func (s *serviceContextImpl) GetStorage() foundationstorage.Storage {
	return s.storage
}

// GetMessageServiceClient 获取 message_service 客户端（懒加载）
func (s *serviceContextImpl) GetMessageServiceClient() (messageservice.Client, error) {
	const serviceName = consts.MessageServiceName

	clientIface, err := s.getServiceClient(serviceName, func(hostPort string) (interface{}, error) {
		c, err := messageservice.NewClient(
			serviceName,
			client.WithHostPorts(hostPort),
			client.WithSuite(tracing.NewClientSuite()),
			client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: serviceName}),
			client.WithRPCTimeout(10*time.Second),
		)
		if err != nil {
			return nil, err
		}
		return c, nil
	})
	if err != nil {
		return nil, err
	}

	return clientIface.(messageservice.Client), nil
}

// GetBackbonServiceClient 获取 backbon 服务客户端（懒加载）
func (s *serviceContextImpl) GetBackbonServiceClient(ctx context.Context) (backbonservice.Client, error) {
	const serviceName = "backbon-service"
	clientIface, err := s.getServiceClient(serviceName, func(hostPort string) (interface{}, error) {
		c, err := backbonservice.NewClient(
			serviceName,
			client.WithHostPorts(hostPort),
			client.WithSuite(tracing.NewClientSuite()),
			client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: serviceName}),
			client.WithRPCTimeout(10*time.Second),
		)
		if err != nil {
			return nil, err
		}
		return c, nil
	})
	if err != nil {
		return nil, err
	}
	return clientIface.(backbonservice.Client), nil
}

func (s *serviceContextImpl) getServiceClient(
	serviceName string,
	clientFactory func(hostPort string) (interface{}, error),
) (interface{}, error) {
	entryIface, _ := s.serviceClients.LoadOrStore(serviceName, &clientEntry{})
	entry := entryIface.(*clientEntry)

	var err error
	entry.once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		instance, discoverErr := s.discovery.GetInstance(ctx, serviceName)
		if discoverErr != nil {
			err = fmt.Errorf("failed to discover service instance %s: %w", serviceName, discoverErr)
			return
		}

		hostPort := fmt.Sprintf("%s:%d", instance.Host, instance.Port)
		newClient, createErr := clientFactory(hostPort)
		if createErr != nil {
			err = fmt.Errorf("failed to create client for %s at %s: %w", serviceName, hostPort, createErr)
			return
		}

		entry.client = newClient
		klog.Infof("Successfully created and cached client for service: %s at %s", serviceName, hostPort)
	})

	if err != nil {
		s.serviceClients.Delete(serviceName)
		return nil, err
	}

	return entry.client, nil
}

// Close 关闭所有全局资源
func (s *serviceContextImpl) Close() error {
	var errs []error

	// 关闭下游服务客户端
	s.serviceClients.Range(func(key, value interface{}) bool {
		serviceName := key.(string)
		entry := value.(*clientEntry)

		if entry.client != nil {
			if closer, ok := entry.client.(interface{ Close() error }); ok {
				if err := closer.Close(); err != nil {
					errs = append(errs, fmt.Errorf("failed to close client for service %s: %w", serviceName, err))
				} else {
					klog.Infof("Successfully closed client for service: %s", serviceName)
				}
			}
		}
		return true
	})

	if s.storage != nil {
		if err := s.storage.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if s.registry != nil {
		if err := s.registry.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

// 编译期检查，确保实现了 ServiceContext 接口
var _ ServiceContext = (*serviceContextImpl)(nil)

