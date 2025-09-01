# 服务发现和负载均衡解决方案

## 概述

本项目实现了一个完整的服务发现和负载均衡解决方案，支持分布式环境下的服务注册、发现和智能路由。

## 主要特性

- **多种注册中心支持**: 当前支持 Etcd，可扩展支持 Consul、ZooKeeper
- **多种负载均衡策略**: 轮询、随机、一致性哈希、加权随机、最少活跃调用数
- **服务实例健康检查**: 自动监控服务实例状态
- **实时服务发现**: 支持服务实例变更的实时通知
- **连接池管理**: 自动管理和复用服务连接

## 架构组件

### 1. 服务注册中心 (Service Registry)

#### 接口定义
```go
type ServiceRegistry interface {
    RegisterInstance(instance *ServiceInstance) error
    DeregisterInstance(instanceID string) error
    DiscoverInstances(serviceName string) ([]ServiceInstance, error)
    SubscribeInstanceChanges(serviceName string, listener InstanceChangeListener) error
    // ...
}
```

#### Etcd 实现
- 基于 Etcd v3 客户端
- 支持服务实例的租约续约
- 支持实时监听服务变更

### 2. 负载均衡器 (Load Balancer)

#### 支持的策略

1. **轮询 (Round Robin)**
   - 按顺序依次分配请求
   - 适用于服务实例性能相近的场景

2. **随机 (Random)**
   - 随机选择服务实例
   - 简单有效，适用于大多数场景

3. **一致性哈希 (Consistent Hash)**
   - 基于请求键进行哈希计算
   - 适用于需要会话一致性的场景

4. **加权随机 (Weighted Random)**
   - 基于权重进行随机选择
   - 适用于服务实例性能不同的场景

5. **最少活跃调用数 (Least Active)**
   - 选择当前活跃请求最少的实例
   - 适用于响应时间敏感的场景

### 3. 服务发现客户端 (Service Discovery Client)

- 封装服务注册中心和负载均衡器
- 提供服务实例缓存
- 支持连接池管理
- 支持服务实例变更的自动更新

## 使用指南

### 1. 环境准备

首先需要启动 Etcd 服务：

```bash
# 使用 Docker 启动 Etcd
docker run -d \
  --name etcd \
  -p 2379:2379 \
  -p 2380:2380 \
  quay.io/coreos/etcd:v3.5.0 \
  etcd \
  --advertise-client-urls http://0.0.0.0:2379 \
  --listen-client-urls http://0.0.0.0:2379 \
  --listen-peer-urls http://0.0.0.0:2380 \
  --initial-advertise-peer-urls http://0.0.0.0:2380 \
  --initial-cluster default=http://0.0.0.0:2380
```

### 2. 服务端配置

#### 消息服务配置

消息服务已经集成了服务注册功能，启动时会自动注册到 Etcd：

```bash
# 设置环境变量
export ETCD_ENDPOINTS=localhost:2379
export MSG_SERVICE_PORT=10100

# 启动消息服务
go run main.go
```

#### 会话服务配置

会话服务也已经集成了服务注册功能：

```bash
# 设置环境变量
export ETCD_ENDPOINTS=localhost:2379
export CONVERSATION_SERVICE_PORT=10200

# 启动会话服务
go run main.go
```

支持的环境变量：
- `ETCD_ENDPOINTS`: Etcd 服务地址
- `MSG_SERVICE_PORT`: 消息服务端口
- `CONVERSATION_SERVICE_PORT`: 会话服务端口
- `ETCD_USERNAME`: Etcd 用户名（可选）
- `ETCD_PASSWORD`: Etcd 密码（可选）

### 3. 客户端使用

#### 基本使用

```go
import (
    "github.com/roc/roc-im-server/internal/msggateway"
)

// 获取消息服务客户端
msgClient, err := msggateway.GetMsgServiceClient()
if err != nil {
    log.Fatal(err)
}

// 使用客户端调用服务
response, err := msgClient.GetMaxSeq(context.Background(), request)

// 获取会话服务客户端
convClient, err := msggateway.GetConversationServiceClient()
if err != nil {
    log.Fatal(err)
}

// 使用会话服务客户端
convResponse, err := convClient.GetConversationMaxSeq(context.Background(), convRequest)
```

#### 使用一致性哈希

```go
// 基于用户ID进行一致性哈希路由
userID := "user123"
msgClient, err := msggateway.GetMsgServiceClientWithKey(userID)
if err != nil {
    log.Fatal(err)
}

// 会话服务也支持一致性哈希
convClient, err := msggateway.GetConversationServiceClientWithKey(userID)
if err != nil {
    log.Fatal(err)
}
```

#### 自定义配置

```go
import (
    "github.com/roc/roc-im-server/tools/serviceDiscovery"
    "github.com/roc/roc-im-server/tools/loadbalancer"
)

// 创建自定义配置
config := servicediscovery.DefaultServiceDiscoveryConfig()
config.LoadBalancer.Type = "consistent_hash"
config.LoadFromEnv()

// 创建注册中心
registry, err := config.CreateServiceRegistry(logger)
if err != nil {
    log.Fatal(err)
}

// 创建负载均衡器
loadBalancer := config.CreateLoadBalancer()

// 创建服务发现客户端
client := servicediscovery.NewServiceDiscoveryClient(servicediscovery.Config{
    Registry:     registry,
    LoadBalancer: loadBalancer,
    Logger:       logger,
})
```

