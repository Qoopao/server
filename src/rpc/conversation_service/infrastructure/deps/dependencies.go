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
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	messageservice "github.com/rhp-QE/roc-im-server/kitex_gen/message/messageservice"
	consts "github.com/rhp-QE/roc-im-server/src/const"
)

// ServiceContext 管理 conversation_service 的全局共享资源。
// 服务发现、负载均衡和实例缓存使用 Kitex resolver/client，业务代码不再持有自研 registry/discovery。
type ServiceContext interface {
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
	storage  foundationstorage.Storage
	resolver kitexdiscovery.Resolver

	serviceClients sync.Map // serviceName -> *clientEntry
}

// NewServiceContext 创建 ServiceContext 实例。
func NewServiceContext(storage foundationstorage.Storage, resolver kitexdiscovery.Resolver) ServiceContext {
	return &serviceContextImpl{
		storage:  storage,
		resolver: resolver,
	}
}

func (s *serviceContextImpl) GetStorage() foundationstorage.Storage {
	return s.storage
}

// GetMessageServiceClient 获取 message_service 客户端（懒加载）
func (s *serviceContextImpl) GetMessageServiceClient() (messageservice.Client, error) {
	clientIface, err := s.getServiceClient(consts.MessageServiceName, func(opts []client.Option) (interface{}, error) {
		return messageservice.NewClient(consts.MessageServiceName, opts...)
	})
	if err != nil {
		return nil, err
	}
	return clientIface.(messageservice.Client), nil
}

// GetBackbonServiceClient 获取 backbon 服务客户端（懒加载）
func (s *serviceContextImpl) GetBackbonServiceClient(_ context.Context) (backbonservice.Client, error) {
	clientIface, err := s.getServiceClient(consts.BackbonServiceName, func(opts []client.Option) (interface{}, error) {
		return backbonservice.NewClient(consts.BackbonServiceName, opts...)
	})
	if err != nil {
		return nil, err
	}
	return clientIface.(backbonservice.Client), nil
}

func (s *serviceContextImpl) getServiceClient(
	serviceName string,
	clientFactory func([]client.Option) (interface{}, error),
) (interface{}, error) {
	entryIface, _ := s.serviceClients.LoadOrStore(serviceName, &clientEntry{})
	entry := entryIface.(*clientEntry)

	var err error
	entry.once.Do(func() {
		opts := s.clientOptions()
		newClient, createErr := clientFactory(opts)
		if createErr != nil {
			err = fmt.Errorf("failed to create kitex client for %s: %w", serviceName, createErr)
			return
		}

		entry.client = newClient
		klog.Infof("conversation_service: cached kitex client, service=%s", serviceName)
	})

	if err != nil {
		s.serviceClients.Delete(serviceName)
		return nil, err
	}

	return entry.client, nil
}

func (s *serviceContextImpl) clientOptions() []client.Option {
	opts := []client.Option{
		client.WithSuite(tracing.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: consts.ConversationServiceName}),
		client.WithRPCTimeout(10 * time.Second),
	}
	if s.resolver != nil {
		opts = append(opts, client.WithResolver(s.resolver))
	}
	return opts
}

// Close 关闭所有全局资源
func (s *serviceContextImpl) Close() error {
	var errs []error

	s.serviceClients.Range(func(key, value interface{}) bool {
		serviceName := key.(string)
		entry := value.(*clientEntry)

		if entry.client != nil {
			if closer, ok := entry.client.(interface{ Close() error }); ok {
				if err := closer.Close(); err != nil {
					errs = append(errs, fmt.Errorf("failed to close client for service %s: %w", serviceName, err))
				} else {
					klog.Infof("conversation_service: closed kitex client, service=%s", serviceName)
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

	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

var _ ServiceContext = (*serviceContextImpl)(nil)
