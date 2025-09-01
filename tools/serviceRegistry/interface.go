package serviceregistry

import "context"

// 服务实例信息
type ServiceInstance struct {
	InstanceID  string            // 实例唯一标识
	ServiceName string            // 所属服务名称
	Host        string            // 主机地址
	Port        int               // 服务端口
	Metadata    map[string]string // 元数据
	Status      InstanceStatus    // 实例状态
	Weight      int               // 负载权重
}

// InstanceStatus 实例状态
type InstanceStatus int

const (
	StatusUnknown InstanceStatus = iota
	StatusHealthy
	StatusUnhealthy
	StatusDraining
)

// String 返回状态字符串表示
func (s InstanceStatus) String() string {
	switch s {
	case StatusHealthy:
		return "healthy"
	case StatusUnhealthy:
		return "unhealthy"
	case StatusDraining:
		return "draining"
	default:
		return "unknown"
	}
}

// 服务实例变更监听器
type InstanceChangeListener func(serviceName string, instances []ServiceInstance)

// 服务注册中心接口
type ServiceRegistry interface {
	// 实例生命周期管理
	RegisterInstance(ctx context.Context, instance *ServiceInstance) error
	DeregisterInstance(ctx context.Context, instanceID string) error

	// 服务发现
	DiscoverInstances(ctx context.Context, serviceName string) ([]ServiceInstance, error)
	GetInstance(ctx context.Context, instanceID string) (*ServiceInstance, error)
	GetAllServices(ctx context.Context) (map[string][]ServiceInstance, error)

	// 服务变更监听
	SubscribeInstanceChanges(ctx context.Context, serviceName string, listener InstanceChangeListener) error
	UnsubscribeInstanceChanges(ctx context.Context, serviceName string) error
	GetSubscribedServices(ctx context.Context) ([]string, error)

	// 系统健康检查
	CheckHealth(ctx context.Context) error

	// 关闭资源
	Close() error
}

// 批量操作接口 (可选扩展)
type BatchServiceRegistry interface {
	ServiceRegistry
	BatchRegisterInstances(ctx context.Context, instances []*ServiceInstance) error
	BatchDeregisterInstances(ctx context.Context, instanceIDs []string) error
}

// 服务注册中心配置接口
type RegistryConfig interface {
	GetRegistryType() string
	GetEndpoints() []string
	GetRootPath() string
	GetTimeout() int
}

// 服务注册中心工厂接口
type RegistryFactory interface {
	CreateRegistry(config RegistryConfig) (ServiceRegistry, error)
}
