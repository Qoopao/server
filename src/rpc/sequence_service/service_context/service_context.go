package servicecontext

import (
	"github.com/redis/go-redis/v9"
	foundationregistry "github.com/rhp-QE/roc-foundation-util-go/service_registry/registry"
)

// ServiceContext 管理服务的全局共享资源
// 注意：sequence 服务依赖 Redis 和注册中心
type ServiceContext interface {
	// 获取服务注册中心
	GetRegistry() foundationregistry.Registry
	// 获取 Redis 客户端
	GetRedisClient() *redis.Client
	// 关闭资源
	Close() error
}

// serviceContextImpl ServiceContext 的实现
type serviceContextImpl struct {
	registry    foundationregistry.Registry
	redisClient *redis.Client
}

// NewServiceContext 创建 ServiceContext 实例
func NewServiceContext(registry foundationregistry.Registry, redisClient *redis.Client) ServiceContext {
	return &serviceContextImpl{
		registry:    registry,
		redisClient: redisClient,
	}
}

// GetRegistry 获取服务注册中心
func (s *serviceContextImpl) GetRegistry() foundationregistry.Registry {
	return s.registry
}

// GetRedisClient 获取 Redis 客户端
func (s *serviceContextImpl) GetRedisClient() *redis.Client {
	return s.redisClient
}

// Close 清理资源
func (s *serviceContextImpl) Close() error {
	if s.redisClient != nil {
		_ = s.redisClient.Close()
	}
	if s.registry != nil {
		return s.registry.Close()
	}
	return nil
}
