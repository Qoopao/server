package serviceregistry

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

type InstanceStatus int

const (
	StatusUnknown InstanceStatus = iota
	StatusHealthy
	StatusUnhealthy
	StatusDraining
)

// 服务实例变更监听器
type InstanceChangeListener func(serviceName string, instances []ServiceInstance)

// 服务注册中心接口
type ServiceRegistry interface {
	// 实例生命周期管理
	RegisterInstance(instance *ServiceInstance) error
	DeregisterInstance(instanceID string) error

	// 服务发现
	DiscoverInstances(serviceName string) ([]ServiceInstance, error)
	SubscribeInstanceChanges(serviceName string, listener InstanceChangeListener) error
	UnsubscribeInstanceChanges(serviceName string) error

	// 实例状态管理
	UpdateInstanceStatus(instanceID string, status InstanceStatus) error
	UpdateInstanceMetadata(instanceID string, metadata map[string]string) error

	// 系统健康检查
	CheckHealth() error

	// 关闭资源
	Close() error
}

// 批量操作接口 (可选扩展)
type BatchServiceRegistry interface {
	ServiceRegistry
	BatchRegisterInstances(instances []*ServiceInstance) error
	BatchDeregisterInstances(instanceIDs []string) error
}
