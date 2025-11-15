#!/bin/bash

# 统一快速启动脚本
# 支持 etcd, kafka, redis, mongodb, rocketmq, otel (OpenTelemetry监控栈)

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 脚本根目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

mode_display() {
    if [ "$1" = "cluster" ]; then
        echo "集群模式"
    else
        echo "单节点模式"
    fi
}

show_service_info() {
    local service="$1"
    local mode="${2:-single}"
    local mode_text
    mode_text=$(mode_display "$mode")
    
    echo ""
    echo -e "${GREEN}✅ ${service} 启动完成（${mode_text}）${NC}"
    case "$service" in
        etcd)
            if [ "$mode" = "cluster" ]; then
                echo "  - Etcd 集群端点: http://localhost:2379,2381,2383"
            else
                echo "  - Etcd 单节点端点: http://localhost:2379"
            fi
            ;;
        kafka)
            if [ "$mode" = "cluster" ]; then
                echo "  - Kafka 集群端点: localhost:9092,9094,9096"
            else
                echo "  - Kafka 单节点端点: localhost:9092"
            fi
            ;;
        redis)
            if [ "$mode" = "cluster" ]; then
                echo "  - Redis 集群端点: localhost:6380,6381,6382"
            else
                echo "  - Redis 单节点端点: localhost:6379"
                echo "  - Redis UI 未启动 (如需 UI: docker-compose up -d redis-commander)"
            fi
            ;;
        mongodb)
            if [ "$mode" = "cluster" ]; then
                echo "  - MongoDB 副本集端点: localhost:27018,27019,27020"
                echo "  - Mongo Express: http://localhost:8082"
            else
                echo "  - MongoDB 单节点端点: localhost:27017"
                echo "  - Mongo Express 未启动 (如需 UI: docker-compose up -d mongo-express)"
            fi
            ;;
        rocketmq)
            if [ "$mode" = "cluster" ]; then
                echo "  - RocketMQ NameServer: localhost:9876,9877"
                echo "  - RocketMQ Broker: localhost:10911,10921"
                echo "  - RocketMQ Console: http://localhost:8083"
            else
                echo "  - RocketMQ NameServer: localhost:9876"
                echo "  - RocketMQ Broker: localhost:10911"
                echo "  - 控制台未启动，如需 Web 管理请在 docker/rocketmq 目录执行:"
                echo "      docker-compose up -d rmqconsole"
            fi
            ;;
        otel)
            echo "  - Grafana: http://localhost:3000 (admin/admin123)"
            echo "  - Jaeger: http://localhost:16686"
            echo "  - Prometheus: http://localhost:9090"
            echo "  - Loki: http://localhost:3100"
            ;;
    esac
    echo ""
}

show_overall_info() {
    local mode="${1:-single}"
    echo ""
    echo -e "${GREEN}✅ 所有服务启动完成（$(mode_display "$mode")）${NC}"
    echo ""
    echo "📋 端点一览:"
    if [ "$mode" = "cluster" ]; then
        echo "  - etcd: http://localhost:2379,2381,2383"
        echo "  - kafka: localhost:9092,9094,9096"
        echo "  - redis: localhost:6380,6381,6382"
        echo "  - mongodb: localhost:27018,27019,27020"
        echo "  - rocketmq namesrv: localhost:9876,9877"
        echo "  - rocketmq broker: localhost:10911,10921"
        echo "  - rocketmq-console: http://localhost:8083"
        echo "  - mongo-express: http://localhost:8082"
    else
        echo "  - etcd: http://localhost:2379"
        echo "  - kafka: localhost:9092"
        echo "  - redis: localhost:6379"
        echo "  - mongodb: localhost:27017"
        echo "  - rocketmq: localhost:9876"
        echo "  - rocketmq-console: 未启动 (可手动 docker-compose up -d rmqconsole)"
        echo "  - mongo-express: 未启动 (可手动 docker-compose up -d mongo-express)"
        echo "  - OpenTelemetry: 未启动 (可手动选择菜单 11)"
    fi
    echo "  - redis-ui: 未启动 (可手动 docker-compose up -d redis-commander)"
    echo "  - grafana: http://localhost:3000 (admin/admin123)"
    echo "  - jaeger: http://localhost:16686"
    echo "  - prometheus: http://localhost:9090"
    echo "  - loki: http://localhost:3100"
    echo ""
    echo "🔧 管理命令:"
    echo "  docker/etcd/deploy.sh status"
    echo "  docker/kafka/deploy.sh status"
    echo "  docker/redis/deploy.sh status"
    echo "  docker/mongodb/deploy.sh status"
    echo "  docker/rocketmq/deploy.sh status"
    echo "  docker/otel/deploy.sh status"
    echo ""
}
echo -e "${GREEN}🚀 统一快速启动脚本${NC}"
echo "================================"

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}❌ Docker 未安装，请先安装 Docker${NC}"
    exit 1
