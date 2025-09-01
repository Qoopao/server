package loadbalancer

import (
	"errors"
	"math/rand"
	"time"

	serviceregistry "github.com/roc/roc-im-server/tools/serviceRegistry"
)

// RandomLoadBalancer 随机负载均衡器
type RandomLoadBalancer struct {
	rand *rand.Rand
}

// NewRandomLoadBalancer 创建随机负载均衡器
func NewRandomLoadBalancer() *RandomLoadBalancer {
	return &RandomLoadBalancer{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select 随机选择实例
func (lb *RandomLoadBalancer) Select(instances []serviceregistry.ServiceInstance, key ...string) (*serviceregistry.ServiceInstance, error) {
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

	// 随机选择
	index := lb.rand.Intn(len(healthyInstances))
	return &healthyInstances[index], nil
}

// GetName 获取负载均衡器名称
func (lb *RandomLoadBalancer) GetName() string {
	return "Random"
}
