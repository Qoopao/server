package loadbalancer

import (
	"errors"
	"sync/atomic"

	serviceregistry "github.com/rhp-QE/roc-im-server/tools/serviceRegistry"
)

// RoundRobinLoadBalancer 轮询负载均衡器
type RoundRobinLoadBalancer struct {
	counter uint64
}

// NewRoundRobinLoadBalancer 创建轮询负载均衡器
func NewRoundRobinLoadBalancer() *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{}
}

// Select 轮询选择实例
func (lb *RoundRobinLoadBalancer) Select(instances []serviceregistry.ServiceInstance, key ...string) (*serviceregistry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, errors.New("no available instances")
	}

	// 过滤健康的实例
	healthyInstances := make([]serviceregistry.ServiceInstance, 0, len(instances))
	for _, instance := range instances {
		if instance.Status == serviceregistry.StatusHealthy {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	if len(healthyInstances) == 0 {
		return nil, errors.New("no healthy instances available")
	}

	// 原子操作获取下一个索引
	index := atomic.AddUint64(&lb.counter, 1) % uint64(len(healthyInstances))
	return &healthyInstances[index], nil
}

// GetName 获取负载均衡器名称
func (lb *RoundRobinLoadBalancer) GetName() string {
	return "RoundRobin"
}
