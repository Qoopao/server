#!/bin/bash

# RocketMQ 部署管理脚本
# 支持单机版和集群版

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# 脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOCKER_DIR="$(dirname "$SCRIPT_DIR")"

# 显示帮助信息
show_help() {
    echo "RocketMQ 部署管理脚本"
    echo ""
    echo "用法: $0 [命令] [选项]"
    echo ""
    echo "命令:"
    echo "  standalone    启动单机版 RocketMQ"
    echo "  cluster       启动集群版 RocketMQ"
    echo "  stop          停止所有 RocketMQ 容器"
    echo "  restart       重启所有 RocketMQ 容器"
    echo "  status        查看 RocketMQ 容器状态"
    echo "  logs          查看 RocketMQ 日志"
    echo "  clean         清理所有 RocketMQ 容器和数据"
    echo "  test          测试 RocketMQ 连接"
    echo "  create-topic  创建测试主题"
    echo "  produce       生产测试消息"
    echo "  consume       消费测试消息"
    echo "  help          显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 standalone          # 启动单机版"
    echo "  $0 cluster             # 启动集群版"
    echo "  $0 status               # 查看状态"
    echo "  $0 test                 # 测试连接"
    echo "  $0 create-topic test    # 创建测试主题"
}

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        log_info "请先安装 Docker: https://docs.docker.com/get-docker/"
        exit 1
    fi

    # 检查 Docker Compose（支持新旧两种格式）
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose 未安装"
        log_info "请先安装 Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi

    # 检查Docker服务是否运行
    if ! docker info &> /dev/null; then
        log_error "Docker 服务未运行"
        log_info "请启动 Docker 服务: sudo systemctl start docker"
        exit 1
    fi
}

# 启动单机版 RocketMQ
start_standalone() {
    log_info "启动单机版 RocketMQ..."
    cd "$DOCKER_DIR"
    
    # 检查是否已经运行
    if command -v docker-compose &> /dev/null; then
        if docker-compose ps | grep -c "Up" | grep -q "3"; then
            log_info "RocketMQ 已经在运行中"
            log_info "Name Server: localhost:9876"
            log_info "Broker: localhost:10911"
            log_info "Console: http://localhost:8083"
            return 0
        fi
    else
        if docker compose ps | grep -c "Up" | grep -q "3"; then
            log_info "RocketMQ 已经在运行中"
            log_info "Name Server: localhost:9876"
            log_info "Broker: localhost:10911"
            log_info "Console: http://localhost:8083"
            return 0
        fi
    fi
    
    # 使用兼容的 Docker Compose 命令
    if command -v docker-compose &> /dev/null; then
        docker-compose up -d rmqnamesrv rmqbroker rmqconsole
    else
        docker compose up -d rmqnamesrv rmqbroker rmqconsole
    fi
    
    log_info "等待 RocketMQ 启动..."
    sleep 5
    
    # 验证启动
    if command -v docker-compose &> /dev/null; then
        # 等待所有容器启动
        local retry_count=0
        while [ $retry_count -lt 3 ]; do
            if docker-compose ps | grep -c "Up" | grep -q "3"; then
                log_info "单机版 RocketMQ 启动成功"
                log_info "Name Server: localhost:9876"
                log_info "Broker: localhost:10911"
                log_info "Console: http://localhost:8083"
                return 0
            fi
            log_info "等待 RocketMQ 容器完全启动... (重试 $((retry_count + 1))/3)"
            sleep 5
            retry_count=$((retry_count + 1))
        done
        
        log_error "单机版 RocketMQ 启动失败"
        docker-compose logs
        exit 1
    else
        # 等待所有容器启动
        local retry_count=0
        while [ $retry_count -lt 3 ]; do
            if docker compose ps | grep -c "Up" | grep -q "3"; then
                log_info "单机版 RocketMQ 启动成功"
                log_info "Name Server: localhost:9876"
                log_info "Broker: localhost:10911"
                log_info "Console: http://localhost:8083"
                return 0
            fi
            log_info "等待 RocketMQ 容器完全启动... (重试 $((retry_count + 1))/3)"
            sleep 5
            retry_count=$((retry_count + 1))
        done
        
        log_error "单机版 RocketMQ 启动失败"
        docker compose logs
        exit 1
    fi
}

