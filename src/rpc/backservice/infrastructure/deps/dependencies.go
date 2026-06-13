// Package deps 管理 backservice 的运行期依赖。
package deps

import (
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/kitex/client"
	kitexdiscovery "github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	conversationservice "github.com/rhp-QE/roc-im-server/kitex_gen/conversation/conversationservice"
	messageservice "github.com/rhp-QE/roc-im-server/kitex_gen/message/messageservice"
	consts "github.com/rhp-QE/roc-im-server/src/const"
)

// ServiceContext 管理 backservice 的下游 Kitex 客户端。
// 服务发现和实例选择交给 Kitex resolver/client。
type ServiceContext interface {
	// GetMessageServiceClient 获取消息服务客户端（懒加载）
	GetMessageServiceClient() (messageservice.Client, error)

	// GetConversationServiceClient 获取会话服务客户端（懒加载）
	GetConversationServiceClient() (conversationservice.Client, error)

	// Close 关闭所有资源
	Close() error
}

type clientEntry struct {
	client interface{}
	once   sync.Once
}

type impl struct {
	resolver       kitexdiscovery.Resolver
	serviceClients sync.Map // serviceName -> *clientEntry
}

// NewServiceContext 创建服务上下文。
func NewServiceContext(resolver kitexdiscovery.Resolver) ServiceContext {
	return &impl{resolver: resolver}
}

// GetMessageServiceClient 获取消息服务客户端（懒加载）
func (i *impl) GetMessageServiceClient() (messageservice.Client, error) {
	clientIface, err := i.getServiceClient(consts.MessageServiceName, func(opts []client.Option) (interface{}, error) {
		return messageservice.NewClient(consts.MessageServiceName, opts...)
	})
	if err != nil {
		return nil, err
	}
	return clientIface.(messageservice.Client), nil
}

// GetConversationServiceClient 获取会话服务客户端（懒加载）
func (i *impl) GetConversationServiceClient() (conversationservice.Client, error) {
	clientIface, err := i.getServiceClient(consts.ConversationServiceName, func(opts []client.Option) (interface{}, error) {
		return conversationservice.NewClient(consts.ConversationServiceName, opts...)
	})
	if err != nil {
		return nil, err
	}
	return clientIface.(conversationservice.Client), nil
}

func (i *impl) getServiceClient(
	serviceName string,
	clientFactory func([]client.Option) (interface{}, error),
) (interface{}, error) {
	entryInterface, _ := i.serviceClients.LoadOrStore(serviceName, &clientEntry{})
	entry := entryInterface.(*clientEntry)

	var err error
	entry.once.Do(func() {
		opts := i.clientOptions()
		newClient, createErr := clientFactory(opts)
		if createErr != nil {
			err = fmt.Errorf("failed to create kitex client for %s: %w", serviceName, createErr)
			return
		}

		entry.client = newClient
		klog.Infof("backservice: cached kitex client, service=%s", serviceName)
	})

	if err != nil {
		i.serviceClients.Delete(serviceName)
		return nil, err
	}

	return entry.client, nil
}

func (i *impl) clientOptions() []client.Option {
	opts := []client.Option{
		client.WithSuite(tracing.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: consts.BackserviceIMName}),
		client.WithRPCTimeout(10 * time.Second),
	}
	if i.resolver != nil {
		opts = append(opts, client.WithResolver(i.resolver))
	}
	return opts
}

// Close 关闭所有资源
func (i *impl) Close() error {
	var errs []error

	i.serviceClients.Range(func(key, value interface{}) bool {
		serviceName := key.(string)
		entry := value.(*clientEntry)

		if entry.client != nil {
			if closer, ok := entry.client.(interface{ Close() error }); ok {
				if err := closer.Close(); err != nil {
					errs = append(errs, fmt.Errorf("failed to close client for service %s: %w", serviceName, err))
				} else {
					klog.Infof("backservice: closed kitex client, service=%s", serviceName)
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

var _ ServiceContext = (*impl)(nil)