fi

# 检查 Docker Compose
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
    echo -e "${BLUE}请选择操作:${NC}"
    echo " 1) 启动 etcd (单节点)"
    echo " 2) 启动 etcd (集群)"
    echo " 3) 启动 kafka (单节点)"
    echo " 4) 启动 kafka (集群)"
    echo " 5) 启动 redis (单节点)"
    echo " 6) 启动 redis (集群)"
    echo " 7) 启动 mongodb (单节点)"
    echo " 8) 启动 mongodb (副本集)"
    echo " 9) 启动 rocketmq (单节点)"
    echo "10) 启动 rocketmq (集群)"
    echo "11) 启动 otel (OpenTelemetry)"
    echo "12) 启动所有服务 (单节点模式)"
    echo "13) 启动所有服务 (集群模式)"
    echo "14) 停止所有服务"
    echo "15) 查看服务状态"
    echo " 0) 退出"
    echo ""
    read -p "请输入选择 (0-15): " choice
}

# 启动 etcd
start_etcd() {
    local mode="${1:-single}"
    echo -e "${GREEN}📦 启动 etcd...${NC}"
    pushd "$SCRIPT_DIR/etcd" > /dev/null
    if [ "$mode" = "cluster" ]; then
        if docker ps | grep -q "etcd-1"; then
            echo -e "${YELLOW}⚠️  etcd 集群已经在运行中${NC}"
        else
            ./deploy.sh cluster
        fi
    else
        if docker ps | grep -q "etcd-standalone"; then
            echo -e "${YELLOW}⚠️  etcd 已经在运行中${NC}"
        else
            ./deploy.sh standalone
        fi
    fi
    popd > /dev/null
}

# 启动 kafka
start_kafka() {
    local mode="${1:-single}"
    echo -e "${GREEN}📦 启动 kafka...${NC}"
    pushd "$SCRIPT_DIR/kafka" > /dev/null
    if [ "$mode" = "cluster" ]; then
        if docker ps | grep -E "kafka-[23]" > /dev/null; then
            echo -e "${YELLOW}⚠️  kafka 集群已经在运行中${NC}"
        else
            ./deploy.sh kraft
        fi
    else
        if docker ps | grep -q "kafka-1"; then
            echo -e "${YELLOW}⚠️  kafka 已经在运行中${NC}"
        else
            ./deploy.sh single
        fi
    fi
    popd > /dev/null
}

# 启动 redis
start_redis() {
    local mode="${1:-single}"
    echo -e "${GREEN}📦 启动 redis...${NC}"
    pushd "$SCRIPT_DIR/redis" > /dev/null
    if [ "$mode" = "cluster" ]; then
        if docker ps | grep -q "redis-2"; then
            echo -e "${YELLOW}⚠️  redis 集群已经在运行中${NC}"
        else
            ./deploy.sh cluster
        fi
    else
        if docker ps | grep -q "redis-standalone"; then
            echo -e "${YELLOW}⚠️  redis 已经在运行中${NC}"
        else
            ./deploy.sh standalone
        fi
    fi
    popd > /dev/null
}