# 启动集群版 RocketMQ
start_cluster() {
    log_info "启动集群版 RocketMQ..."
    cd "$DOCKER_DIR"
    
    # 检查是否已经运行
    if command -v docker-compose &> /dev/null; then
        if docker-compose ps | grep -c "Up" | grep -q "6"; then
            log_info "集群版 RocketMQ 已经在运行中"
            log_info "Name Server:"
            log_info "  - namesrv-1: localhost:9876"
            log_info "  - namesrv-2: localhost:9877"
            log_info "Broker:"
            log_info "  - broker-1: localhost:10911"
            log_info "  - broker-2: localhost:10921"
            log_info "Console: http://localhost:8083"
            return 0
        fi
    else
        if docker compose ps | grep -c "Up" | grep -q "6"; then
            log_info "集群版 RocketMQ 已经在运行中"
            log_info "Name Server:"
            log_info "  - namesrv-1: localhost:9876"
            log_info "  - namesrv-2: localhost:9877"
            log_info "Broker:"
            log_info "  - broker-1: localhost:10911"
            log_info "  - broker-2: localhost:10921"
            log_info "Console: http://localhost:8083"
            return 0
        fi
    fi
    
    # 使用兼容的 Docker Compose 命令
    if command -v docker-compose &> /dev/null; then
        docker-compose up -d
    else
        docker compose up -d
    fi
    
    log_info "等待 RocketMQ 集群启动..."
    sleep 8
    
    # 验证启动
    if command -v docker-compose &> /dev/null; then
        if docker-compose ps | grep -c "Up" | grep -q "6"; then
            log_info "集群版 RocketMQ 启动成功"
            log_info "Name Server:"
            log_info "  - namesrv-1: localhost:9876"
            log_info "  - namesrv-2: localhost:9877"
            log_info "Broker:"
            log_info "  - broker-1: localhost:10911"
            log_info "  - broker-2: localhost:10921"
            log_info "Console: http://localhost:8083"
        else
            log_error "集群版 RocketMQ 启动失败"
            docker-compose logs
            exit 1
        fi
    else
        if docker compose ps | grep -c "Up" | grep -q "6"; then
            log_info "集群版 RocketMQ 启动成功"
            log_info "Name Server:"
            log_info "  - namesrv-1: localhost:9876"
            log_info "  - namesrv-2: localhost:9877"
            log_info "Broker:"
            log_info "  - broker-1: localhost:10911"
            log_info "  - broker-2: localhost:10921"
            log_info "Console: http://localhost:8083"
        else
            log_error "集群版 RocketMQ 启动失败"
            docker compose logs
            exit 1
        fi
    fi
}

# 停止 RocketMQ
stop_rocketmq() {
    log_info "停止 RocketMQ 容器..."
    cd "$DOCKER_DIR"
    
    # 使用兼容的 Docker Compose 命令
    if command -v docker-compose &> /dev/null; then
        docker-compose down
    else
        docker compose down
    fi
    
    log_info "RocketMQ 容器已停止"
}

# 重启 RocketMQ
restart_rocketmq() {
    log_info "重启 RocketMQ 容器..."
    stop_rocketmq
    sleep 2
    start_standalone
}

# 查看状态
show_status() {
    log_info "RocketMQ 容器状态:"
    echo ""
    
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep rocketmq
}

# 查看日志
show_logs() {
    log_info "RocketMQ 容器日志:"
    echo ""
    
    for container in $(docker ps --format "{{.Names}}" | grep rocketmq); do
        log_info "$container 日志:"
        docker logs "$container" --tail 10
        echo ""
    done
}

# 清理容器和数据
clean_rocketmq() {
    log_warn "这将删除所有 RocketMQ 容器和数据，确定继续吗? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        log_info "清理 RocketMQ 容器和数据..."
        cd "$DOCKER_DIR"
        
        docker-compose down -v
        docker rmi apache/rocketmq:5.1.4 apacherocketmq/rocketmq-dashboard:1.0.0 2>/dev/null || true
        
        log_info "RocketMQ 容器和数据已清理"
    else
        log_info "取消清理操作"
    fi
}

