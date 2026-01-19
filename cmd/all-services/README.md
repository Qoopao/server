# All Services Launcher

这是一个统一的启动器，用于同时启动所有 IM 服务器的服务组件。

## 支持的服务

- **Message Service** - 消息服务
- **Conversation Service** - 会话服务
- **Sequence Service** - 序列号服务
- **Back Service** - 后端服务
- **Conversation Message Consumer** - 会话消息消费者
- **Long Connection Service** - 长连接服务

## 使用方法

### 编译

```bash
cd roc-im-server/cmd/all-services
go build -o all-services .
```

### 运行

```bash
# 直接运行
./all-services

# 或者使用 go run
go run main.go
```

## 功能特性

1. **并发启动**: 所有服务并发启动，提高启动速度
2. **启动监控**: 实时显示启动进度和状态
3. **错误处理**: 任一服务启动失败将停止整个启动过程
4. **优雅关闭**: 支持 SIGINT 和 SIGTERM 信号的优雅关闭
5. **超时控制**: 启动超时时间为 60 秒
6. **详细日志**: 包含启动时间和服务状态信息

## 环境变量配置

各个服务支持以下环境变量配置：

### Message Service
- `KAFKA_BROKERS` - Kafka 代理地址 (默认: localhost:9092)
- `MESSAGE_MONGODB_URI` - MongoDB 连接 URI (默认: mongodb://localhost:27017)
- `MESSAGE_MONGODB_DATABASE` - MongoDB 数据库名 (默认: im_message)
- `MESSAGE_SERVICE_PORT` - 服务端口 (默认: 10400)
- `ETCD_ENDPOINTS` - Etcd 端点 (默认: localhost:2379)

### Conversation Service
- `CONVERSATION_MONGODB_URI` - MongoDB 连接 URI
- `CONVERSATION_MONGODB_DATABASE` - MongoDB 数据库名
- `CONVERSATION_SERVICE_PORT` - 服务端口
- `ETCD_ENDPOINTS` - Etcd 端点

### Sequence Service
- `SEQUENCE_MONGODB_URI` - MongoDB 连接 URI
- `SEQUENCE_MONGODB_DATABASE` - MongoDB 数据库名
- `SEQUENCE_SERVICE_PORT` - 服务端口
- `ETCD_ENDPOINTS` - Etcd 端点

### Back Service
- `BACK_SERVICE_PORT` - 服务端口
- `ETCD_ENDPOINTS` - Etcd 端点

## 停止服务

使用 `Ctrl+C` 或发送 `SIGTERM` 信号来优雅停止所有服务。

## 日志输出示例

```
2024/01/19 10:30:00 启动所有 IM 服务...
2024/01/19 10:30:00 正在启动 Message Service...
2024/01/19 10:30:00 正在启动 Conversation Service...
2024/01/19 10:30:00 正在启动 Sequence Service...
2024/01/19 10:30:00 正在启动 Back Service...
2024/01/19 10:30:00 正在启动 Conversation Message Consumer...
2024/01/19 10:30:00 正在启动 Message Consumer...
2024/01/19 10:30:01 Message Service 启动成功 (耗时: 1.2s)
2024/01/19 10:30:01 服务启动进度: 1/6 (Message Service)
2024/01/19 10:30:02 Conversation Service 启动成功 (耗时: 2.1s)
2024/01/19 10:30:02 服务启动进度: 2/6 (Conversation Service)
2024/01/19 10:30:03 Sequence Service 启动成功 (耗时: 1.8s)
2024/01/19 10:30:03 服务启动进度: 3/6 (Sequence Service)
2024/01/19 10:30:04 Back Service 启动成功 (耗时: 2.3s)
2024/01/19 10:30:04 服务启动进度: 4/6 (Back Service)
2024/01/19 10:30:05 Conversation Message Consumer 启动成功 (耗时: 1.5s)
2024/01/19 10:30:05 服务启动进度: 5/6 (Conversation Message Consumer)
2024/01/19 10:30:06 Message Consumer 启动成功 (耗时: 1.9s)
2024/01/19 10:30:06 服务启动进度: 6/6 (Message Consumer)
2024/01/19 10:30:06 所有服务启动完成！
```

## 注意事项

1. 确保所有依赖服务（如 MongoDB、Kafka、Etcd）已启动并可访问
2. 如果需要修改服务配置，请设置相应的环境变量
3. 启动过程中如果任一服务失败，整个启动过程将终止
4. 建议在生产环境中使用进程管理工具（如 systemd、supervisor）来管理此启动器
