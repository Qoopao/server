package serviceregistry

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// EtcdRegistry 基于 Etcd 的服务注册中心实现
type EtcdRegistry struct {
	client         *clientv3.Client
	rootPath       string
	leaseID        clientv3.LeaseID
	keepAliveResp  <-chan *clientv3.LeaseKeepAliveResponse
	logger         *zap.Logger
	listeners      map[string][]InstanceChangeListener
	listenersMutex sync.RWMutex
	watchCancel    map[string]context.CancelFunc
	watchMutex     sync.RWMutex
}

// EtcdConfig Etcd配置
type EtcdConfig struct {
	Endpoints   []string      `yaml:"endpoints"` // 支持多个etcd端点
	Username    string        `yaml:"username"`
	Password    string        `yaml:"password"`
	RootPath    string        `yaml:"rootPath"`
	DialTimeout time.Duration `yaml:"dialTimeout"`
	// 新增集群相关配置
	ClusterMode bool        `yaml:"clusterMode"` // 是否为集群模式
	RetryPolicy RetryPolicy `yaml:"retryPolicy"` // 重试策略
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxRetries    int           `yaml:"maxRetries"`    // 最大重试次数
	RetryInterval time.Duration `yaml:"retryInterval"` // 重试间隔
	BackoffFactor float64       `yaml:"backoffFactor"` // 退避因子
}

// NewEtcdRegistry 创建 Etcd 服务注册中心
func NewEtcdRegistry(config EtcdConfig, logger *zap.Logger) (*EtcdRegistry, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   config.Endpoints,
		Username:    config.Username,
		Password:    config.Password,
		DialTimeout: config.DialTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	registry := &EtcdRegistry{
		client:      client,
		rootPath:    config.RootPath,
		logger:      logger,
		listeners:   make(map[string][]InstanceChangeListener),
		watchCancel: make(map[string]context.CancelFunc),
	}

	return registry, nil
}

// RegisterInstance 注册服务实例
func (r *EtcdRegistry) RegisterInstance(ctx context.Context, instance *ServiceInstance) error {
	// 创建租约
	lease, err := r.client.Grant(ctx, 30) // 30秒租约
	if err != nil {
		return fmt.Errorf("failed to create lease: %w", err)
	}

	r.leaseID = lease.ID

	// 序列化实例信息
	instanceData, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("failed to marshal instance: %w", err)
	}

	// 构建 key
	key := fmt.Sprintf("%s/%s/%s", r.rootPath, instance.ServiceName, instance.InstanceID)

	// 注册实例
	_, err = r.client.Put(ctx, key, string(instanceData), clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("failed to register instance: %w", err)
	}

	// 开始续约
	r.keepAliveResp, err = r.client.KeepAlive(ctx, lease.ID)
	if err != nil {
		return fmt.Errorf("failed to keep alive: %w", err)
	}

	// 启动续约协程
	go r.processKeepAlive()

	r.logger.Info("Service instance registered",
		zap.String("service", instance.ServiceName),
		zap.String("instance", instance.InstanceID),
		zap.String("address", fmt.Sprintf("%s:%d", instance.Host, instance.Port)))

	return nil
}

// DeregisterInstance 注销服务实例
func (r *EtcdRegistry) DeregisterInstance(ctx context.Context, instanceID string) error {
	// 撤销租约，这会自动删除相关的 key
	if r.leaseID != 0 {
		_, err := r.client.Revoke(ctx, r.leaseID)
		if err != nil {
			r.logger.Error("Failed to revoke lease", zap.Error(err))
		}
	}

	r.logger.Info("Service instance deregistered", zap.String("instance", instanceID))
	return nil
}

