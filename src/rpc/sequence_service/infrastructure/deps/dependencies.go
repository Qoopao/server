package deps

import (
	"github.com/redis/go-redis/v9"
)

// ServiceContext 管理服务的全局共享资源
// 注意：sequence 服务只依赖 Redis；Kitex 服务注册由 server.WithRegistry 管理。
type ServiceContext interface {
	// 获取 Redis 客户端
	GetRedisClient() *redis.Client
	// 关闭资源
	Close() error
}

// serviceContextImpl ServiceContext 的实现
type serviceContextImpl struct {
	redisClient *redis.Client
}

// NewServiceContext 创建 ServiceContext 实例
func NewServiceContext(redisClient *redis.Client) ServiceContext {
	return &serviceContextImpl{
		redisClient: redisClient,
	}
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
	return nil
}
