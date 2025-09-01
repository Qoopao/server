#!/bin/bash

# Kafka 快速启动脚本
# 一键启动 KRaft 模式的 Kafka 集群

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}🚀 Kafka 快速启动脚本${NC}"
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

# 进入 Kafka 目录
cd "$(dirname "$0")"

echo -e "${GREEN}📦 启动 KRaft 模式 Kafka 集群...${NC}"

# 启动 Kafka
docker-compose -f docker-compose.kraft.yml up -d

echo -e "${GREEN}⏳ 等待 Kafka 启动...${NC}"
sleep 15

# 检查启动状态
if docker-compose -f docker-compose.kraft.yml ps | grep -c "Up" | grep -q "3"; then
    echo -e "${GREEN}✅ Kafka 启动成功！${NC}"
    echo ""
    echo "📋 服务信息:"
    echo "  - Kafka-1: localhost:9092"
    echo "  - Kafka-2: localhost:9094" 
    echo "  - Kafka-3: localhost:9096"
    echo "  - Kafka UI: http://localhost:8080"
    echo ""
    echo "🔧 常用命令:"
    echo "  ./scripts/deploy.sh status     # 查看状态"
    echo "  ./scripts/deploy.sh test       # 测试连接"
    echo "  ./scripts/deploy.sh logs       # 查看日志"
    echo "  ./scripts/deploy.sh stop       # 停止服务"
    echo ""
    echo "📝 测试命令:"
    echo "  ./scripts/deploy.sh create-topic test-topic"
    echo "  ./scripts/deploy.sh produce test-topic 'Hello Kafka!'"
    echo "  ./scripts/deploy.sh consume test-topic"
else
    echo -e "${YELLOW}❌ Kafka 启动失败，查看日志:${NC}"
    docker-compose -f docker-compose.kraft.yml logs
    exit 1
fi