# 启动 mongodb
start_mongodb() {
    local mode="${1:-single}"
    echo -e "${GREEN}📦 启动 mongodb...${NC}"
    pushd "$SCRIPT_DIR/mongodb" > /dev/null
    if [ "$mode" = "cluster" ]; then
        if docker ps | grep -q "mongodb-2"; then
            echo -e "${YELLOW}⚠️  mongodb 副本集已经在运行中${NC}"
        else
            ./deploy.sh replica
        fi
    else
        if docker ps | grep -q "mongodb-standalone"; then
            echo -e "${YELLOW}⚠️  mongodb 已经在运行中${NC}"
        else
            ./deploy.sh standalone
        fi
    fi
    popd > /dev/null
}

# 启动 rocketmq
start_rocketmq() {
    local mode="${1:-single}"
    echo -e "${GREEN}📦 启动 rocketmq...${NC}"
    pushd "$SCRIPT_DIR/rocketmq" > /dev/null
    if [ "$mode" = "cluster" ]; then
        if docker ps | grep -q "namesrv-2"; then
            echo -e "${YELLOW}⚠️  rocketmq 集群已经在运行中${NC}"
        else
            ./deploy.sh cluster
        fi
    else
        if docker ps | grep -q "rmqnamesrv"; then
            echo -e "${YELLOW}⚠️  rocketmq 已经在运行中${NC}"
        else
            ./deploy.sh standalone
        fi
    fi
    popd > /dev/null
}

# 启动 OpenTelemetry 监控栈
start_otel() {
    local mode="${1:-single}"
    local mode_text
    mode_text=$(mode_display "$mode")
    echo -e "${GREEN}📦 启动 OpenTelemetry 监控栈 (${mode_text})...${NC}"
    pushd "$SCRIPT_DIR/otel" > /dev/null
    # 检查是否已经运行
    if docker ps | grep -q "otel-collector"; then
        echo -e "${YELLOW}⚠️  OpenTelemetry 监控栈已经在运行中${NC}"
    else
        if [ "$mode" = "cluster" ]; then
            if ./deploy.sh start; then
                echo -e "${GREEN}✅ 完整版 OpenTelemetry 监控栈启动成功${NC}"
            else
                echo -e "${YELLOW}❌ OpenTelemetry 监控栈启动失败${NC}"
            fi
        else
            # 单节点默认使用简化版
            if ./deploy.sh start-simple; then
                echo -e "${GREEN}✅ 简化版 OpenTelemetry 监控栈启动成功${NC}"
            else
                echo -e "${YELLOW}❌ OpenTelemetry 监控栈启动失败${NC}"
            fi
        fi
    fi
    popd > /dev/null
}

# 启动全部服务
start_all() {
    local mode="${1:-single}"
    local mode_text
    mode_text=$(mode_display "$mode")
    echo -e "${GREEN}📦 启动所有服务 (${mode_text})...${NC}"
    start_etcd "$mode"
    start_kafka "$mode"
    start_redis "$mode"
    start_mongodb "$mode"
    start_rocketmq "$mode"
    if [ "$mode" = "cluster" ]; then
        start_otel "cluster"
    else
        echo -e "${YELLOW}⚠️  单节点模式默认不启动 OpenTelemetry 监控栈${NC}"
    fi
}

# 停止所有服务
stop_all() {
    echo -e "${YELLOW}🛑 停止所有服务...${NC}"
    
    # 停止 etcd
    echo -e "${YELLOW}停止 etcd...${NC}"
    pushd "$SCRIPT_DIR/etcd" > /dev/null
    ./deploy.sh stop
    popd > /dev/null
    
    # 停止 kafka
    echo -e "${YELLOW}停止 kafka...${NC}"
    pushd "$SCRIPT_DIR/kafka" > /dev/null
    ./deploy.sh stop
    popd > /dev/null
    
    # 停止 redis
    echo -e "${YELLOW}停止 redis...${NC}"
    pushd "$SCRIPT_DIR/redis" > /dev/null
    ./deploy.sh stop
    popd > /dev/null
    
    # 停止 mongodb
    echo -e "${YELLOW}停止 mongodb...${NC}"
    pushd "$SCRIPT_DIR/mongodb" > /dev/null
    ./deploy.sh stop
    popd > /dev/null
    
    # 停止 rocketmq
    echo -e "${YELLOW}停止 rocketmq...${NC}"
    pushd "$SCRIPT_DIR/rocketmq" > /dev/null
    ./deploy.sh stop
    popd > /dev/null
    
    # 停止 otel
    echo -e "${YELLOW}停止 OpenTelemetry 监控栈...${NC}"
    pushd "$SCRIPT_DIR/otel" > /dev/null
    ./deploy.sh stop
    popd > /dev/null
    
    echo -e "${GREEN}✅ 所有服务已停止${NC}"
}

