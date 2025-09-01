package loadbalancer

import (
	"errors"
	"math/rand"
	"strconv"
	"time"

	serviceregistry "github.com/roc/roc-im-server/tools/serviceRegistry"
)

// LeastActiveLoadBalancer 最少活跃调用数负载均衡器
type LeastActiveLoadBalancer struct {
	rand *rand.Rand
}

// NewLeastActiveLoadBalancer 创建最少活跃调用数负载均衡器
func NewLeastActiveLoadBalancer() *LeastActiveLoadBalancer {
	return &LeastActiveLoadBalancer{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select 选择活跃调用数最少的实例
func (lb *LeastActiveLoadBalancer) Select(instances []serviceregistry.ServiceInstance, key ...string) (*serviceregistry.ServiceInstance, error) {
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

	// 找到最少活跃调用数
	minActive := -1
	var candidates []serviceregistry.ServiceInstance

	for _, instance := range healthyInstances {
		// 从元数据中获取活跃调用数，默认为0
		activeStr := instance.Metadata["active_count"]
		active := 0
		if activeStr != "" {
			if parsed, err := strconv.Atoi(activeStr); err == nil {
				active = parsed
			}
		}

		if minActive == -1 || active < minActive {
			minActive = active
			candidates = []serviceregistry.ServiceInstance{instance}
		} else if active == minActive {
			candidates = append(candidates, instance)
		}
	}

	// 如果有多个相同最少活跃数的实例，随机选择一个
	if len(candidates) == 1 {
		return &candidates[0], nil
	}

	index := lb.rand.Intn(len(candidates))
	return &candidates[index], nil
}

// GetName 获取负载均衡器名称
func (lb *LeastActiveLoadBalancer) GetName() string {
	return "LeastActive"
}
