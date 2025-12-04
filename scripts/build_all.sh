#!/bin/bash

# 构建所有服务的脚本

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "开始构建所有服务..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 创建 bin 目录
mkdir -p bin

# 构建 msggateway
echo "📦 构建 msggateway..."
go build -o bin/msggateway cmd/msggateway/main.go
echo "✅ msggateway 构建完成"

# 构建 msg-service
echo "📦 构建 msg-service..."
go build -o bin/msg-service cmd/msgservice/main.go
echo "✅ msg-service 构建完成"

# 构建 conversation-service
echo "📦 构建 conversation-service..."
go build -o bin/conversation-service cmd/convservice/main.go
echo "✅ conversation-service 构建完成"

# 构建 msgconsumer
echo "📦 构建 msgconsumer..."
go build -o bin/msgconsumer cmd/msgconsumer/main.go
echo "✅ msgconsumer 构建完成"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 所有服务构建完成！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "构建产物："
ls -lh bin/
echo ""
echo "启动方式："
echo "  ./bin/msggateway          # 启动 WebSocket 网关"
echo "  ./bin/msg-service         # 启动消息服务"
echo "  ./bin/conversation-service # 启动会话服务"
echo "  ./bin/msgconsumer         # 启动消息消费者"
echo ""

