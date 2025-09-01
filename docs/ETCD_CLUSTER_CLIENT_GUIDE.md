# etcd 集群客户端连接指南

## 🎯 概述

在分布式环境中，etcd 集群的客户端连接不应该写死 IP 地址，而应该采用多种灵活的配置策略。本文档详细介绍了不同部署场景下的最佳实践。

## 🏗️ 连接策略

### 1. 环境变量配置（推荐）

**适用场景**: 容器化部署、CI/CD 环境

```bash
# 设置环境变量
export ETCD_ENDPOINTS="etcd-1:2379,etcd-2:2379,etcd-3:2379"
export ETCD_CLUSTER_MODE="true"
export LOAD_BALANCER_TYPE="round_robin"
```

**优点**:
- ✅ 配置灵活，易于管理
- ✅ 支持不同环境使用不同配置
- ✅ 符合 12-Factor App 原则
- ✅ 便于自动化部署

### 2. 配置文件配置

**适用场景**: 传统部署、需要复杂配置

```yaml
# configs/etcd_cluster_example.yaml
registry:
  type: "etcd"
  etcd:
    endpoints:
      - "192.168.1.10:2379"
      - "192.168.1.11:2379"
      - "192.168.1.12:2379"
    clusterMode: true
    retryPolicy:
      maxRetries: 3
      retryInterval: 1s
      backoffFactor: 2.0
```

**优点**:
- ✅ 配置结构清晰
- ✅ 支持复杂配置
- ✅ 版本控制友好
- ✅ 便于配置验证

### 3. 服务发现动态配置

**适用场景**: Kubernetes、Consul 等环境

```go
// 通过服务发现获取 etcd 端点
func discoverEtcdEndpoints() []string {
    // 从 Kubernetes Service 或 Consul 获取
    return []string{
        "etcd-service-1:2379",
        "etcd-service-2:2379", 
        "etcd-service-3:2379",
    }
}
```

**优点**:
- ✅ 完全动态化
- ✅ 自动故障转移
- ✅ 支持服务注册发现
- ✅ 高可用性

### 4. DNS 解析配置

**适用场景**: 云环境、负载均衡器

```bash
# 使用 DNS 解析
export ETCD_ENDPOINTS="etcd-cluster.example.com:2379"
```

**优点**:
- ✅ 利用 DNS 负载均衡
- ✅ 配置简单
- ✅ 支持地理分布
- ✅ 自动故障转移

## 🚀 不同部署场景

### Docker Compose 集群

```yaml
# docker-compose.cluster.yml
services:
  etcd-1:
    ports:
      - "2379:2379"
  etcd-2:
    ports:
      - "2381:2379"
  etcd-3:
    ports:
      - "2383:2379"
```

**客户端配置**:
```bash
# 容器内通信
export ETCD_ENDPOINTS="etcd-1:2379,etcd-2:2379,etcd-3:2379"

# 外部访问
export ETCD_ENDPOINTS="localhost:2379,localhost:2381,localhost:2383"
```

### Kubernetes 集群

```yaml
# etcd-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: etcd-cluster
spec:
  ports:
  - port: 2379
    targetPort: 2379
  selector:
    app: etcd
```

**客户端配置**:
```bash
# 使用 Kubernetes Service
export ETCD_ENDPOINTS="etcd-cluster:2379"
```

### 物理机集群

```bash
# 使用固定 IP
export ETCD_ENDPOINTS="192.168.1.10:2379,192.168.1.11:2379,192.168.1.12:2379"
```

### 云环境集群

```bash
# 使用负载均衡器
export ETCD_ENDPOINTS="etcd-lb.example.com:2379"

# 或使用多个负载均衡器
export ETCD_ENDPOINTS="etcd-lb-1.example.com:2379,etcd-lb-2.example.com:2379"
```

## 🔧 配置示例

### 完整配置示例

