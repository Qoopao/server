package loadbalancer

import (
	"errors"
	"math/rand"
	"time"

	serviceregistry "github.com/rhp-QE/roc-im-server/tools/serviceRegistry"
)

// WeightedRandomLoadBalancer 加权随机负载均衡器
type WeightedRandomLoadBalancer struct {
	rand *rand.Rand
}

// NewWeightedRandomLoadBalancer 创建加权随机负载均衡器
func NewWeightedRandomLoadBalancer() *WeightedRandomLoadBalancer {
	return &WeightedRandomLoadBalancer{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select 根据权重随机选择实例
func (lb *WeightedRandomLoadBalancer) Select(instances []serviceregistry.ServiceInstance, key ...string) (*serviceregistry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, errors.New("no available instances")
	}

	// 过滤健康的实例并计算总权重
	healthyInstances := make([]serviceregistry.ServiceInstance, 0, len(instances))
	totalWeight := 0

	for _, instance := range instances {
		if instance.Status == serviceregistry.StatusHealthy {
			weight := instance.Weight
			if weight <= 0 {
				weight = 1 // 默认权重为1
			}
			healthyInstances = append(healthyInstances, instance)
			totalWeight += weight
		}
	}

	if len(healthyInstances) == 0 {
		return nil, errors.New("no healthy instances available")
	}

	if totalWeight == 0 {
		// 如果所有权重都是0，使用普通随机选择
		index := lb.rand.Intn(len(healthyInstances))
		return &healthyInstances[index], nil
	}

	// 生成随机数
	randomWeight := lb.rand.Intn(totalWeight)

	// 根据权重选择实例
	currentWeight := 0
	for _, instance := range healthyInstances {
		weight := instance.Weight
		if weight <= 0 {
			weight = 1
		}
		currentWeight += weight
		if randomWeight < currentWeight {
			return &instance, nil
		}
	}

	// 理论上不应该到达这里，但作为保险
	return &healthyInstances[len(healthyInstances)-1], nil
}

// GetName 获取负载均衡器名称
func (lb *WeightedRandomLoadBalancer) GetName() string {
	return "WeightedRandom"
}