### 4. 多实例部署

#### 手动启动多个实例

启动多个消息服务实例：

```bash
# 消息服务实例 1
export MSG_SERVICE_PORT=10100
go run main.go &

# 消息服务实例 2  
export MSG_SERVICE_PORT=10101
go run main.go &

# 会话服务实例 1
export CONVERSATION_SERVICE_PORT=10200
go run main.go &

# 会话服务实例 2
export CONVERSATION_SERVICE_PORT=10201
go run main.go &
```

#### 使用部署脚本

使用提供的部署脚本可以更方便地管理多个服务实例：

```bash
# 启动所有服务实例（默认各启动2个实例）
./scripts/deploy_services.sh start

# 查看服务状态
./scripts/deploy_services.sh status

# 查看注册的服务
./scripts/deploy_services.sh services

# 查看特定服务的日志
./scripts/deploy_services.sh logs msg 0      # 查看消息服务实例0的日志
./scripts/deploy_services.sh logs conversation 1  # 查看会话服务实例1的日志

# 停止所有服务
./scripts/deploy_services.sh stop
```

客户端会自动发现所有实例并进行负载均衡。

## 配置选项

### 负载均衡策略配置

通过环境变量 `LOAD_BALANCER_TYPE` 设置：

```bash
# 轮询（默认）
export LOAD_BALANCER_TYPE=round_robin

# 随机
export LOAD_BALANCER_TYPE=random

# 一致性哈希
export LOAD_BALANCER_TYPE=consistent_hash

# 加权随机
export LOAD_BALANCER_TYPE=weighted_random

# 最少活跃调用数
export LOAD_BALANCER_TYPE=least_active
```

### Etcd 配置

```bash
export ETCD_ENDPOINTS=localhost:2379,localhost:2380,localhost:2381
export ETCD_USERNAME=your_username
export ETCD_PASSWORD=your_password
export ETCD_ROOT_PATH=/your-app/services
```

## 监控和调试

### 查看注册的服务实例

```bash
# 使用 etcdctl 查看注册的服务
etcdctl get --prefix /roc-im-server/services/msg-service/
etcdctl get --prefix /roc-im-server/services/conversation-service/

# 或使用部署脚本
./scripts/deploy_services.sh services
```

### 服务发现客户端状态

```go
// 获取所有服务实例
client := msggateway.GetServiceDiscoveryClient()
services := client.GetAllServices()
for serviceName, instances := range services {
    fmt.Printf("Service: %s, Instances: %d\n", serviceName, len(instances))
}

// 健康检查
err := client.HealthCheck()
if err != nil {
    log.Printf("Health check failed: %v", err)
}
```

### 刷新服务缓存

```go
// 手动刷新消息服务实例缓存
err := msggateway.RefreshMsgService()
if err != nil {
    log.Printf("Failed to refresh service: %v", err)
}

// 手动刷新会话服务实例缓存
err = msggateway.RefreshConversationService()
if err != nil {
    log.Printf("Failed to refresh conversation service: %v", err)
}
```

## 最佳实践

1. **服务实例权重设置**: 根据服务器性能设置不同的权重值
2. **健康检查**: 定期更新服务实例的健康状态
3. **优雅关闭**: 服务停止时正确注销实例
4. **连接池管理**: 合理设置连接池大小和超时时间
5. **监控告警**: 监控服务注册中心的健康状态

## 扩展性

### 添加新的注册中心

1. 实现 `ServiceRegistry` 接口
2. 在配置中添加新的注册中心类型
3. 更新工厂方法

### 添加新的负载均衡策略

1. 实现 `LoadBalancer` 接口
2. 在工厂中注册新策略
3. 更新配置选项

## 故障排查

### 常见问题

1. **服务注册失败**
   - 检查 Etcd 服务是否正常运行
   - 验证网络连接和防火墙设置
   - 检查认证信息是否正确

2. **服务发现失败**
   - 确认服务名称是否正确
   - 检查服务实例是否已注册
   - 验证根路径配置

3. **负载均衡不均匀**
   - 检查实例权重设置
   - 验证负载均衡策略配置
   - 确认所有实例状态为健康

### 日志调试

启用详细日志：

```go
logger, _ := zap.NewDevelopment()
// 或者
logger, _ := zap.NewProduction()
```

## 性能优化

1. **缓存策略**: 合理设置实例缓存 TTL
2. **连接复用**: 启用连接池和长连接
3. **批量操作**: 使用批量注册和注销接口
4. **本地缓存**: 启用本地服务实例缓存

## 安全考虑

1. **认证**: 启用 Etcd 认证
2. **加密**: 使用 TLS 加密通信
3. **访问控制**: 设置适当的权限控制
4. **网络隔离**: 在安全的网络环境中部署
