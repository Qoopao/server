#!/bin/bash

# MongoDB 快速启动脚本
# 一键启动单机版 MongoDB

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}🚀 MongoDB 快速启动脚本${NC}"
echo "================================"

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}❌ Docker 未安装，请先安装 Docker${NC}"
    exit 1
fi

# 检查 Docker Compose（支持新旧两种格式）
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${YELLOW}❌ Docker Compose 未安装，请先安装 Docker Compose${NC}"
    exit 1
fi

# 检查 Docker 服务
if ! docker info &> /dev/null; then
    echo -e "${YELLOW}❌ Docker 服务未运行，正在启动...${NC}"
    sudo systemctl start docker
fi

# 进入 MongoDB 目录
cd "$(dirname "$0")"

echo -e "${GREEN}📦 启动单机版 MongoDB...${NC}"

# 检查是否已经运行
if command -v docker-compose &> /dev/null; then
    if docker-compose ps | grep -q "mongodb-standalone.*Up"; then
        echo -e "${GREEN}✅ MongoDB 已经在运行中！${NC}"
        echo ""
        echo "📋 服务信息:"
        echo "  - MongoDB: localhost:27017"
        echo "  - MongoDB UI: http://localhost:8082"
        echo "  - 用户名: admin"
        echo "  - 密码: admin123"
        echo ""
        echo "🔧 常用命令:"
        echo "  ./scripts/deploy.sh status     # 查看状态"
        echo "  ./scripts/deploy.sh test       # 测试连接"
        echo "  ./scripts/deploy.sh logs       # 查看日志"
        echo "  ./scripts/deploy.sh stop       # 停止服务"
        exit 0
    fi
else
    if docker compose ps | grep -q "mongodb-standalone.*Up"; then
        echo -e "${GREEN}✅ MongoDB 已经在运行中！${NC}"
        echo ""
        echo "📋 服务信息:"
        echo "  - MongoDB: localhost:27017"
        echo "  - MongoDB UI: http://localhost:8082"
        echo "  - 用户名: admin"
        echo "  - 密码: admin123"
        echo ""
        echo "🔧 常用命令:"
        echo "  ./scripts/deploy.sh status     # 查看状态"
        echo "  ./scripts/deploy.sh test       # 测试连接"
        echo "  ./scripts/deploy.sh logs       # 查看日志"
        echo "  ./scripts/deploy.sh stop       # 停止服务"
        exit 0
    fi
fi

# 启动 MongoDB
if command -v docker-compose &> /dev/null; then
    docker-compose up -d mongodb-standalone mongo-express
else
    docker compose up -d mongodb-standalone mongo-express
fi

echo -e "${GREEN}⏳ 等待 MongoDB 启动...${NC}"
sleep 10

# 检查启动状态
if command -v docker-compose &> /dev/null; then
    if docker-compose ps | grep -q "mongodb-standalone.*Up"; then
        echo -e "${GREEN}✅ MongoDB 启动成功！${NC}"
        echo ""
        echo "📋 服务信息:"
        echo "  - MongoDB: localhost:27017"
        echo "  - MongoDB UI: http://localhost:8082"
        echo "  - 用户名: admin"
        echo "  - 密码: admin123"
        echo ""
        echo "🔧 常用命令:"
        echo "  ./scripts/deploy.sh status     # 查看状态"
        echo "  ./scripts/deploy.sh test       # 测试连接"
        echo "  ./scripts/deploy.sh logs       # 查看日志"
        echo "  ./scripts/deploy.sh stop       # 停止服务"
        echo ""
        echo "📝 测试命令:"
        echo "  ./scripts/deploy.sh init       # 初始化数据库"
        echo "  docker exec -it mongodb-standalone mongosh"
    else
        echo -e "${YELLOW}❌ MongoDB 启动失败，查看日志:${NC}"
        docker-compose logs
        exit 1
    fi
else
    if docker compose ps | grep -q "mongodb-standalone.*Up"; then
        echo -e "${GREEN}✅ MongoDB 启动成功！${NC}"
        echo ""
        echo "📋 服务信息:"
        echo "  - MongoDB: localhost:27017"
        echo "  - MongoDB UI: http://localhost:8082"
        echo "  - 用户名: admin"
        echo "  - 密码: admin123"
        echo ""
        echo "🔧 常用命令:"
        echo "  ./scripts/deploy.sh status     # 查看状态"
        echo "  ./scripts/deploy.sh test       # 测试连接"
        echo "  ./scripts/deploy.sh logs       # 查看日志"
        echo "  ./scripts/deploy.sh stop       # 停止服务"
        echo ""
        echo "📝 测试命令:"
        echo "  ./scripts/deploy.sh init       # 初始化数据库"
        echo "  docker exec -it mongodb-standalone mongosh"
    else
        echo -e "${YELLOW}❌ MongoDB 启动失败，查看日志:${NC}"
        docker compose logs
        exit 1
    fi
fi
