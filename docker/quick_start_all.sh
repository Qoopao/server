#!/bin/bash

# 统一快速启动脚本
# 支持 etcd, kafka, redis, mongodb, rocketmq

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${GREEN}🚀 统一快速启动脚本${NC}"
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

# 显示菜单
show_menu() {
    echo ""
    echo -e "${BLUE}请选择要启动的服务:${NC}"
    echo "1) etcd (服务发现)"
    echo "2) kafka (消息队列)"
    echo "3) redis (缓存)"
    echo "4) mongodb (数据库)"
    echo "5) rocketmq (消息队列)"
    echo "6) 全部启动"
    echo "0) 退出"
    echo ""
    read -p "请输入选择 (0-6): " choice
}

# 启动 etcd
start_etcd() {
    echo -e "${GREEN}📦 启动 etcd...${NC}"
    cd docker/etcd
    ./scripts/deploy.sh standalone
    cd ../..
}

# 启动 kafka
start_kafka() {
    echo -e "${GREEN}📦 启动 kafka...${NC}"
    cd docker/kafka
    ./quick_start.sh
    cd ../..
}

# 启动 redis
start_redis() {
    echo -e "${GREEN}📦 启动 redis...${NC}"
    cd docker/redis
    ./scripts/deploy.sh standalone
    cd ../..
}

# 启动 mongodb
start_mongodb() {
    echo -e "${GREEN}📦 启动 mongodb...${NC}"
    cd docker/mongodb
    docker-compose up -d mongodb-standalone mongo-express
    echo "等待 MongoDB 启动..."
    sleep 10
    cd ../..
}

# 启动 rocketmq
start_rocketmq() {
    echo -e "${GREEN}📦 启动 rocketmq...${NC}"
    cd docker/rocketmq
    docker-compose up -d rmqnamesrv rmqbroker rmqconsole
    echo "等待 RocketMQ 启动..."
    sleep 15
    cd ../..
}

# 启动全部服务
start_all() {
    echo -e "${GREEN}📦 启动所有服务...${NC}"
    start_etcd
    start_kafka
    start_redis
    start_mongodb
    start_rocketmq
}

# 显示服务信息
show_info() {
    echo ""
    echo -e "${GREEN}✅ 服务启动完成！${NC}"
    echo ""
    echo "📋 服务信息:"
    echo "  - etcd: http://localhost:2379"
    echo "  - kafka: localhost:9092,9094,9096"
    echo "  - kafka-ui: http://localhost:8080"
    echo "  - redis: localhost:6379"
    echo "  - redis-ui: http://localhost:8081"
    echo "  - mongodb: localhost:27017"
    echo "  - mongo-express: http://localhost:8082"
    echo "  - rocketmq: localhost:9876"
    echo "  - rocketmq-console: http://localhost:8083"
    echo ""
    echo "🔧 管理命令:"
    echo "  docker/etcd/scripts/deploy.sh status"
    echo "  docker/kafka/scripts/deploy.sh status"
    echo "  docker/redis/scripts/deploy.sh status"
    echo "  docker-compose -f docker/mongodb/docker-compose.yml ps"
    echo "  docker-compose -f docker/rocketmq/docker-compose.yml ps"
}

# 主循环
while true; do
    show_menu
    
    case $choice in
        1)
            start_etcd
            show_info
            ;;
        2)
            start_kafka
            show_info
            ;;
        3)
            start_redis
            show_info
            ;;
        4)
            start_mongodb
            show_info
            ;;
        5)
            start_rocketmq
            show_info
            ;;
        6)
            start_all
            show_info
            ;;
        0)
            echo "退出..."
            exit 0
            ;;
        *)
            echo "无效选择，请重新输入"
            ;;
    esac
    
    echo ""
    read -p "按回车键继续..."
done
