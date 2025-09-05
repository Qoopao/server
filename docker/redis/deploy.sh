#!/bin/bash

# Redis 部署管理脚本
# 支持单机版、集群版和哨兵模式

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
DOCKER_DIR="$SCRIPT_DIR"

# 显示帮助信息
show_help() {
    echo "Redis 部署管理脚本"
    echo ""
    echo "用法: $0 [命令] [选项]"
    echo ""
    echo "命令:"
    echo "  standalone    启动单机版 Redis"
    echo "  cluster       启动集群版 Redis"
    echo "  sentinel      启动哨兵模式 Redis"
    echo "  stop          停止所有 Redis 容器"
    echo "  restart       重启所有 Redis 容器"
    echo "  status        查看 Redis 容器状态"
    echo "  logs          查看 Redis 日志"
    echo "  clean         清理所有 Redis 容器和数据"
    echo "  test          测试 Redis 连接"
    echo "  info          查看 Redis 信息"
    echo "  monitor       监控 Redis 命令"
    echo "  help          显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 standalone          # 启动单机版"
    echo "  $0 cluster             # 启动集群版"
    echo "  $0 sentinel            # 启动哨兵模式"
    echo "  $0 status              # 查看状态"
    echo "  $0 test                # 测试连接"
}

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        log_info "请先安装 Docker: https://docs.docker.com/get-docker/"
        exit 1
    fi

    # 检查 Docker Compose
    if ! command -v docker-compose &> /dev/null; then
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

# 启动单机版 Redis
start_standalone() {
    log_info "启动单机版 Redis..."
    cd "$DOCKER_DIR"
    
    # 检查是否已经运行
    if docker-compose ps | grep -q "redis-standalone.*Up"; then
        log_info "单机版 Redis 已经在运行中"
        log_info "Redis 端点: localhost:6379"
        log_info "Redis UI: http://localhost:8081"
        return 0
    fi
    
    # 使用 Docker Compose 命令（暂时不启动 redis-commander）
    docker-compose up -d redis-standalone
    
    log_info "等待 Redis 启动..."
    sleep 3
    
    # 验证启动
    # 等待容器启动
    local retry_count=0
    while [ $retry_count -lt 3 ]; do
        if docker-compose ps | grep -q "redis-standalone.*Up"; then
            log_info "单机版 Redis 启动成功"
            log_info "Redis 端点: localhost:6379"
            log_info "注意: Redis UI 暂时不可用（镜像拉取问题）"
            return 0
        fi
        log_info "等待 Redis 容器完全启动... (重试 $((retry_count + 1))/3)"
        sleep 3
        retry_count=$((retry_count + 1))
    done
    
    log_error "单机版 Redis 启动失败"
    docker-compose logs redis-standalone
    exit 1
}

# 启动集群版 Redis
start_cluster() {
    log_info "启动集群版 Redis..."
    cd "$DOCKER_DIR"
    
    # 检查是否已经运行
    if docker-compose ps | grep -c "redis-[1-3].*Up" | grep -q "3"; then
        log_info "集群版 Redis 已经在运行中"
        log_info "Redis 集群端点:"
        log_info "  - redis-1: localhost:6380"
        log_info "  - redis-2: localhost:6381"
        log_info "  - redis-3: localhost:6382"
        return 0
    fi
    
    # 使用 Docker Compose 命令
    docker-compose up -d redis-1 redis-2 redis-3
    
    log_info "等待 Redis 集群启动..."
    sleep 5
    
    # 验证启动
    if docker-compose ps | grep -c "redis-[1-3].*Up" | grep -q "3"; then
        log_info "集群版 Redis 启动成功"
        log_info "Redis 集群端点:"
        log_info "  - redis-1: localhost:6380"
        log_info "  - redis-2: localhost:6381"
        log_info "  - redis-3: localhost:6382"
        
        # 初始化集群
        log_info "初始化 Redis 集群..."
        docker exec redis-1 redis-cli -a redis123 --cluster create \
            127.0.0.1:6380 127.0.0.1:6381 127.0.0.1:6382 \
            --cluster-replicas 0 --cluster-yes
    else
        log_error "集群版 Redis 启动失败"
        docker-compose logs redis-1 redis-2 redis-3
        exit 1
    fi
}

