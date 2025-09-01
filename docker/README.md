# Docker 部署总览

本目录包含了完整的 Docker 部署解决方案，支持 etcd、Kafka、Redis、MongoDB 和 RocketMQ 的快速部署。

## 📁 目录结构

```
docker/
├── etcd/                    # 服务发现
│   ├── docker-compose.yml
│   ├── scripts/deploy.sh
│   └── README.md
├── kafka/                   # 消息队列
│   ├── docker-compose.kraft.yml
│   ├── scripts/deploy.sh
│   ├── quick_start.sh
│   └── README.md
├── redis/                   # 缓存
│   ├── docker-compose.yml
│   ├── redis.conf
│   ├── sentinel.conf
│   ├── scripts/deploy.sh
│   └── README.md
├── mongodb/                 # 数据库
│   ├── docker-compose.yml
│   ├── mongo-init.js
│   ├── replica-init.js
│   └── README.md
├── rocketmq/               # 消息队列
│   ├── docker-compose.yml
│   └── README.md
├── quick_start_all.sh      # 统一启动脚本
└── README.md              # 本文档
```

## 🚀 快速开始

### 方法1: 统一启动脚本（推荐）

```bash
# 进入 docker 目录
cd docker

# 运行统一启动脚本
./quick_start_all.sh
```

### 方法2: 单独启动服务

```bash
# 启动 etcd
cd docker/etcd
./scripts/deploy.sh standalone

# 启动 kafka
cd docker/kafka
./quick_start.sh

# 启动 redis
cd docker/redis
./scripts/deploy.sh standalone

# 启动 mongodb
cd docker/mongodb
docker-compose up -d mongodb-standalone mongo-express

# 启动 rocketmq
cd docker/rocketmq
docker-compose up -d rmqnamesrv rmqbroker rmqconsole
```

## 📋 服务信息

| 服务 | 端口 | 管理界面 | 说明 |
|------|------|----------|------|
| **etcd** | 2379 | - | 服务发现 |
| **kafka** | 9092,9094,9096 | http://localhost:8080 | 消息队列 |
| **redis** | 6379 | http://localhost:8081 | 缓存 |
| **mongodb** | 27017 | http://localhost:8082 | 数据库 |
| **rocketmq** | 9876 | http://localhost:8083 | 消息队列 |

## 🔧 管理命令

### etcd
```bash
cd docker/etcd
./scripts/deploy.sh status    # 查看状态
./scripts/deploy.sh test     # 测试连接
./scripts/deploy.sh stop     # 停止服务
```

### kafka
```bash
cd docker/kafka
./scripts/deploy.sh status   # 查看状态
./scripts/deploy.sh test     # 测试连接
./scripts/deploy.sh stop     # 停止服务
```

### redis
```bash
cd docker/redis
./scripts/deploy.sh status   # 查看状态
./scripts/deploy.sh test     # 测试连接
./scripts/deploy.sh stop     # 停止服务
```

### mongodb
```bash
cd docker/mongodb
docker-compose ps            # 查看状态
docker-compose logs          # 查看日志
docker-compose down          # 停止服务
```

### rocketmq
```bash
cd docker/rocketmq
docker-compose ps            # 查看状态
docker-compose logs          # 查看日志
docker-compose down          # 停止服务
```

## 🛠️ 配置说明

### 环境变量配置

```bash
# etcd 配置
export ETCD_ENDPOINTS="localhost:2379"
export ETCD_CLUSTER_MODE="false"

# kafka 配置
export KAFKA_BROKER_ID=1
export KAFKA_NODE_ID=1

# redis 配置
export REDIS_PASSWORD="redis123"

# mongodb 配置
export MONGO_INITDB_ROOT_USERNAME="admin"
export MONGO_INITDB_ROOT_PASSWORD="mongodb123"

# rocketmq 配置
export ROCKETMQ_NAMESRV_ADDR="localhost:9876"
```

### 网络配置

所有服务都使用独立的 Docker 网络：
- `etcd-network`
- `kafka-network`
- `redis-network`
- `mongodb-network`
- `rocketmq-network`

## 🔍 监控和调试

### 1. 服务健康检查

```bash
# 检查所有服务状态
docker ps | grep -E "(etcd|kafka|redis|mongo|rocketmq)"

# 检查服务日志
docker logs <container_name> --tail 20
```

### 2. 端口检查

```bash
# 检查端口占用
netstat -tlnp | grep -E "(2379|9092|6379|27017|9876)"
```

### 3. 连接测试

```bash
# 测试 etcd
docker exec etcd-standalone etcdctl endpoint health

# 测试 kafka
docker exec kafka-1 kafka-topics --bootstrap-server kafka-1:29092 --list

# 测试 redis
docker exec redis-standalone redis-cli -a redis123 ping

# 测试 mongodb
docker exec mongodb-standalone mongosh --eval "db.adminCommand('ping')"

# 测试 rocketmq
docker exec rmqnamesrv sh mqadmin clusterList -n localhost:9876
```

## 🚨 故障排除

### 1. 容器启动失败

```bash
# 查看详细日志
docker logs <container_name>

# 检查资源使用
docker stats

# 检查网络连接
docker network ls
docker network inspect <network_name>
```

### 2. 端口冲突

```bash
# 检查端口占用
lsof -i :<port>

# 修改端口映射
# 编辑对应的 docker-compose.yml 文件
```

### 3. 数据持久化

```bash
# 查看数据卷
docker volume ls

# 备份数据
docker run --rm -v <volume_name>:/data -v $(pwd):/backup alpine tar czf /backup/backup.tar.gz -C /data .

# 恢复数据
docker run --rm -v <volume_name>:/data -v $(pwd):/backup alpine tar xzf /backup/backup.tar.gz -C /data
```

## 📊 性能调优

### 1. 资源限制

```yaml
# 在 docker-compose.yml 中添加
deploy:
  resources:
    limits:
      memory: 1G
      cpus: '0.5'
    reservations:
      memory: 512M
      cpus: '0.25'
```

### 2. JVM 调优（Java 服务）

```yaml
environment:
  - JAVA_OPT_EXT=-server -Xms512m -Xmx1g -Xmn256m
```

### 3. 存储优化

```yaml
volumes:
  - data:/data:rw
  - /dev/shm:/dev/shm:rw  # 共享内存
```

## 🔐 安全配置

### 1. 密码配置

```bash
# 修改默认密码
export ETCD_PASSWORD="your_etcd_password"
export REDIS_PASSWORD="your_redis_password"
export MONGO_PASSWORD="your_mongo_password"
```

### 2. 网络安全

```yaml
# 限制网络访问
networks:
  internal:
    driver: bridge
    internal: true
```

### 3. 数据加密

```yaml
# 启用 TLS
environment:
  - KAFKA_SSL_ENABLED=true
  - REDIS_TLS_ENABLED=true
```

## 📝 注意事项

1. **资源要求**: 确保有足够的内存和存储空间
2. **端口管理**: 避免端口冲突
3. **数据备份**: 定期备份重要数据
4. **监控告警**: 配置监控和告警
5. **版本兼容**: 注意服务版本兼容性

## 🤝 贡献

如有问题或建议，请提交 Issue 或 Pull Request。
