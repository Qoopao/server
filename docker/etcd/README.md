# Docker etcd 部署指南

本目录包含了使用 Docker 部署 etcd 的完整解决方案，支持单机版和集群版部署。

## 📁 目录结构

```
docker/etcd/
├── docker-compose.yml          # 完整配置（包含单机和集群）
├── docker-compose.standalone.yml # 单机版配置
├── docker-compose.cluster.yml   # 集群版配置
├── scripts/
│   └── deploy.sh               # 部署管理脚本
└── README.md                   # 本文档
```

## 🚀 快速开始

### 1. 单机版部署（开发环境）

```bash
# 进入etcd目录
cd docker/etcd

# 启动单机版etcd
./scripts/deploy.sh standalone

# 验证部署
./scripts/deploy.sh test
```

### 2. 集群版部署（生产环境）

```bash
# 启动3节点etcd集群
./scripts/deploy.sh cluster

# 验证集群状态
./scripts/deploy.sh test
```

## 📋 部署选项

### 单机版 (Standalone)
- **用途**: 开发、测试环境
- **节点数**: 1个
- **端口**: 2379 (客户端), 2380 (对等)
- **数据持久化**: 本地卷
- **高可用**: 无

### 集群版 (Cluster)
- **用途**: 生产环境
- **节点数**: 3个
- **端口映射**:
  - etcd-1: 2379, 2380
  - etcd-2: 2381, 2382
  - etcd-3: 2383, 2384
- **数据持久化**: 独立卷
- **高可用**: 支持

## 🛠️ 管理命令

```bash
# 查看帮助
./scripts/deploy.sh help

# 查看状态
./scripts/deploy.sh status

# 查看日志
./scripts/deploy.sh logs

# 停止服务
./scripts/deploy.sh stop

# 重启服务
./scripts/deploy.sh restart

# 清理所有数据
./scripts/deploy.sh clean
```

## 🔧 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| ETCD_NAME | 节点名称 | etcd-1 |
| ETCD_DATA_DIR | 数据目录 | /etcd-data |
| ETCD_LISTEN_CLIENT_URLS | 客户端监听地址 | http://0.0.0.0:2379 |
| ETCD_ADVERTISE_CLIENT_URLS | 客户端广播地址 | http://localhost:2379 |
| ETCD_LISTEN_PEER_URLS | 对等监听地址 | http://0.0.0.0:2380 |
| ETCD_INITIAL_CLUSTER | 初始集群配置 | 根据部署模式 |
| ETCD_INITIAL_CLUSTER_STATE | 集群状态 | new |
| ETCD_INITIAL_CLUSTER_TOKEN | 集群令牌 | etcd-cluster-1 |

### 网络配置

- **网络名称**: etcd-network
- **网络类型**: bridge
- **容器间通信**: 通过容器名称

## 🔍 监控和调试

### 1. 查看注册的服务（推荐）

使用提供的监控脚本：

```bash
# 查看所有服务（默认命名空间 /services）
cd docker/etcd
./scripts/view_services.sh

# 查看指定服务
./scripts/view_services.sh service backbon-service

# 查看长连接服务的命名空间
ETCD_NAMESPACE=/long-connection-service ./scripts/view_services.sh

# 查看所有命名空间的服务
ETCD_NAMESPACE=/ ./scripts/view_services.sh
```

### 2. Web UI 管理界面

启动 etcd-ui 后，访问 http://localhost:8084 可以在浏览器中查看和管理 etcd 数据。

```bash
# 启动 etcd-ui（如果未启动）
cd docker/etcd
docker-compose up -d etcd-ui

# 访问 http://localhost:8084
```

### 3. 健康检查

```bash
# 检查单机版
docker exec etcd-standalone etcdctl endpoint health

# 检查集群版
docker exec etcd-1 etcdctl endpoint health
docker exec etcd-2 etcdctl endpoint health
docker exec etcd-3 etcdctl endpoint health
```

### 4. 集群信息

```bash
# 查看集群成员
docker exec etcd-1 etcdctl member list

# 查看集群状态
docker exec etcd-1 etcdctl endpoint status
```

### 5. 直接使用 etcdctl 查看服务

```bash
# 查看默认命名空间的服务
docker exec etcd-standalone etcdctl get /services --prefix

# 查看长连接服务的命名空间
docker exec etcd-standalone etcdctl get /long-connection-service --prefix

# 查看所有键
docker exec etcd-standalone etcdctl get "" --prefix --keys-only
```

### 6. 数据操作

```bash
# 设置键值
docker exec etcd-standalone etcdctl put /test/key value

# 获取键值
docker exec etcd-standalone etcdctl get /test/key

# 删除键值
docker exec etcd-standalone etcdctl del /test/key
```

## 🔐 安全配置

### 1. TLS 配置（生产环境推荐）

```yaml
environment:
  - ETCD_LISTEN_CLIENT_URLS=https://0.0.0.0:2379
  - ETCD_ADVERTISE_CLIENT_URLS=https://etcd-1:2379
  - ETCD_CERT_FILE=/etc/etcd/cert.pem
  - ETCD_KEY_FILE=/etc/etcd/key.pem
  - ETCD_CA_FILE=/etc/etcd/ca.pem
```

### 2. 认证配置

```yaml
environment:
  - ETCD_AUTH_TOKEN=simple
  - ETCD_ROOT_PASSWORD=your-password
```

## 🚨 故障排除

### 1. 容器启动失败

```bash
# 查看详细日志
docker logs etcd-1

# 检查端口冲突
netstat -tlnp | grep 2379
```

### 2. 集群无法形成

```bash
# 检查网络连接
docker exec etcd-1 ping etcd-2
docker exec etcd-1 ping etcd-3

# 检查集群配置
docker exec etcd-1 etcdctl member list
```

### 3. 数据丢失

```bash
# 备份数据
docker exec etcd-1 etcdctl snapshot save /tmp/backup.db

# 恢复数据
docker exec etcd-1 etcdctl snapshot restore /tmp/backup.db
```

## 📊 性能调优

### 1. 资源限制

```yaml
deploy:
  resources:
    limits:
      memory: 1G
      cpus: '0.5'
    reservations:
      memory: 512M
      cpus: '0.25'
```

### 2. 存储优化

```yaml
volumes:
  - etcd_data:/etcd-data:rw
  - /dev/shm:/dev/shm:rw  # 共享内存
```

## 🔄 升级和迁移

### 1. 版本升级

```bash
# 停止服务
./scripts/deploy.sh stop

# 修改镜像版本
# 重新启动
./scripts/deploy.sh cluster
```

### 2. 数据迁移

```bash
# 导出数据
docker exec etcd-1 etcdctl snapshot save /tmp/migration.db

# 导入数据
docker exec etcd-new etcdctl snapshot restore /tmp/migration.db
```

## 📝 注意事项

1. **生产环境**: 建议使用集群版，确保高可用性
2. **数据备份**: 定期备份 etcd 数据
3. **监控**: 配置监控告警，监控 etcd 集群状态
4. **安全**: 生产环境必须配置 TLS 和认证
5. **资源**: 确保有足够的内存和存储空间

## 🤝 贡献

如有问题或建议，请提交 Issue 或 Pull Request。
