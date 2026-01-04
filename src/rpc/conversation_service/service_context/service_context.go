package servicecontext

import (
	foundationregistry "github.com/rhp-QE/roc-foundation-util-go/service_registry/registry"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
)

// ServiceContext 管理 conversation_service 的全局共享资源
//
// 职责：
// 1. 管理服务注册中心（Registry）
// 2. 管理 MongoDB 存储
type ServiceContext interface {
	// GetRegistry 获取服务注册中心
	GetRegistry() foundationregistry.Registry

	// GetStorage 获取 MongoDB 存储
	GetStorage() foundationstorage.Storage

	// Close 关闭所有全局资源
	Close() error
}

type serviceContextImpl struct {
	registry foundationregistry.Registry
	storage  foundationstorage.Storage
}

// NewServiceContext 创建 ServiceContext 实例
// registry: 服务注册中心，用于服务发现
// storage: MongoDB 存储，用于查询会话和消息
func NewServiceContext(registry foundationregistry.Registry, storage foundationstorage.Storage) ServiceContext {
	return &serviceContextImpl{
		registry: registry,
		storage:  storage,
	}
}

func (s *serviceContextImpl) GetRegistry() foundationregistry.Registry {
	return s.registry
}

func (s *serviceContextImpl) GetStorage() foundationstorage.Storage {
	return s.storage
}

// Close 关闭所有全局资源
func (s *serviceContextImpl) Close() error {
	var errs []error

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

