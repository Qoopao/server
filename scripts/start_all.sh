#!/bin/bash

# 启动所有服务的脚本

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "启动所有服务..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 创建日志目录
mkdir -p logs

# 检查二进制文件是否存在
if [ ! -f "bin/msggateway" ] || [ ! -f "bin/msg-service" ] || [ ! -f "bin/conversation-service" ] || [ ! -f "bin/backservice" ] || [ ! -f "bin/longconnection" ] || [ ! -f "bin/msgconsumer" ]; then
    echo "❌ 二进制文件不存在，请先运行 ./scripts/build_all.sh"
    exit 1
fi

# 启动 msgconsumer
echo "🚀 启动 msgconsumer..."
nohup ./bin/msgconsumer > /dev/null 2>&1 &
echo $! > logs/msgconsumer.pid
echo "✅ msgconsumer 已启动 (PID: $(cat logs/msgconsumer.pid))"
sleep 1

# 启动 msg-service
echo "🚀 启动 msg-service (端口: 10100)..."
export MSG_SERVICE_PORT=10100
nohup ./bin/msg-service > /dev/null 2>&1 &
echo $! > logs/msg-service.pid
echo "✅ msg-service 已启动 (PID: $(cat logs/msg-service.pid))"
sleep 2

# 启动 conversation-service
echo "🚀 启动 conversation-service (端口: 10200)..."
export CONV_SERVICE_PORT=10200
nohup ./bin/conversation-service > /dev/null 2>&1 &
echo $! > logs/conversation-service.pid
echo "✅ conversation-service 已启动 (PID: $(cat logs/conversation-service.pid))"
sleep 2

# 启动 backservice
echo "🚀 启动 backservice (端口: 10300)..."
export BACKSERVICE_PORT=10300
nohup ./bin/backservice > /dev/null 2>&1 &
echo $! > logs/backservice.pid
echo "✅ backservice 已启动 (PID: $(cat logs/backservice.pid))"
sleep 2

# 启动 longconnection
echo "🚀 启动 longconnection (WebSocket: 6060, RPC: 8956)..."
export LONG_CONNECTION_CONFIG=../../roc-foundation-service/long_connection_service/config.yaml
nohup ./bin/longconnection > /dev/null 2>&1 &
echo $! > logs/longconnection.pid
echo "✅ longconnection 已启动 (PID: $(cat logs/longconnection.pid))"
sleep 2

# 启动 msggateway
echo "🚀 启动 msggateway (端口: 10010)..."
nohup ./bin/msggateway > /dev/null 2>&1 &
echo $! > logs/msggateway.pid
echo "✅ msggateway 已启动 (PID: $(cat logs/msggateway.pid))"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 所有服务启动完成！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "服务列表："
echo "  msggateway          : localhost:10010 (WebSocket)"
echo "  msg-service         : localhost:10100 (gRPC)"
echo "  conversation-service: localhost:10200 (gRPC)"
echo "  backservice         : localhost:10300 (gRPC)"
echo "  longconnection      : localhost:6060 (WebSocket), localhost:8956 (RPC)"
echo "  msgconsumer         : 后台消费者"
echo ""
echo "查看日志："
echo "  tail -f logs/msggateway.log"
echo "  tail -f logs/msg-service.log"
echo "  tail -f logs/conversation-service.log"
echo "  tail -f logs/backservice.log"
echo "  tail -f logs/longconnection.log"
echo "  tail -f logs/msgconsumer.log"
echo ""
echo "停止所有服务："
echo "  ./scripts/stop_all.sh"
echo ""