# 测试 RocketMQ 连接
test_rocketmq() {
    log_info "测试 RocketMQ 连接..."
    
    # 检查是否有 RocketMQ 容器运行
    if ! docker ps --format "{{.Names}}" | grep -q "rocketmq"; then
        log_error "没有运行中的 RocketMQ 容器"
        return 1
    fi
    
    # 测试 Name Server
    if docker ps -q -f name=rmqnamesrv | grep -q .; then
        log_info "测试 Name Server..."
        if docker exec rmqnamesrv sh mqadmin clusterList -n localhost:9876; then
            log_info "Name Server 连接正常"
        else
            log_error "Name Server 连接失败"
        fi
    fi
    
    # 测试 Broker
    if docker ps -q -f name=rmqbroker | grep -q .; then
        log_info "测试 Broker..."
        if docker exec rmqbroker sh mqadmin brokerStatus -n localhost:9876 -b rmqbroker:10911; then
            log_info "Broker 连接正常"
        else
            log_error "Broker 连接失败"
        fi
    fi
}

# 创建主题
create_topic() {
    local topic_name="${1:-test-topic}"
    local partitions="${2:-4}"
    
    log_info "创建主题: $topic_name (分区: $partitions)"
    
    # 检查是否有 RocketMQ 容器运行
    if ! docker ps --format "{{.Names}}" | grep -q "rocketmq"; then
        log_error "没有运行中的 RocketMQ 容器"
        return 1
    fi
    
    # 创建主题
    if docker ps -q -f name=rmqnamesrv | grep -q .; then
        docker exec rmqnamesrv sh mqadmin updateTopic -n localhost:9876 -t "$topic_name" -c DefaultCluster -p "$partitions"
        log_info "主题 $topic_name 创建成功"
    else
        log_error "无法找到 Name Server 容器"
    fi
}

# 生产消息
produce_message() {
    local topic_name="${1:-test-topic}"
    local message="${2:-Hello RocketMQ!}"
    
    log_info "生产消息到主题: $topic_name"
    log_info "消息内容: $message"
    
    # 检查是否有 RocketMQ 容器运行
    if ! docker ps --format "{{.Names}}" | grep -q "rocketmq"; then
        log_error "没有运行中的 RocketMQ 容器"
        return 1
    fi
    
    # 生产消息
    if docker ps -q -f name=rmqnamesrv | grep -q .; then
        echo "$message" | docker exec -i rmqnamesrv sh mqadmin sendMessage -n localhost:9876 -t "$topic_name" -p "$message"
        log_info "消息生产成功"
    else
        log_error "无法找到 Name Server 容器"
    fi
}

# 消费消息
consume_message() {
    local topic_name="${1:-test-topic}"
    local group_name="${2:-test-group}"
    
    log_info "消费消息从主题: $topic_name (消费组: $group_name)"
    
    # 检查是否有 RocketMQ 容器运行
    if ! docker ps --format "{{.Names}}" | grep -q "rocketmq"; then
        log_error "没有运行中的 RocketMQ 容器"
        return 1
    fi
    
    # 消费消息
    if docker ps -q -f name=rmqnamesrv | grep -q .; then
        docker exec rmqnamesrv sh mqadmin printMsg -n localhost:9876 -t "$topic_name" -g "$group_name"
        log_info "消息消费完成"
    else
        log_error "无法找到 Name Server 容器"
    fi
}

# 主函数
main() {
    # 检查Docker
    check_docker
    
    # 解析参数
    case "${1:-help}" in
        standalone)
            start_standalone
            ;;
        cluster)
            start_cluster
            ;;
        stop)
            stop_rocketmq
            ;;
        restart)
            restart_rocketmq
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        clean)
            clean_rocketmq
            ;;
        test)
            test_rocketmq
            ;;
        create-topic)
            create_topic "${2:-test-topic}" "${3:-4}"
            ;;
        produce)
            produce_message "${2:-test-topic}" "${3:-Hello RocketMQ!}"
            ;;
        consume)
            consume_message "${2:-test-topic}" "${3:-test-group}"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
