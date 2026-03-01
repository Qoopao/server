// Package servicecontext 服务上下文，管理全局共享资源
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
	"github.com/rhp-QE/roc-foundation-util-go/service_registry/discovery"
	"github.com/rhp-QE/roc-foundation-util-go/service_registry/loadbalancer"
	foundationregistry "github.com/rhp-QE/roc-foundation-util-go/service_registry/registry"
	conversationservice "github.com/rhp-QE/roc-im-server/kitex_gen/conversation/conversationservice"
	messageservice "github.com/rhp-QE/roc-im-server/kitex_gen/message/messageservice"
	consts "github.com/rhp-QE/roc-im-server/src/const"
)

// ServiceContext 服务上下文接口，定义全局共享资源的访问方法
type ServiceContext interface {
	// GetRegistry 获取注册中心实例
	GetRegistry() foundationregistry.Registry

	// GetDiscovery 获取服务发现实例
	GetDiscovery() discovery.Discovery

	// GetMessageServiceClient 获取消息服务客户端（懒加载）
	GetMessageServiceClient() (messageservice.Client, error)

	// GetConversationServiceClient 获取会话服务客户端（懒加载）
	GetConversationServiceClient() (conversationservice.Client, error)

	// Close 关闭所有资源
	Close() error
}

// clientEntry 客户端条目，包含客户端和创建同步原语
type clientEntry struct {
	client interface{}
	once   sync.Once // 确保客户端只创建一次
}

// impl ServiceContext 接口的实现
type impl struct {
	// Etcd Registry
	registry foundationregistry.Registry

	// 服务发现
	discovery discovery.Discovery

	// 服务客户端缓存：serviceName -> *clientEntry
	// 使用 sync.Map 保证并发安全
	serviceClients sync.Map // map[string]*clientEntry
}

// NewServiceContext 创建服务上下文
func NewServiceContext(registry foundationregistry.Registry) ServiceContext {
	// 创建服务发现客户端
	lb := loadbalancer.NewRoundRobinLoadBalancer()
	discovery := discovery.NewDiscovery(registry, lb)

	return &impl{
		registry:  registry,
		discovery: discovery,
	}
}

// GetRegistry 获取注册中心实例
func (i *impl) GetRegistry() foundationregistry.Registry {
	return i.registry
}

// GetDiscovery 获取服务发现实例
func (i *impl) GetDiscovery() discovery.Discovery {
	return i.discovery
}

// GetMessageServiceClient 获取消息服务客户端（懒加载）
func (i *impl) GetMessageServiceClient() (messageservice.Client, error) {
	serviceName := consts.MessageServiceName
	client, err := i.getServiceClient(serviceName, func(hostPort string) (interface{}, error) {
		client, err := messageservice.NewClient(
			serviceName,
			client.WithHostPorts(hostPort),
			client.WithSuite(tracing.NewClientSuite()),
			client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: serviceName}),
			client.WithRPCTimeout(10*time.Second), // 设置 RPC 超时时间为 10 秒
		)
		if err != nil {
			return nil, err
		}
		return client, nil
	})
	if err != nil {
		return nil, err
	}
	return client.(messageservice.Client), nil
}

// GetConversationServiceClient 获取会话服务客户端（懒加载）
func (i *impl) GetConversationServiceClient() (conversationservice.Client, error) {
	serviceName := consts.ConversationServiceName
	client, err := i.getServiceClient(serviceName, func(hostPort string) (interface{}, error) {
		client, err := conversationservice.NewClient(
			serviceName,
			client.WithHostPorts(hostPort),
			client.WithSuite(tracing.NewClientSuite()),
			client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: serviceName}),
			client.WithRPCTimeout(10*time.Second), // 设置 RPC 超时时间为 10 秒
		)
		if err != nil {
			return nil, err
		}
		return client, nil
	})
	if err != nil {
		return nil, err
	}
	return client.(conversationservice.Client), nil
}

// getServiceClient 通用的服务客户端获取方法（懒加载）
// serviceName: 服务名称
// clientFactory: 客户端工厂函数，接收 hostPort 返回客户端实例
func (i *impl) getServiceClient(
	serviceName string,
	clientFactory func(hostPort string) (interface{}, error),
) (interface{}, error) {
	// 获取或创建客户端条目
	entryInterface, _ := i.serviceClients.LoadOrStore(serviceName, &clientEntry{})
	entry := entryInterface.(*clientEntry)

	// 使用 sync.Once 确保客户端只创建一次
	var err error
	entry.once.Do(func() {
		// 从注册中心获取服务实例
		discoverCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		instance, discoverErr := i.discovery.GetInstance(discoverCtx, serviceName)
		if discoverErr != nil {
			err = fmt.Errorf("failed to discover service instance %s: %w", serviceName, discoverErr)
			return
		}

		// 构建 hostPort
		hostPort := fmt.Sprintf("%s:%d", instance.Host, instance.Port)

		// 创建客户端
		newClient, createErr := clientFactory(hostPort)
		if createErr != nil {
			err = fmt.Errorf("failed to create client for %s at %s: %w", serviceName, hostPort, createErr)
			return
		}

		// 存储到条目
		entry.client = newClient
		klog.Infof("Successfully created and cached client for service: %s at %s", serviceName, hostPort)
	})

	// 如果创建失败，返回错误
	if err != nil {
		// 创建失败时，从缓存中删除条目，允许下次重试 (下次就会创建新的 entery 新的 once)
		i.serviceClients.Delete(serviceName)
		return nil, err
	}

	// 返回客户端
	return entry.client, nil
}

// Close 关闭所有资源
func (i *impl) Close() error {
	var errs []error

	// 关闭所有服务客户端
	i.serviceClients.Range(func(key, value interface{}) bool {
		serviceName := key.(string)
		entry := value.(*clientEntry)

		client := entry.client

		if client != nil {
			if closer, ok := client.(interface{ Close() error }); ok {
				if err := closer.Close(); err != nil {
					errs = append(errs, fmt.Errorf("failed to close client for service %s: %w", serviceName, err))
				} else {
					klog.Infof("Successfully closed client for service: %s", serviceName)
				}
			}
		}
		return true
	})

	if len(errs) > 0 {
		return fmt.Errorf("errors closing resources: %v", errs)
	}

	return nil
}

// 编译时检查，确保 impl 实现了 ServiceContext 接口
var _ ServiceContext = (*impl)(nil)
