# RocketMQ Docker 部署指南

本目录包含了使用 Docker 部署 RocketMQ 的完整解决方案，支持单机版和集群版。

## 📁 目录结构

```
docker/rocketmq/
├── docker-compose.yml       # Docker Compose 配置
├── broker.conf              # Broker 配置文件
├── broker2.conf             # Broker 2 配置文件
├── scripts/
│   └── deploy.sh            # 部署管理脚本
├── quick_start.sh           # 快速启动脚本
└── README.md                # 本文档
```

## 🚀 快速开始

### 方法1: 一键启动（推荐）

```bash
# 进入 RocketMQ 目录
cd docker/rocketmq

# 一键启动单机版 RocketMQ
./quick_start.sh
```

### 方法2: 使用管理脚本

```bash
# 启动单机版 RocketMQ
./scripts/deploy.sh standalone

# 启动集群版 RocketMQ
./scripts/deploy.sh cluster

# 验证部署
./scripts/deploy.sh test
```

## 📋 部署选项

### 单机版（推荐）
- **组件**: Name Server + Broker + Console
- **端口映射**:
  - Name Server: 9876
  - Broker: 10909, 10911, 10912
  - Console: 8083
- **管理界面**: http://localhost:8083

### 集群版
- **组件**: 2个 Name Server + 2个 Broker + Console
- **端口映射**:
  - Name Server 1: 9876
  - Name Server 2: 9877
  - Broker 1: 10909, 10911, 10912
  - Broker 2: 10919, 10921, 10922
  - Console: 8083

## 🛠️ 管理命令

```bash
# 查看帮助
./scripts/deploy.sh help

# 启动服务
./scripts/deploy.sh standalone    # 单机版
./scripts/deploy.sh cluster       # 集群版

# 管理服务
./scripts/deploy.sh status        # 查看状态
./scripts/deploy.sh logs          # 查看日志
./scripts/deploy.sh stop          # 停止服务
./scripts/deploy.sh restart       # 重启服务
./scripts/deploy.sh clean         # 清理所有数据

# 测试功能
./scripts/deploy.sh test          # 测试连接
./scripts/deploy.sh create-topic  # 创建主题
./scripts/deploy.sh produce       # 生产消息
./scripts/deploy.sh consume       # 消费消息
```

## 🔧 配置说明

### Broker 配置

```properties
# broker.conf
brokerClusterName=DefaultCluster
brokerName=broker-a
brokerId=0
brokerRole=ASYNC_MASTER
flushDiskType=ASYNC_FLUSH
namesrvAddr=rmqnamesrv:9876
listenPort=10911
autoCreateTopicEnable=true
autoCreateSubscriptionGroup=true
```

### 环境变量配置

```bash
# JVM 配置
export JAVA_OPT_EXT="-server -Xms1g -Xmx1g -Xmn512m"

# Name Server 地址
export ROCKETMQ_NAMESRV_ADDR="localhost:9876"
```

## 🔍 监控和调试

### 1. 健康检查

```bash
# 检查 Name Server
./scripts/deploy.sh test

# 查看集群状态
docker exec rmqnamesrv sh mqadmin clusterList -n localhost:9876

# 查看 Broker 状态
docker exec rmqbroker sh mqadmin brokerStatus -n localhost:9876 -b rmqbroker:10911
```

### 2. 主题管理

```bash
# 创建主题
./scripts/deploy.sh create-topic my-topic 4

# 查看主题列表
docker exec rmqnamesrv sh mqadmin topicList -n localhost:9876

# 查看主题详情
docker exec rmqnamesrv sh mqadmin topicStatus -n localhost:9876 -t my-topic
```

### 3. 消息测试

```bash
# 生产消息
./scripts/deploy.sh produce my-topic "Hello RocketMQ!"

# 消费消息
./scripts/deploy.sh consume my-topic test-group

# 查看消息
docker exec rmqnamesrv sh mqadmin printMsg -n localhost:9876 -t my-topic
```

## 🔐 安全配置

### 1. ACL 配置

```properties
# 在 broker.conf 中添加
aclEnable=true
```

### 2. 用户认证

```bash
# 创建 ACL 用户
docker exec rmqnamesrv sh mqadmin updateAclConfig -n localhost:9876 -c DefaultCluster -u admin -p admin123 -t admin
```

### 3. 网络安全

```yaml
# 在 docker-compose.yml 中限制网络访问
networks:
  rocketmq-network:
    driver: bridge
    internal: true
```

## 🚨 故障排除

### 1. 容器启动失败

```bash
# 查看详细日志
docker logs rmqnamesrv
docker logs rmqbroker
docker logs rmqconsole

# 检查端口冲突
netstat -tlnp | grep 9876
netstat -tlnp | grep 10911
```

### 2. 集群无法形成

```bash
# 检查网络连接
docker exec rmqnamesrv ping rmqbroker
docker exec rmqbroker ping rmqnamesrv

# 检查配置
docker exec rmqbroker cat /home/rocketmq/rocketmq-5.1.4/conf/broker.conf
```

### 3. 消息丢失

```bash
# 检查存储路径
docker exec rmqbroker ls -la /home/rocketmq/store

# 检查日志
docker exec rmqbroker tail -f /home/rocketmq/logs/rocketmqlogs/broker.log
```

## 📊 性能调优

### 1. JVM 调优

```yaml
environment:
  - JAVA_OPT_EXT=-server -Xms2g -Xmx2g -Xmn1g -XX:+UseG1GC -XX:MaxGCPauseMillis=20
```

### 2. Broker 调优

```properties
# broker.conf
maxMessageSize=65536
flushIntervalCommitLog=500
flushIntervalConsumeQueue=1000
maxTransferCountOnMessageInMemory=32
maxTransferSizeOnMessageInMemory=262144
```

### 3. 存储优化

```yaml
volumes:
  - broker_store:/home/rocketmq/store:rw
  - /dev/shm:/dev/shm:rw  # 共享内存
```

## 🔄 升级和迁移

### 1. 版本升级

```bash
# 停止服务
./scripts/deploy.sh stop

# 修改镜像版本
# 重新启动
./scripts/deploy.sh standalone
```

### 2. 数据迁移

```bash
# 备份数据
docker run --rm -v rocketmq_broker_store:/data -v $(pwd):/backup alpine tar czf /backup/rocketmq-backup.tar.gz -C /data .

# 恢复数据
docker run --rm -v rocketmq_broker_store:/data -v $(pwd):/backup alpine tar xzf /backup/rocketmq-backup.tar.gz -C /data
```

## 📝 注意事项

1. **资源要求**: RocketMQ 需要足够的内存和存储空间
2. **端口管理**: 避免端口冲突，特别是 9876 和 10911
3. **数据备份**: 定期备份 Broker 存储数据
4. **监控告警**: 配置监控和告警，监控 RocketMQ 集群状态
5. **版本兼容**: 注意客户端和服务端版本兼容性

## 🤝 贡献

如有问题或建议，请提交 Issue 或 Pull Request。