# 启动哨兵模式 Redis
start_sentinel() {
    log_info "启动哨兵模式 Redis..."
    cd "$DOCKER_DIR"
    
    docker-compose up -d redis-1 redis-sentinel-1 redis-sentinel-2 redis-sentinel-3
    
    log_info "等待 Redis 哨兵启动..."
    sleep 5
    
    # 验证启动
    if docker-compose ps | grep -c "redis.*Up" | grep -q "4"; then
        log_info "哨兵模式 Redis 启动成功"
        log_info "Redis 主节点: localhost:6380"
        log_info "Redis 哨兵:"
        log_info "  - sentinel-1: localhost:26379"
        log_info "  - sentinel-2: localhost:26380"
        log_info "  - sentinel-3: localhost:26381"
    else
        log_error "哨兵模式 Redis 启动失败"
        docker-compose logs redis-1 redis-sentinel-1
        exit 1
    fi
}

# 停止 Redis
stop_redis() {
    log_info "停止 Redis 容器..."
    cd "$DOCKER_DIR"
    
    docker-compose down
    
    log_info "Redis 容器已停止"
}

# 重启 Redis
restart_redis() {
    log_info "重启 Redis 容器..."
    stop_redis
    sleep 1
    start_standalone
}

# 查看状态
show_status() {
    log_info "Redis 容器状态:"
    echo ""
    
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep redis
}

# 查看日志
show_logs() {
    log_info "Redis 容器日志:"
    echo ""
    
    for container in $(docker ps --format "{{.Names}}" | grep redis); do
        log_info "$container 日志:"
        docker logs "$container" --tail 10
        echo ""
    done
}

# 清理容器和数据
clean_redis() {
    log_warn "这将删除所有 Redis 容器和数据，确定继续吗? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        log_info "清理 Redis 容器和数据..."
        cd "$DOCKER_DIR"
        
        docker-compose down -v
        docker rmi redis:7.2-alpine rediscommander/redis-commander:latest 2>/dev/null || true
        
        log_info "Redis 容器和数据已清理"
    else
        log_info "取消清理操作"
    fi
}

# 测试 Redis 连接
test_redis() {
    log_info "测试 Redis 连接..."
    
    # 检查是否有 Redis 容器运行
    if ! docker ps --format "{{.Names}}" | grep -q "redis"; then
        log_error "没有运行中的 Redis 容器"
        return 1
    fi
    
    # 测试单机版
    if docker ps -q -f name=redis-standalone | grep -q .; then
        log_info "测试单机版 Redis..."
        if docker exec redis-standalone redis-cli -a redis123 ping | grep -q "PONG"; then
            log_info "单机版 Redis 连接正常"
        else
            log_error "单机版 Redis 连接失败"
        fi
    fi
    
    # 测试集群版
    if docker ps -q -f name=redis-1 | grep -q .; then
        log_info "测试集群版 Redis..."
        if docker exec redis-1 redis-cli -a redis123 ping | grep -q "PONG"; then
            log_info "集群版 Redis 连接正常"
            
            # 显示集群信息
            log_info "集群信息:"
            docker exec redis-1 redis-cli -a redis123 cluster info
        else
            log_error "集群版 Redis 连接失败"
        fi
    fi
}

# 查看 Redis 信息
show_info() {
    log_info "Redis 信息:"
    echo ""
    
    if docker ps -q -f name=redis-standalone | grep -q .; then
        log_info "单机版 Redis 信息:"
        docker exec redis-standalone redis-cli -a redis123 info server
        docker exec redis-standalone redis-cli -a redis123 info memory
        docker exec redis-standalone redis-cli -a redis123 info stats
    fi
    
    if docker ps -q -f name=redis-1 | grep -q .; then
        log_info "集群版 Redis 信息:"
        docker exec redis-1 redis-cli -a redis123 cluster info
        docker exec redis-1 redis-cli -a redis123 cluster nodes
    fi
}

# 监控 Redis 命令
monitor_redis() {
    log_info "监控 Redis 命令 (按 Ctrl+C 停止)..."
    echo ""
    
    if docker ps -q -f name=redis-standalone | grep -q .; then
        docker exec redis-standalone redis-cli -a redis123 monitor
    elif docker ps -q -f name=redis-1 | grep -q .; then
        docker exec redis-1 redis-cli -a redis123 monitor
    else
        log_error "没有运行中的 Redis 容器"
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
        sentinel)
            start_sentinel
            ;;
        stop)
            stop_redis
            ;;
        restart)
            restart_redis
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        clean)
            clean_redis
            ;;
        test)
            test_redis
            ;;
        info)
            show_info
            ;;
        monitor)
            monitor_redis
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
