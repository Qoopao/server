package msggateway

import (
	"os"
	"sync"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/transport"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	"github.com/roc/roc-im-server/internal/kitex_gen/conversation/conversationservice"
	"github.com/roc/roc-im-server/internal/kitex_gen/msg/messageservice"
	"github.com/roc/roc-im-server/tools/loadbalancer"
	servicediscovery "github.com/roc/roc-im-server/tools/serviceDiscovery"
	serviceregistry "github.com/roc/roc-im-server/tools/serviceRegistry"
	"go.uber.org/zap"
)

var (
	discoveryClient *servicediscovery.ServiceDiscoveryClient
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

		// 创建 Etcd 注册中心
		etcdEndpoints := []string{"localhost:2379"}
		if endpoints := os.Getenv("ETCD_ENDPOINTS"); endpoints != "" {
			etcdEndpoints = []string{endpoints}
		}

		config := serviceregistry.EtcdConfig{
			Endpoints:   etcdEndpoints,
			RootPath:    "/roc-im-server/services",
			DialTimeout: 5 * time.Second,
		}

		registry, err := serviceregistry.NewEtcdRegistry(config, logger)
		if err != nil {
			initErr = err
			return
		}

		// 创建负载均衡器工厂
		lbFactory := &loadbalancer.DefaultLoadBalancerFactory{}
		loadBalancer := lbFactory.CreateLoadBalancer(loadbalancer.RoundRobin)

		// 创建服务发现客户端
		discoveryClient = servicediscovery.NewServiceDiscoveryClient(servicediscovery.Config{
			Registry:     registry,
			LoadBalancer: loadBalancer,
			Logger:       logger,
		})
	})

	return initErr
}

// GetMsgServiceClient 获取消息服务客户端
func GetMsgServiceClient() (messageservice.Client, error) {
	// 初始化服务发现
	if err := initServiceDiscovery(); err != nil {
		return nil, err
	}

	// 使用服务发现获取客户端
	clientInterface, err := discoveryClient.GetKitexClient("msg-service", func(hostPort string) (interface{}, error) {
		return messageservice.NewClient("msg-service",
			client.WithHostPorts(hostPort),
			client.WithTransportProtocol(transport.GRPC),
			client.WithSuite(tracing.NewClientSuite()),
			client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "msg-service"}),
		)
	})

	if err != nil {
		return nil, err
	}

	// 类型断言
	msgClient, ok := clientInterface.(messageservice.Client)
	if !ok {
		logger.Error("Failed to cast client to messageservice.Client")
		return nil, err
	}

	return msgClient, nil
}

// GetMsgServiceClientWithKey 使用指定key获取消息服务客户端（用于一致性哈希）
func GetMsgServiceClientWithKey(key string) (messageservice.Client, error) {
	// 初始化服务发现
	if err := initServiceDiscovery(); err != nil {
		return nil, err
	}

	// 使用服务发现获取客户端
	clientInterface, err := discoveryClient.GetKitexClient("msg-service", func(hostPort string) (interface{}, error) {
		return messageservice.NewClient("msg-service",
			client.WithHostPorts(hostPort),
			client.WithTransportProtocol(transport.GRPC))
	}, key)

	if err != nil {
		return nil, err
	}

	// 类型断言
	msgClient, ok := clientInterface.(messageservice.Client)
	if !ok {
		logger.Error("Failed to cast client to messageservice.Client")
		return nil, err
	}

	return msgClient, nil
}

// GetConversationServiceClient 获取会话服务客户端
func GetConversationServiceClient() (conversationservice.Client, error) {
	// 初始化服务发现
	if err := initServiceDiscovery(); err != nil {
		return nil, err
	}

	// 使用服务发现获取客户端
	clientInterface, err := discoveryClient.GetKitexClient("conversation-service", func(hostPort string) (interface{}, error) {
		return conversationservice.NewClient("conversation-service",
			client.WithHostPorts(hostPort),
			client.WithTransportProtocol(transport.GRPC))
	})

	if err != nil {
		return nil, err
	}

	// 类型断言
	convClient, ok := clientInterface.(conversationservice.Client)
	if !ok {
		logger.Error("Failed to cast client to conversationservice.Client")
		return nil, err
	}

	return convClient, nil
}

// GetConversationServiceClientWithKey 使用指定key获取会话服务客户端（用于一致性哈希）
func GetConversationServiceClientWithKey(key string) (conversationservice.Client, error) {
	// 初始化服务发现
	if err := initServiceDiscovery(); err != nil {
		return nil, err
	}

	// 使用服务发现获取客户端
	clientInterface, err := discoveryClient.GetKitexClient("conversation-service", func(hostPort string) (interface{}, error) {
		return conversationservice.NewClient("conversation-service",
			client.WithHostPorts(hostPort),
			client.WithTransportProtocol(transport.GRPC))
	}, key)

	if err != nil {
		return nil, err
	}

	// 类型断言
	convClient, ok := clientInterface.(conversationservice.Client)
	if !ok {
		logger.Error("Failed to cast client to conversationservice.Client")
		return nil, err
	}

	return convClient, nil
}

// RefreshMsgService 刷新消息服务实例缓存
func RefreshMsgService() error {
	if discoveryClient == nil {
		return initServiceDiscovery()
	}
	return discoveryClient.RefreshServiceInstances("msg-service")
}

// RefreshConversationService 刷新会话服务实例缓存
func RefreshConversationService() error {
	if discoveryClient == nil {
		return initServiceDiscovery()
	}
	return discoveryClient.RefreshServiceInstances("conversation-service")
}

// GetServiceDiscoveryClient 获取服务发现客户端（用于其他服务）
func GetServiceDiscoveryClient() *servicediscovery.ServiceDiscoveryClient {
	initServiceDiscovery()
	return discoveryClient
}
