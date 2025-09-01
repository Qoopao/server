package examples

import (
	"context"
	"fmt"
	"os"
	"strings"

	servicediscovery "github.com/roc/roc-im-server/tools/serviceDiscovery"
	"go.uber.org/zap"
)

// EtcdClusterClientExample_main 演示etcd集群客户端连接
func EtcdClusterClientExample_main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	logger.Info("开始 etcd 集群客户端连接示例")

	// 方法1: 通过环境变量配置集群端点
	example1_EnvConfig(logger)

	// 方法2: 通过配置文件配置集群端点
	example2_FileConfig(logger)

	// 方法3: 通过服务发现动态获取集群端点
	example3_ServiceDiscovery(logger)

	// 方法4: 通过DNS解析获取集群端点
	example4_DNSResolution(logger)
}

// 方法1: 通过环境变量配置集群端点
func example1_EnvConfig(logger *zap.Logger) {
	logger.Info("=== 方法1: 环境变量配置 ===")

	// 设置环境变量
	os.Setenv("ETCD_ENDPOINTS", "etcd-1:2379,etcd-2:2379,etcd-3:2379")
	os.Setenv("ETCD_CLUSTER_MODE", "true")
	os.Setenv("LOAD_BALANCER_TYPE", "round_robin")

	// 创建配置
	config := servicediscovery.DefaultServiceDiscoveryConfig()
	config.LoadFromEnv()

	logger.Info("集群端点配置",
		zap.Strings("endpoints", config.Registry.Etcd.Endpoints),
		zap.Bool("clusterMode", config.Registry.Etcd.ClusterMode))

	// 创建服务发现客户端
	registry, err := config.CreateServiceRegistry(logger)
	if err != nil {
		logger.Error("创建服务注册中心失败", zap.Error(err))
		return
	}

	// 测试连接
	ctx := context.Background()
	if err := registry.CheckHealth(ctx); err != nil {
		logger.Error("etcd 集群连接失败", zap.Error(err))
	} else {
		logger.Info("etcd 集群连接成功")
	}
}

// 方法2: 通过配置文件配置集群端点
func example2_FileConfig(logger *zap.Logger) {
	logger.Info("=== 方法2: 配置文件配置 ===")

	// 这里可以加载YAML配置文件
	// 示例配置内容已在 configs/etcd_cluster_example.yaml 中定义

	// 模拟从配置文件读取的配置
	config := servicediscovery.DefaultServiceDiscoveryConfig()
	config.Registry.Etcd.Endpoints = []string{
		"192.168.1.10:2379",
		"192.168.1.11:2379",
		"192.168.1.12:2379",
	}
	config.Registry.Etcd.ClusterMode = true
	config.Registry.Etcd.RetryPolicy.MaxRetries = 5

	logger.Info("从配置文件加载的集群配置",
		zap.Strings("endpoints", config.Registry.Etcd.Endpoints),
		zap.Int("maxRetries", config.Registry.Etcd.RetryPolicy.MaxRetries))

	// 创建服务发现客户端
	registry, err := config.CreateServiceRegistry(logger)
	if err != nil {
		logger.Error("创建服务注册中心失败", zap.Error(err))
		return
	}

	// 测试连接
	ctx := context.Background()
	if err := registry.CheckHealth(ctx); err != nil {
		logger.Error("etcd 集群连接失败", zap.Error(err))
	} else {
		logger.Info("etcd 集群连接成功")
	}
}

// 方法3: 通过服务发现动态获取集群端点
func example3_ServiceDiscovery(logger *zap.Logger) {
	logger.Info("=== 方法3: 服务发现动态配置 ===")

	// 使用服务发现机制动态获取etcd集群端点
	// 这通常用于Kubernetes环境或其他服务发现平台

	// 示例：从Consul或Kubernetes服务发现获取etcd端点
	etcdEndpoints := discoverEtcdEndpoints(logger)

	config := servicediscovery.DefaultServiceDiscoveryConfig()
	config.Registry.Etcd.Endpoints = etcdEndpoints
	config.Registry.Etcd.ClusterMode = true

	logger.Info("通过服务发现获取的集群端点",
		zap.Strings("endpoints", config.Registry.Etcd.Endpoints))

	// 创建服务发现客户端
	registry, err := config.CreateServiceRegistry(logger)
	if err != nil {
		logger.Error("创建服务注册中心失败", zap.Error(err))
		return
	}

	// 测试连接
	ctx := context.Background()
	if err := registry.CheckHealth(ctx); err != nil {
		logger.Error("etcd 集群连接失败", zap.Error(err))
	} else {
		logger.Info("etcd 集群连接成功")
	}
}

