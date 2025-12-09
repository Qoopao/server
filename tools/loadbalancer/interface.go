package loadbalancer

import (
	serviceregistry "github.com/rhp-QE/roc-im-server/tools/serviceRegistry"
)

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	// Select 从可用实例中选择一个
	Select(instances []serviceregistry.ServiceInstance, key ...string) (*serviceregistry.ServiceInstance, error)
	// GetName 获取负载均衡器名称
	GetName() string
}

// LoadBalancerType 负载均衡类型
type LoadBalancerType string

const (
	RoundRobin     LoadBalancerType = "round_robin"
	Random         LoadBalancerType = "random"
	ConsistentHash LoadBalancerType = "consistent_hash"
	WeightedRandom LoadBalancerType = "weighted_random"
	LeastActive    LoadBalancerType = "least_active"
)

// LoadBalancerFactory 负载均衡器工厂
type LoadBalancerFactory interface {
	CreateLoadBalancer(lbType LoadBalancerType) LoadBalancer
}

// DefaultLoadBalancerFactory 默认负载均衡器工厂
type DefaultLoadBalancerFactory struct{}

// CreateLoadBalancer 创建负载均衡器
func (f *DefaultLoadBalancerFactory) CreateLoadBalancer(lbType LoadBalancerType) LoadBalancer {
	switch lbType {
	case RoundRobin:
		return NewRoundRobinLoadBalancer()
	case Random:
		return NewRandomLoadBalancer()
	case ConsistentHash:
		return NewConsistentHashLoadBalancer()
	case WeightedRandom:
		return NewWeightedRandomLoadBalancer()
	case LeastActive:
		return NewLeastActiveLoadBalancer()
	default:
		return NewRoundRobinLoadBalancer() // 默认使用轮询
	}
}
