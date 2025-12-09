package servicediscovery

import (
	"os"
	"strings"
	"time"

	"github.com/rhp-QE/roc-im-server/tools/loadbalancer"
	serviceregistry "github.com/rhp-QE/roc-im-server/tools/serviceRegistry"
	"go.uber.org/zap"
)

// ServiceDiscoveryConfig 服务发现配置
type ServiceDiscoveryConfig struct {
	// 注册中心配置
	Registry RegistryConfig `yaml:"registry"`

	// 负载均衡配置
	LoadBalancer LoadBalancerConfig `yaml:"loadBalancer"`

	// 缓存配置
	Cache CacheConfig `yaml:"cache"`
}

// RegistryConfig 注册中心配置
type RegistryConfig struct {
	Type      string       `yaml:"type"` // etcd, consul, zookeeper
	Etcd      EtcdConfig   `yaml:"etcd"`
	Consul    ConsulConfig `yaml:"consul"`
	ZooKeeper ZKConfig     `yaml:"zookeeper"`
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

// ConsulConfig Consul配置
type ConsulConfig struct {
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Token    string `yaml:"token"`
}

// ZKConfig ZooKeeper配置
type ZKConfig struct {
	Servers        []string      `yaml:"servers"`
	RootPath       string        `yaml:"rootPath"`
	SessionTimeout time.Duration `yaml:"sessionTimeout"`
}

// LoadBalancerConfig 负载均衡配置
type LoadBalancerConfig struct {
	Type string `yaml:"type"` // round_robin, random, consistent_hash, weighted_random, least_active
}

// CacheConfig 缓存配置
type CacheConfig struct {
	TTL           time.Duration `yaml:"ttl"`           // 缓存过期时间
	RefreshPeriod time.Duration `yaml:"refreshPeriod"` // 刷新周期
}

// DefaultServiceDiscoveryConfig 默认配置
func DefaultServiceDiscoveryConfig() *ServiceDiscoveryConfig {
	return &ServiceDiscoveryConfig{
		Registry: RegistryConfig{
			Type: "etcd",
			Etcd: EtcdConfig{
				Endpoints:   []string{"localhost:2379"},
				RootPath:    "/roc-im-server/services",
				DialTimeout: 5 * time.Second,
				ClusterMode: false,
				RetryPolicy: RetryPolicy{
					MaxRetries:    3,
					RetryInterval: 1 * time.Second,
					BackoffFactor: 2.0,
				},
			},
		},
		LoadBalancer: LoadBalancerConfig{
			Type: "round_robin",
		},
		Cache: CacheConfig{
			TTL:           30 * time.Second,
			RefreshPeriod: 10 * time.Second,
		},
	}
}

// LoadFromEnv 从环境变量加载配置
func (c *ServiceDiscoveryConfig) LoadFromEnv() {
	// Etcd 配置
	if endpoints := os.Getenv("ETCD_ENDPOINTS"); endpoints != "" {
		// 支持多个端点，用逗号分隔
		c.Registry.Etcd.Endpoints = strings.Split(endpoints, ",")
	}
	if username := os.Getenv("ETCD_USERNAME"); username != "" {
		c.Registry.Etcd.Username = username
	}
	if password := os.Getenv("ETCD_PASSWORD"); password != "" {
		c.Registry.Etcd.Password = password
	}
	if rootPath := os.Getenv("ETCD_ROOT_PATH"); rootPath != "" {
		c.Registry.Etcd.RootPath = rootPath
	}

	// 集群模式配置
	if clusterMode := os.Getenv("ETCD_CLUSTER_MODE"); clusterMode != "" {
		c.Registry.Etcd.ClusterMode = clusterMode == "true"
	}

	// 负载均衡配置
	if lbType := os.Getenv("LOAD_BALANCER_TYPE"); lbType != "" {
		c.LoadBalancer.Type = lbType
	}
}

// CreateServiceRegistry 创建服务注册中心
func (c *ServiceDiscoveryConfig) CreateServiceRegistry(logger *zap.Logger) (serviceregistry.ServiceRegistry, error) {
	switch c.Registry.Type {
	case "etcd":
		config := serviceregistry.EtcdConfig{
			Endpoints:   c.Registry.Etcd.Endpoints,
			Username:    c.Registry.Etcd.Username,
			Password:    c.Registry.Etcd.Password,
			RootPath:    c.Registry.Etcd.RootPath,
			DialTimeout: c.Registry.Etcd.DialTimeout,
		}
		return serviceregistry.NewEtcdRegistry(config, logger)
	default:
		// 默认使用 Etcd
		config := serviceregistry.EtcdConfig{
			Endpoints:   c.Registry.Etcd.Endpoints,
			Username:    c.Registry.Etcd.Username,
			Password:    c.Registry.Etcd.Password,
			RootPath:    c.Registry.Etcd.RootPath,
			DialTimeout: c.Registry.Etcd.DialTimeout,
		}
		return serviceregistry.NewEtcdRegistry(config, logger)
	}
}

// CreateLoadBalancer 创建负载均衡器
func (c *ServiceDiscoveryConfig) CreateLoadBalancer() loadbalancer.LoadBalancer {
	factory := &loadbalancer.DefaultLoadBalancerFactory{}

	switch c.LoadBalancer.Type {
	case "round_robin":
		return factory.CreateLoadBalancer(loadbalancer.RoundRobin)
	case "random":
		return factory.CreateLoadBalancer(loadbalancer.Random)
	case "consistent_hash":
		return factory.CreateLoadBalancer(loadbalancer.ConsistentHash)
	case "weighted_random":
		return factory.CreateLoadBalancer(loadbalancer.WeightedRandom)
	case "least_active":
		return factory.CreateLoadBalancer(loadbalancer.LeastActive)
	default:
		return factory.CreateLoadBalancer(loadbalancer.RoundRobin)
	}
}
