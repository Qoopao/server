# Kafka Docker 部署指南

本目录包含了使用 Docker 部署 Kafka 的完整解决方案，支持 KRaft 模式和 ZooKeeper 模式。

## 📁 目录结构

```
docker/kafka/
├── docker-compose.kraft.yml     # KRaft 模式配置（推荐）
├── scripts/
│   └── deploy.sh                # 部署管理脚本
├── quick_start.sh               # 快速启动脚本
└── README.md                    # 本文档
```

## 🚀 快速开始

### 方法1: 一键启动（推荐）

```bash
# 进入 Kafka 目录
cd docker/kafka

# 一键启动 KRaft 模式 Kafka
./quick_start.sh
```

### 方法2: 使用管理脚本

```bash
# 启动 KRaft 模式 Kafka
./scripts/deploy.sh kraft

# 验证部署
./scripts/deploy.sh test
```

## 📋 部署选项

### KRaft 模式（推荐）
- **版本要求**: Kafka >= 2.8.0
- **依赖**: 无外部依赖
- **节点数**: 3个
- **端口映射**:
  - kafka-1: 9092, 9093, 9101
  - kafka-2: 9094, 9095, 9102
  - kafka-3: 9096, 9097, 9103
- **管理界面**: http://localhost:8080

### ZooKeeper 模式（传统）
- **版本要求**: 所有版本
- **依赖**: 需要 ZooKeeper
- **复杂度**: 高（需要管理两个系统）

## 🛠️ 管理命令

```bash
# 查看帮助
./scripts/deploy.sh help

# 启动服务
./scripts/deploy.sh kraft          # KRaft 模式
./scripts/deploy.sh zookeeper      # ZooKeeper 模式

# 管理服务
./scripts/deploy.sh status         # 查看状态
./scripts/deploy.sh logs           # 查看日志
./scripts/deploy.sh stop           # 停止服务
./scripts/deploy.sh restart        # 重启服务
./scripts/deploy.sh clean          # 清理所有数据

# 测试功能
./scripts/deploy.sh test           # 测试连接
./scripts/deploy.sh create-topic   # 创建主题
./scripts/deploy.sh produce        # 生产消息
./scripts/deploy.sh consume        # 消费消息
```

## 🔧 配置说明

### KRaft 模式配置

```yaml
# docker-compose.kraft.yml
services:
  kafka-1:
    environment:
      KAFKA_NODE_ID: 1
      KAFKA_PROCESS_ROLES: "broker,controller"
      KAFKA_CONTROLLER_QUORUM_VOTERS: "1@kafka-1:9093,2@kafka-2:9093,3@kafka-3:9093"
      KAFKA_LISTENERS: "PLAINTEXT://0.0.0.0:29092,CONTROLLER://0.0.0.0:9093,EXTERNAL://0.0.0.0:9092"
      KAFKA_ADVERTISED_LISTENERS: "PLAINTEXT://kafka-1:29092,EXTERNAL://localhost:9092"
```

### 环境变量配置

```bash
# 自定义配置
export KAFKA_BROKER_ID=1
export KAFKA_NODE_ID=1
export KAFKA_PROCESS_ROLES="broker,controller"
export KAFKA_CONTROLLER_QUORUM_VOTERS="1@kafka-1:9093,2@kafka-2:9093,3@kafka-3:9093"
```

## 🔍 监控和调试

### 1. 健康检查

```bash
# 检查 Kafka 集群状态
./scripts/deploy.sh test

# 查看主题列表
docker exec kafka-1 kafka-topics --bootstrap-server kafka-1:29092 --list

# 查看集群信息
docker exec kafka-1 kafka-broker-api-versions --bootstrap-server kafka-1:29092
```

### 2. 主题管理

```bash
# 创建主题
./scripts/deploy.sh create-topic my-topic 3 3

# 查看主题详情
docker exec kafka-1 kafka-topics --bootstrap-server kafka-1:29092 --describe --topic my-topic

# 删除主题
docker exec kafka-1 kafka-topics --bootstrap-server kafka-1:29092 --delete --topic my-topic
```

### 3. 消息测试

