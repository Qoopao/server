package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/roc/roc-im-server/tools/loadbalancer"
	servicediscovery "github.com/roc/roc-im-server/tools/serviceDiscovery"
	serviceregistry "github.com/roc/roc-im-server/tools/serviceRegistry"
	"go.uber.org/zap"
)

// ServiceDiscovery_main 服务发现示例主函数
func ServiceDiscovery_main() {
	// 创建日志
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 示例1: 基本的服务注册和发现
	fmt.Println("=== 示例1: 基本服务注册和发现 ===")
	basicExample(logger)

	// 示例2: 不同负载均衡策略
	fmt.Println("\n=== 示例2: 不同负载均衡策略 ===")
	loadBalancerExample(logger)

	// 示例3: 服务实例变更监听
	fmt.Println("\n=== 示例3: 服务实例变更监听 ===")
	watchExample(logger)
}

// basicExample 基本的服务注册和发现示例
func basicExample(logger *zap.Logger) {
	// 创建 Etcd 注册中心
	config := serviceregistry.EtcdConfig{
		Endpoints:   []string{"localhost:2379"},
		RootPath:    "/example/services",
		DialTimeout: 5 * time.Second,
	}

	registry, err := serviceregistry.NewEtcdRegistry(config, logger)
	if err != nil {
		log.Printf("Failed to create registry: %v", err)
		return
	}
	defer registry.Close()

	// 注册服务实例
	instance := &serviceregistry.ServiceInstance{
		InstanceID:  "example-service-001",
		ServiceName: "example-service",
		Host:        "192.168.1.100",
		Port:        8080,
		Status:      serviceregistry.StatusHealthy,
		Weight:      1,
		Metadata: map[string]string{
			"version": "1.0.0",
			"zone":    "zone-a",
		},
	}

	if err := registry.RegisterInstance(context.Background(), instance); err != nil {
		log.Printf("Failed to register instance: %v", err)
		return
	}

	// 发现服务实例
	instances, err := registry.DiscoverInstances(context.Background(), "example-service")
	if err != nil {
		log.Printf("Failed to discover instances: %v", err)
		return
	}

	fmt.Printf("发现 %d 个服务实例:\n", len(instances))
	for _, inst := range instances {
		fmt.Printf("  - %s:%d (权重: %d, 状态: %v)\n",
			inst.Host, inst.Port, inst.Weight, inst.Status)
	}
}

// loadBalancerExample 不同负载均衡策略示例
func loadBalancerExample(logger *zap.Logger) {
	// 模拟服务实例
	instances := []serviceregistry.ServiceInstance{
		{InstanceID: "inst-1", ServiceName: "test-service", Host: "192.168.1.10", Port: 8080, Status: serviceregistry.StatusHealthy, Weight: 1},
		{InstanceID: "inst-2", ServiceName: "test-service", Host: "192.168.1.11", Port: 8080, Status: serviceregistry.StatusHealthy, Weight: 2},
		{InstanceID: "inst-3", ServiceName: "test-service", Host: "192.168.1.12", Port: 8080, Status: serviceregistry.StatusHealthy, Weight: 3},
	}

	factory := &loadbalancer.DefaultLoadBalancerFactory{}

	// 测试不同的负载均衡策略
	strategies := []loadbalancer.LoadBalancerType{
		loadbalancer.RoundRobin,
		loadbalancer.Random,
		loadbalancer.WeightedRandom,
		loadbalancer.ConsistentHash,
	}

	for _, strategy := range strategies {
		lb := factory.CreateLoadBalancer(strategy)
		fmt.Printf("\n%s 策略测试:\n", lb.GetName())

		// 进行5次选择
		for i := 0; i < 5; i++ {
			var selected *serviceregistry.ServiceInstance
			var err error

			if strategy == loadbalancer.ConsistentHash {
				// 一致性哈希需要提供键
				selected, err = lb.Select(instances, fmt.Sprintf("user_%d", i))
			} else {
				selected, err = lb.Select(instances)
			}

			if err != nil {
				log.Printf("选择失败: %v", err)
				continue
			}

			fmt.Printf("  选择: %s:%d (权重: %d)\n",
				selected.Host, selected.Port, selected.Weight)
		}
	}
}

// watchExample 服务实例变更监听示例
func watchExample(logger *zap.Logger) {
	config := serviceregistry.EtcdConfig{
		Endpoints:   []string{"localhost:2379"},
		RootPath:    "/example/services",
		DialTimeout: 5 * time.Second,
	}

	registry, err := serviceregistry.NewEtcdRegistry(config, logger)
	if err != nil {
		log.Printf("Failed to create registry: %v", err)
		return
	}
	defer registry.Close()

	// 创建服务发现客户端
	discoveryClient := servicediscovery.NewServiceDiscoveryClient(servicediscovery.Config{
		Registry:     registry,
		LoadBalancer: loadbalancer.NewRoundRobinLoadBalancer(),
		Logger:       logger,
	})
	defer discoveryClient.Close()

	// 订阅服务变更
	err = registry.SubscribeInstanceChanges(context.Background(), "watch-service", func(serviceName string, instances []serviceregistry.ServiceInstance) {
		fmt.Printf("服务 %s 实例发生变更，当前实例数: %d\n", serviceName, len(instances))
		for _, inst := range instances {
			fmt.Printf("  - %s:%d\n", inst.Host, inst.Port)
		}
	})

	if err != nil {
		log.Printf("Failed to subscribe: %v", err)
		return
	}

	// 模拟注册新实例
	instance1 := &serviceregistry.ServiceInstance{
		InstanceID:  "watch-service-001",
		ServiceName: "watch-service",
		Host:        "192.168.1.200",
		Port:        9000,
		Status:      serviceregistry.StatusHealthy,
		Weight:      1,
	}

	fmt.Println("注册第一个实例...")
	registry.RegisterInstance(context.Background(), instance1)

	// 等待一下
	time.Sleep(2 * time.Second)

	// 注册第二个实例
	instance2 := &serviceregistry.ServiceInstance{
		InstanceID:  "watch-service-002",
		ServiceName: "watch-service",
		Host:        "192.168.1.201",
		Port:        9000,
		Status:      serviceregistry.StatusHealthy,
		Weight:      1,
	}

	fmt.Println("注册第二个实例...")
	registry.RegisterInstance(context.Background(), instance2)

	// 等待一下再退出
	time.Sleep(2 * time.Second)
	fmt.Println("示例结束")
}
