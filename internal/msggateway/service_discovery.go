package msggateway

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/transport"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	"github.com/roc/roc-foundation-util-go/service_registry/discovery"
	"github.com/roc/roc-foundation-util-go/service_registry/loadbalancer"
	"github.com/roc/roc-foundation-util-go/service_registry/registry/etcd"
	"github.com/roc/roc-im-server/internal/kitex_gen/conversation/conversationservice"
	"github.com/roc/roc-im-server/internal/kitex_gen/msg/messageservice"
	"go.uber.org/zap"
)

var (
	discoveryClient discovery.Discovery
	discoveryOnce   sync.Once
	logger          *zap.Logger
)

// initServiceDiscovery 初始化服务发现客户端
func initServiceDiscovery() error {
	var initErr error
	discoveryOnce.Do(func() {
		// 初始化日志
		logger, _ = zap.NewProduction()
		if logger == nil {
			logger = zap.NewNop()
		}

		// 创建 Etcd Registry
		registry, err := etcd.NewEtcdRegistry(
			etcd.WithEndpoints([]string{"localhost:2379"}),
			etcd.WithDialTimeout(5*time.Second),
		)
		if err != nil {
			initErr = err
			return
		}

		// 创建服务发现客户端
		lb := loadbalancer.NewRoundRobinLoadBalancer()
		discoveryClient = discovery.NewDiscovery(registry, lb)
	})

	return initErr
}

// GetMsgServiceClient 获取消息服务客户端
func GetMsgServiceClient() (messageservice.Client, error) {
	// 初始化服务发现
	if err := initServiceDiscovery(); err != nil {
		return nil, err
	}

	// 使用服务发现获取服务地址
	instances, err := discoveryClient.GetInstances(context.Background(), "msg-service")
	if err != nil || len(instances) == 0 {
		return nil, fmt.Errorf("no available msg-service instances")
	}

	// 创建客户端
	hostPort := fmt.Sprintf("%s:%d", instances[0].Host, instances[0].Port)
	msgClient, err := messageservice.NewClient("msg-service",
		client.WithHostPorts(hostPort),
		client.WithTransportProtocol(transport.GRPC),
		client.WithSuite(tracing.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "msg-service"}),
	)

	return msgClient, err
}

// GetConversationServiceClient 获取会话服务客户端
func GetConversationServiceClient() (conversationservice.Client, error) {
	// 初始化服务发现
	if err := initServiceDiscovery(); err != nil {
		return nil, err
	}

	// 使用服务发现获取服务地址
	instances, err := discoveryClient.GetInstances(context.Background(), "conversation-service")
	if err != nil || len(instances) == 0 {
		return nil, fmt.Errorf("no available conversation-service instances")
	}

	// 创建客户端
	hostPort := fmt.Sprintf("%s:%d", instances[0].Host, instances[0].Port)
	convClient, err := conversationservice.NewClient("conversation-service",
		client.WithHostPorts(hostPort),
		client.WithTransportProtocol(transport.GRPC),
	)

	return convClient, err
}

// RefreshMsgService 刷新消息服务实例缓存
func RefreshMsgService() error {
	if discoveryClient == nil {
		return initServiceDiscovery()
	}
	return discoveryClient.Refresh(context.Background(), "msg-service")
}

// RefreshConversationService 刷新会话服务实例缓存
func RefreshConversationService() error {
	if discoveryClient == nil {
		return initServiceDiscovery()
	}
	return discoveryClient.Refresh(context.Background(), "conversation-service")
}

// GetServiceDiscoveryClient 获取服务发现客户端（用于其他服务）