```bash
# 生产消息
./scripts/deploy.sh produce my-topic "Hello Kafka!"

# 消费消息
./scripts/deploy.sh consume my-topic 10

# 实时消费
docker exec kafka-1 kafka-console-consumer --bootstrap-server kafka-1:29092 --topic my-topic --from-beginning
```

## 🔐 安全配置

### 1. TLS 配置（生产环境推荐）

```yaml
environment:
  KAFKA_LISTENERS: "SSL://0.0.0.0:29092,CONTROLLER://0.0.0.0:9093,EXTERNAL://0.0.0.0:9092"
  KAFKA_SSL_KEYSTORE_LOCATION: /etc/kafka/ssl/kafka.server.keystore.jks
  KAFKA_SSL_KEYSTORE_PASSWORD: password
  KAFKA_SSL_KEY_PASSWORD: password
  KAFKA_SSL_TRUSTSTORE_LOCATION: /etc/kafka/ssl/kafka.server.truststore.jks
  KAFKA_SSL_TRUSTSTORE_PASSWORD: password
```

### 2. SASL 认证

```yaml
environment:
  KAFKA_SASL_ENABLED_MECHANISMS: PLAIN
  KAFKA_SASL_MECHANISM_INTER_BROKER_PROTOCOL: PLAIN
  KAFKA_AUTHORIZER_CLASS_NAME: kafka.security.authorizer.AclAuthorizer
```

## 🚨 故障排除

### 1. 容器启动失败

```bash
# 查看详细日志
docker logs kafka-1
docker logs kafka-2
docker logs kafka-3

# 检查端口冲突
netstat -tlnp | grep 9092
netstat -tlnp | grep 9094
netstat -tlnp | grep 9096
```

### 2. 集群无法形成

```bash
# 检查网络连接
docker exec kafka-1 ping kafka-2
docker exec kafka-1 ping kafka-3

# 检查 KRaft 配置
docker exec kafka-1 kafka-metadata-shell.sh --snapshot /var/lib/kafka/data/__cluster_metadata-0/00000000000000000000.log
```

### 3. 数据丢失

```bash
# 备份数据
docker exec kafka-1 tar -czf /tmp/kafka-backup.tar.gz /var/lib/kafka/data

# 恢复数据
docker exec kafka-1 tar -xzf /tmp/kafka-backup.tar.gz -C /
```

## 📊 性能调优

### 1. 资源限制

```yaml
deploy:
  resources:
    limits:
      memory: 2G
      cpus: '1.0'
    reservations:
      memory: 1G
      cpus: '0.5'
```

### 2. JVM 调优

```yaml
environment:
  KAFKA_HEAP_OPTS: "-Xmx1G -Xms1G"
  KAFKA_JVM_PERFORMANCE_OPTS: "-server -XX:+UseG1GC -XX:MaxGCPauseMillis=20 -XX:InitiatingHeapOccupancyPercent=35 -XX:+ExplicitGCInvokesConcurrent -XX:MaxInlineLevel=15"
```

### 3. 存储优化

```yaml
volumes:
  - kafka_data:/var/lib/kafka/data:rw
  - /dev/shm:/dev/shm:rw  # 共享内存
```

## 🔄 升级和迁移

### 1. 版本升级

```bash
# 停止服务
./scripts/deploy.sh stop

# 修改镜像版本
# 重新启动
./scripts/deploy.sh kraft
```

### 2. 从 ZooKeeper 迁移到 KRaft

```bash
# 1. 备份现有数据
# 2. 停止 ZooKeeper 模式
./scripts/deploy.sh stop

# 3. 启动 KRaft 模式
./scripts/deploy.sh kraft

# 4. 验证迁移
./scripts/deploy.sh test
```

## 📝 注意事项

1. **生产环境**: 建议使用 KRaft 模式，确保高可用性
2. **数据备份**: 定期备份 Kafka 数据
3. **监控**: 配置监控告警，监控 Kafka 集群状态
4. **安全**: 生产环境必须配置 TLS 和认证
5. **资源**: 确保有足够的内存和存储空间

## 🤝 贡献

如有问题或建议，请提交 Issue 或 Pull Request。
