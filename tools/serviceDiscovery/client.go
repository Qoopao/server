package servicediscovery

import (
	"context"
	"fmt"
	"sync"

	"github.com/roc/roc-im-server/tools/loadbalancer"
	serviceregistry "github.com/roc/roc-im-server/tools/serviceRegistry"
	"go.uber.org/zap"
)

// ServiceDiscoveryClient 服务发现客户端
type ServiceDiscoveryClient struct {
	registry     serviceregistry.ServiceRegistry
	loadBalancer loadbalancer.LoadBalancer
	logger       *zap.Logger

	// 服务实例缓存
	instanceCache map[string][]serviceregistry.ServiceInstance
	cacheMutex    sync.RWMutex

	// 服务连接缓存
	connCache map[string]interface{} // 存储各种类型的客户端连接
	connMutex sync.RWMutex
}

// Config 服务发现客户端配置
type Config struct {
	Registry     serviceregistry.ServiceRegistry
	LoadBalancer loadbalancer.LoadBalancer
	Logger       *zap.Logger
}

// NewServiceDiscoveryClient 创建服务发现客户端
func NewServiceDiscoveryClient(config Config) *ServiceDiscoveryClient {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}

	if config.LoadBalancer == nil {
		factory := &loadbalancer.DefaultLoadBalancerFactory{}
		config.LoadBalancer = factory.CreateLoadBalancer(loadbalancer.RoundRobin)
	}

	client := &ServiceDiscoveryClient{
		registry:      config.Registry,
		loadBalancer:  config.LoadBalancer,
		logger:        config.Logger,
		instanceCache: make(map[string][]serviceregistry.ServiceInstance),
		connCache:     make(map[string]interface{}),
	}

	return client
}

// GetServiceInstance 获取服务实例
func (c *ServiceDiscoveryClient) GetServiceInstance(serviceName string, key ...string) (*serviceregistry.ServiceInstance, error) {
	// 先从缓存获取
	c.cacheMutex.RLock()
	instances, exists := c.instanceCache[serviceName]
	c.cacheMutex.RUnlock()

	if !exists {
		// 缓存中没有，从注册中心获取
		var err error
		instances, err = c.registry.DiscoverInstances(context.Background(), serviceName)
		if err != nil {
			return nil, fmt.Errorf("failed to discover instances for service %s: %w", serviceName, err)
		}

		// 更新缓存
		c.cacheMutex.Lock()
		c.instanceCache[serviceName] = instances
		c.cacheMutex.Unlock()

		// 订阅变更
		c.registry.SubscribeInstanceChanges(context.Background(), serviceName, func(svcName string, newInstances []serviceregistry.ServiceInstance) {
			c.cacheMutex.Lock()
			c.instanceCache[svcName] = newInstances
			c.cacheMutex.Unlock()

			// 清除连接缓存，强制重新创建连接
			c.connMutex.Lock()
			delete(c.connCache, svcName)
			c.connMutex.Unlock()

			c.logger.Info("Service instances updated",
				zap.String("service", svcName),
				zap.Int("count", len(newInstances)))
		})
	}

	// 使用负载均衡器选择实例
	return c.loadBalancer.Select(instances, key...)
}

// GetKitexClient 获取 Kitex 客户端连接
func (c *ServiceDiscoveryClient) GetKitexClient(serviceName string, clientFactory func(hostPort string) (interface{}, error), key ...string) (interface{}, error) {
	// 检查连接缓存
	c.connMutex.RLock()
	if cachedClient, exists := c.connCache[serviceName]; exists {
		c.connMutex.RUnlock()
		return cachedClient, nil
	}
	c.connMutex.RUnlock()

	// 获取服务实例
	instance, err := c.GetServiceInstance(serviceName, key...)
	if err != nil {
		return nil, err
	}

	// 创建客户端连接
	hostPort := fmt.Sprintf("%s:%d", instance.Host, instance.Port)
	kitexClient, err := clientFactory(hostPort)
	if err != nil {
		return nil, fmt.Errorf("failed to create kitex client for %s: %w", hostPort, err)
	}

	// 缓存连接
	c.connMutex.Lock()
	c.connCache[serviceName] = kitexClient
	c.connMutex.Unlock()

	c.logger.Info("Created new kitex client",
		zap.String("service", serviceName),
		zap.String("instance", hostPort))

	return kitexClient, nil
}

// RefreshServiceInstances 刷新服务实例缓存
func (c *ServiceDiscoveryClient) RefreshServiceInstances(serviceName string) error {
	instances, err := c.registry.DiscoverInstances(context.Background(), serviceName)
	if err != nil {
		return err
	}

	c.cacheMutex.Lock()
	c.instanceCache[serviceName] = instances
	c.cacheMutex.Unlock()

	// 清除连接缓存
	c.connMutex.Lock()
	delete(c.connCache, serviceName)
	c.connMutex.Unlock()

	return nil
}

// Close 关闭客户端
func (c *ServiceDiscoveryClient) Close() error {
	// 清理缓存
	c.cacheMutex.Lock()
	c.instanceCache = make(map[string][]serviceregistry.ServiceInstance)
	c.cacheMutex.Unlock()

	c.connMutex.Lock()
	c.connCache = make(map[string]interface{})
	c.connMutex.Unlock()

	return nil
}

// HealthCheck 健康检查
func (c *ServiceDiscoveryClient) HealthCheck() error {
	return c.registry.CheckHealth(context.Background())
}

// GetAllServices 获取所有服务的实例信息
func (c *ServiceDiscoveryClient) GetAllServices() map[string][]serviceregistry.ServiceInstance {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	result := make(map[string][]serviceregistry.ServiceInstance)
	for service, instances := range c.instanceCache {
		result[service] = make([]serviceregistry.ServiceInstance, len(instances))
		copy(result[service], instances)
	}
	return result
}