// DiscoverInstances 发现服务实例
func (r *EtcdRegistry) DiscoverInstances(ctx context.Context, serviceName string) ([]ServiceInstance, error) {
	prefix := fmt.Sprintf("%s/%s/", r.rootPath, serviceName)

	resp, err := r.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get instances: %w", err)
	}

	var instances []ServiceInstance
	for _, kv := range resp.Kvs {
		var instance ServiceInstance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			r.logger.Warn("Failed to unmarshal instance", zap.Error(err))
			continue
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

// GetInstance 获取单个服务实例
func (r *EtcdRegistry) GetInstance(ctx context.Context, instanceID string) (*ServiceInstance, error) {
	// 需要遍历所有服务来查找实例
	services, err := r.GetAllServices(ctx)
	if err != nil {
		return nil, err
	}

	for _, instances := range services {
		for _, instance := range instances {
			if instance.InstanceID == instanceID {
				return &instance, nil
			}
		}
	}

	return nil, fmt.Errorf("instance not found: %s", instanceID)
}

// GetAllServices 获取所有服务
func (r *EtcdRegistry) GetAllServices(ctx context.Context) (map[string][]ServiceInstance, error) {
	prefix := fmt.Sprintf("%s/", r.rootPath)

	resp, err := r.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get all services: %w", err)
	}

	services := make(map[string][]ServiceInstance)
	for _, kv := range resp.Kvs {
		var instance ServiceInstance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			r.logger.Warn("Failed to unmarshal instance", zap.Error(err))
			continue
		}
		services[instance.ServiceName] = append(services[instance.ServiceName], instance)
	}

	return services, nil
}

// SubscribeInstanceChanges 订阅服务实例变更
func (r *EtcdRegistry) SubscribeInstanceChanges(ctx context.Context, serviceName string, listener InstanceChangeListener) error {
	r.listenersMutex.Lock()
	defer r.listenersMutex.Unlock()

	// 添加监听器
	r.listeners[serviceName] = append(r.listeners[serviceName], listener)

	// 如果是第一个监听器，启动 watch
	if len(r.listeners[serviceName]) == 1 {
		r.startWatch(serviceName)
	}

	// 立即触发一次，获取当前实例
	instances, err := r.DiscoverInstances(ctx, serviceName)
	if err == nil {
		listener(serviceName, instances)
	}

	return nil
}

// UnsubscribeInstanceChanges 取消订阅服务实例变更
func (r *EtcdRegistry) UnsubscribeInstanceChanges(ctx context.Context, serviceName string) error {
	r.listenersMutex.Lock()
	defer r.listenersMutex.Unlock()

	delete(r.listeners, serviceName)

	r.watchMutex.Lock()
	if cancel, exists := r.watchCancel[serviceName]; exists {
		cancel()
		delete(r.watchCancel, serviceName)
	}
	r.watchMutex.Unlock()

	return nil
}

// GetSubscribedServices 获取已订阅的服务列表
func (r *EtcdRegistry) GetSubscribedServices(ctx context.Context) ([]string, error) {
	r.listenersMutex.RLock()
	defer r.listenersMutex.RUnlock()

	services := make([]string, 0, len(r.listeners))
	for serviceName := range r.listeners {
		services = append(services, serviceName)
	}
	return services, nil
}

// CheckHealth 健康检查
func (r *EtcdRegistry) CheckHealth(ctx context.Context) error {
	_, err := r.client.Status(ctx, r.client.Endpoints()[0])
	return err
}

// Close 关闭资源
func (r *EtcdRegistry) Close() error {
	// 停止所有 watch
	r.watchMutex.Lock()
	for _, cancel := range r.watchCancel {
		cancel()
	}
	r.watchMutex.Unlock()

	// 注销实例
	if r.leaseID != 0 {
		r.client.Revoke(context.Background(), r.leaseID)
	}

	return r.client.Close()
}

// processKeepAlive 处理租约续约
func (r *EtcdRegistry) processKeepAlive() {
	for resp := range r.keepAliveResp {
		if resp == nil {
			r.logger.Error("Keep alive channel closed")
			return
		}
		// 可以在这里处理续约响应
	}
}

// startWatch 启动服务监听
func (r *EtcdRegistry) startWatch(serviceName string) {
	prefix := fmt.Sprintf("%s/%s/", r.rootPath, serviceName)
	ctx, cancel := context.WithCancel(context.Background())

	r.watchMutex.Lock()
	r.watchCancel[serviceName] = cancel
	r.watchMutex.Unlock()

	watchChan := r.client.Watch(ctx, prefix, clientv3.WithPrefix())

	go func() {
		defer cancel()
		for watchResp := range watchChan {
			if watchResp.Err() != nil {
				r.logger.Error("Watch error", zap.Error(watchResp.Err()))
				continue
			}

			// 获取最新的实例列表
			instances, err := r.DiscoverInstances(context.Background(), serviceName)
			if err != nil {
				r.logger.Error("Failed to get instances after watch event", zap.Error(err))
				continue
			}

			// 通知所有监听器
			r.listenersMutex.RLock()
			listeners := r.listeners[serviceName]
			r.listenersMutex.RUnlock()

			for _, listener := range listeners {
				go listener(serviceName, instances)
			}
		}
	}()
}