```go
package main

import (
    "github.com/roc/roc-im-server/tools/serviceDiscovery"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewDevelopment()
    
    // 创建配置
    config := servicediscovery.DefaultServiceDiscoveryConfig()
    
    // 从环境变量加载
    config.LoadFromEnv()
    
    // 或者手动设置
    config.Registry.Etcd.Endpoints = []string{
        "etcd-1:2379",
        "etcd-2:2379", 
        "etcd-3:2379",
    }
    config.Registry.Etcd.ClusterMode = true
    config.Registry.Etcd.RetryPolicy.MaxRetries = 3
    
    // 创建服务发现客户端
    registry, err := config.CreateServiceRegistry(logger)
    if err != nil {
        logger.Fatal("创建服务注册中心失败", zap.Error(err))
    }
    
    // 使用客户端
    // ...
}
```

### 环境变量配置

```bash
#!/bin/bash
# 设置 etcd 集群环境变量

# 开发环境
export ETCD_ENDPOINTS="localhost:2379"
export ETCD_CLUSTER_MODE="false"

# 测试环境
export ETCD_ENDPOINTS="etcd-test-1:2379,etcd-test-2:2379,etcd-test-3:2379"
export ETCD_CLUSTER_MODE="true"

# 生产环境
export ETCD_ENDPOINTS="etcd-prod-1:2379,etcd-prod-2:2379,etcd-prod-3:2379"
export ETCD_CLUSTER_MODE="true"
export ETCD_USERNAME="etcd-user"
export ETCD_PASSWORD="etcd-password"
```

## 🛡️ 高可用性配置

### 1. 重试策略

```yaml
retryPolicy:
  maxRetries: 5
  retryInterval: 1s
  backoffFactor: 2.0
```

### 2. 连接池配置

```yaml
connectionPool:
  maxConnections: 100
  maxIdleConnections: 10
  connectionTimeout: 5s
  idleTimeout: 30s
```

### 3. 健康检查

```go
// 定期健康检查
func healthCheck(registry serviceregistry.ServiceRegistry) {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        if err := registry.CheckHealth(context.Background()); err != nil {
            log.Error("etcd 健康检查失败", zap.Error(err))
            // 触发重连逻辑
        }
    }
}
```

## 🔍 故障排除

### 1. 连接失败

```bash
# 检查网络连通性
telnet etcd-1 2379
telnet etcd-2 2379
telnet etcd-3 2379

# 检查 DNS 解析
nslookup etcd-cluster.example.com

# 检查防火墙
iptables -L | grep 2379
```

### 2. 集群状态检查

```bash
# 检查集群成员
etcdctl member list

# 检查集群健康状态
etcdctl endpoint health

# 检查集群状态
etcdctl endpoint status
```

### 3. 日志分析

```bash
# 查看 etcd 日志
docker logs etcd-1
docker logs etcd-2
docker logs etcd-3

# 查看客户端日志
tail -f /var/log/app/etcd-client.log
```

## 📊 监控指标

### 1. 连接指标

- 连接成功率
- 连接延迟
- 连接池使用率
- 重试次数

### 2. 集群指标

- 集群成员状态
- 领导者选举
- 数据同步状态
- 存储使用量

### 3. 性能指标

- 读写延迟
- 吞吐量
- 错误率
- 资源使用率

## 🚨 最佳实践

### 1. 配置管理

- ✅ 使用环境变量进行配置
- ✅ 避免硬编码 IP 地址
- ✅ 支持配置热更新
- ✅ 配置验证和默认值

### 2. 高可用性

- ✅ 配置多个 etcd 端点
- ✅ 实现自动重试机制
- ✅ 监控集群健康状态
- ✅ 实现故障转移

### 3. 安全性

- ✅ 使用 TLS 加密
- ✅ 配置认证机制
- ✅ 限制网络访问
- ✅ 定期更新证书

### 4. 性能优化

- ✅ 使用连接池
- ✅ 配置合适的超时时间
- ✅ 实现缓存机制
- ✅ 监控性能指标

## 📝 总结

etcd 集群客户端连接应该采用灵活的配置策略，避免写死 IP 地址。根据不同的部署环境选择合适的配置方式：

1. **容器化环境**: 使用环境变量 + 服务发现
2. **传统部署**: 使用配置文件 + DNS 解析
3. **云环境**: 使用负载均衡器 + 服务发现
4. **混合环境**: 支持多种配置方式

通过合理的配置策略，可以实现高可用、易维护的 etcd 集群客户端连接。