// 方法4: 通过DNS解析获取集群端点
func example4_DNSResolution(logger *zap.Logger) {
	logger.Info("=== 方法4: DNS解析配置 ===")

	// 使用DNS解析获取etcd集群端点
	// 适用于使用DNS服务发现的环境

	// 示例：从DNS解析获取etcd端点
	etcdEndpoints := resolveEtcdEndpoints(logger)

	config := servicediscovery.DefaultServiceDiscoveryConfig()
	config.Registry.Etcd.Endpoints = etcdEndpoints
	config.Registry.Etcd.ClusterMode = true

	logger.Info("通过DNS解析获取的集群端点",
		zap.Strings("endpoints", config.Registry.Etcd.Endpoints))

	// 创建服务发现客户端
	registry, err := config.CreateServiceRegistry(logger)
	if err != nil {
		logger.Error("创建服务注册中心失败", zap.Error(err))
		return
	}

	// 测试连接
	ctx := context.Background()
	if err := registry.CheckHealth(ctx); err != nil {
		logger.Error("etcd 集群连接失败", zap.Error(err))
	} else {
		logger.Info("etcd 集群连接成功")
	}
}

// 模拟服务发现获取etcd端点
func discoverEtcdEndpoints(logger *zap.Logger) []string {
	// 这里可以集成Consul、Kubernetes服务发现等
	// 示例实现
	logger.Info("通过服务发现获取etcd端点")

	// 模拟从服务发现获取的端点
	return []string{
		"etcd-service-1:2379",
		"etcd-service-2:2379",
		"etcd-service-3:2379",
	}
}

// 模拟DNS解析获取etcd端点
func resolveEtcdEndpoints(logger *zap.Logger) []string {
	// 这里可以集成DNS解析
	// 示例实现
	logger.Info("通过DNS解析获取etcd端点")

	// 模拟DNS解析结果
	return []string{
		"etcd-cluster.example.com:2379",
		"etcd-cluster-1.example.com:2379",
		"etcd-cluster-2.example.com:2379",
	}
}

// 演示不同部署场景的连接配置
func demonstrateDeploymentScenarios() {
	fmt.Println("\n=== 不同部署场景的连接配置 ===")

	scenarios := []struct {
		name        string
		description string
		endpoints   []string
		envVars     map[string]string
	}{
		{
			name:        "Docker Compose 集群",
			description: "使用容器名称进行通信",
			endpoints:   []string{"etcd-1:2379", "etcd-2:2379", "etcd-3:2379"},
			envVars: map[string]string{
				"ETCD_ENDPOINTS": "etcd-1:2379,etcd-2:2379,etcd-3:2379",
			},
		},
		{
			name:        "Kubernetes 集群",
			description: "使用Kubernetes服务发现",
			endpoints:   []string{"etcd-cluster:2379"},
			envVars: map[string]string{
				"ETCD_ENDPOINTS": "etcd-cluster:2379",
			},
		},
		{
			name:        "物理机集群",
			description: "使用固定IP地址",
			endpoints:   []string{"192.168.1.10:2379", "192.168.1.11:2379", "192.168.1.12:2379"},
			envVars: map[string]string{
				"ETCD_ENDPOINTS": "192.168.1.10:2379,192.168.1.11:2379,192.168.1.12:2379",
			},
		},
		{
			name:        "云环境集群",
			description: "使用负载均衡器端点",
			endpoints:   []string{"etcd-lb.example.com:2379"},
			envVars: map[string]string{
				"ETCD_ENDPOINTS": "etcd-lb.example.com:2379",
			},
		},
	}

	for _, scenario := range scenarios {
		fmt.Printf("\n📋 %s\n", scenario.name)
		fmt.Printf("   描述: %s\n", scenario.description)
		fmt.Printf("   端点: %s\n", strings.Join(scenario.endpoints, ", "))
		fmt.Printf("   环境变量: ETCD_ENDPOINTS=%s\n", scenario.envVars["ETCD_ENDPOINTS"])
	}
}
