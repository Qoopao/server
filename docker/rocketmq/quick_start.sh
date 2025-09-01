#!/bin/bash

# RocketMQ 快速启动脚本
# 一键启动单机版 RocketMQ

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}🚀 RocketMQ 快速启动脚本${NC}"
echo "================================"

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}❌ Docker 未安装，请先安装 Docker${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo -e "${YELLOW}❌ Docker Compose 未安装，请先安装 Docker Compose${NC}"
    exit 1
fi

# 检查 Docker 服务
if ! docker info &> /dev/null; then
    echo -e "${YELLOW}❌ Docker 服务未运行，正在启动...${NC}"
    sudo systemctl start docker
fi

# 进入 RocketMQ 目录
cd "$(dirname "$0")"

echo -e "${GREEN}📦 启动单机版 RocketMQ...${NC}"

# 检查是否已经运行
if docker-compose ps | grep -c "Up" | grep -q "3"; then
    echo -e "${GREEN}✅ RocketMQ 已经在运行中！${NC}"
    echo ""
    echo "📋 服务信息:"
    echo "  - Name Server: localhost:9876"
    echo "  - Broker: localhost:10911"
    echo "  - Console: http://localhost:8083"
    echo ""
    echo "🔧 常用命令:"
    echo "  ./scripts/deploy.sh status     # 查看状态"
    echo "  ./scripts/deploy.sh test       # 测试连接"
    echo "  ./scripts/deploy.sh logs       # 查看日志"
    echo "  ./scripts/deploy.sh stop       # 停止服务"
    exit 0
fi

# 启动 RocketMQ
docker-compose up -d rmqnamesrv rmqbroker rmqconsole

echo -e "${GREEN}⏳ 等待 RocketMQ 启动...${NC}"
sleep 15

# 检查启动状态
if docker-compose ps | grep -c "Up" | grep -q "3"; then
    echo -e "${GREEN}✅ RocketMQ 启动成功！${NC}"
    echo ""
    echo "📋 服务信息:"
    echo "  - Name Server: localhost:9876"
    echo "  - Broker: localhost:10911"
    echo "  - Console: http://localhost:8083"
    echo ""
    echo "🔧 常用命令:"
    echo "  ./scripts/deploy.sh status     # 查看状态"
    echo "  ./scripts/deploy.sh test       # 测试连接"
    echo "  ./scripts/deploy.sh logs       # 查看日志"
    echo "  ./scripts/deploy.sh stop       # 停止服务"
    echo ""
    echo "📝 测试命令:"
    echo "  ./scripts/deploy.sh create-topic test-topic"
    echo "  ./scripts/deploy.sh produce test-topic 'Hello RocketMQ!'"
    echo "  ./scripts/deploy.sh consume test-topic"
else
    echo -e "${YELLOW}❌ RocketMQ 启动失败，查看日志:${NC}"
    docker-compose logs
    exit 1
fi
