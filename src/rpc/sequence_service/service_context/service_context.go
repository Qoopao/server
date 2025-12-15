package servicecontext

import (
	foundationregistry "github.com/rhp-QE/roc-foundation-util-go/service_registry/registry"
)

// ServiceContext 管理服务的全局共享资源
// 注意：sequence服务只依赖Redis，不需要获取其他服务客户端
type ServiceContext interface {
	// 获取服务注册中心
	GetRegistry() foundationregistry.Registry

	// 关闭资源
	Close() error
}

// serviceContextImpl ServiceContext的实现
type serviceContextImpl struct {
	registry foundationregistry.Registry
}

// NewServiceContext 创建ServiceContext实例
func NewServiceContext(registry foundationregistry.Registry) ServiceContext {
	return &serviceContextImpl{
		registry: registry,
	}
}

// GetRegistry 获取服务注册中心
func (s *serviceContextImpl) GetRegistry() foundationregistry.Registry {
	return s.registry
}

// Close 清理资源
func (s *serviceContextImpl) Close() error {
	if s.registry != nil {
		return s.registry.Close()
	}
	return nil
}