# 查看服务状态
show_status() {
    echo -e "${BLUE}📊 服务状态:${NC}"
    echo ""
    
    # 检查 etcd
    if docker ps | grep -q "etcd-standalone"; then
        echo -e "${GREEN}✅ etcd: 单节点运行中${NC}"
    elif docker ps | grep -q "etcd-1"; then
        echo -e "${GREEN}✅ etcd: 集群运行中${NC}"
    else
        echo -e "${YELLOW}❌ etcd: 未运行${NC}"
    fi
    
    # 检查 kafka
    if docker ps | grep -E "kafka-[23]" > /dev/null; then
        echo -e "${GREEN}✅ kafka (集群): 运行中${NC}"
    elif docker ps | grep -q "kafka-1"; then
        echo -e "${GREEN}✅ kafka (单节点): 运行中${NC}"
    else
        echo -e "${YELLOW}❌ kafka: 未运行${NC}"
    fi
    
    # 检查 redis
    if docker ps | grep -q "redis-2"; then
        echo -e "${GREEN}✅ redis (集群): 运行中${NC}"
    elif docker ps | grep -q "redis-standalone"; then
        echo -e "${GREEN}✅ redis (单节点): 运行中${NC}"
    else
        echo -e "${YELLOW}❌ redis: 未运行${NC}"
    fi
    
    # 检查 mongodb
    if docker ps | grep -q "mongodb-2"; then
        echo -e "${GREEN}✅ mongodb (副本集): 运行中${NC}"
    elif docker ps | grep -q "mongodb-standalone"; then
        echo -e "${GREEN}✅ mongodb (单节点): 运行中${NC}"
    else
        echo -e "${YELLOW}❌ mongodb: 未运行${NC}"
    fi
    
    # 检查 rocketmq
    if docker ps | grep -q "namesrv-2"; then
        echo -e "${GREEN}✅ rocketmq (集群): 运行中${NC}"
    elif docker ps | grep -q "rmqnamesrv"; then
        echo -e "${GREEN}✅ rocketmq (单节点): 运行中${NC}"
    else
        echo -e "${YELLOW}❌ rocketmq: 未运行${NC}"
    fi
    
    # 检查 otel
    if docker ps | grep -q "otel-collector"; then
        echo -e "${GREEN}✅ otel: 运行中${NC}"
    else
        echo -e "${YELLOW}❌ otel: 未运行${NC}"
    fi
    
    echo ""
}

# 主循环
while true; do
    show_menu
    
    case $choice in
        1)
            start_etcd "single"
            show_service_info "etcd" "single"
            ;;
        2)
            start_etcd "cluster"
            show_service_info "etcd" "cluster"
            ;;
        3)
            start_kafka "single"
            show_service_info "kafka" "single"
            ;;
        4)
            start_kafka "cluster"
            show_service_info "kafka" "cluster"
            ;;
        5)
            start_redis "single"
            show_service_info "redis" "single"
            ;;
        6)
            start_redis "cluster"
            show_service_info "redis" "cluster"
            ;;
        7)
            start_mongodb "single"
            show_service_info "mongodb" "single"
            ;;
        8)
            start_mongodb "cluster"
            show_service_info "mongodb" "cluster"
            ;;
        9)
            start_rocketmq "single"
            show_service_info "rocketmq" "single"
            ;;
        10)
            start_rocketmq "cluster"
            show_service_info "rocketmq" "cluster"
            ;;
        11)
            start_otel
            show_service_info "otel" "single"
            ;;
        12)
            start_all "single"
            show_overall_info "single"
            ;;
        13)
            start_all "cluster"
            show_overall_info "cluster"
            ;;
        14)
            stop_all
            ;;
        15)
            show_status
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
done
