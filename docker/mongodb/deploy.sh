#!/bin/bash

# MongoDB 部署管理脚本
# 支持单机版、副本集和分片集群

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
    echo "MongoDB 部署管理脚本"
    echo ""
    echo "用法: $0 [命令] [选项]"
    echo ""
    echo "命令:"
    echo "  standalone    启动单机版 MongoDB"
    echo "  replica       启动副本集 MongoDB"
    echo "  shard         启动分片集群 MongoDB"
    echo "  stop          停止所有 MongoDB 容器"
    echo "  restart       重启所有 MongoDB 容器"
    echo "  status        查看 MongoDB 容器状态"
    echo "  logs          查看 MongoDB 日志"
    echo "  clean         清理所有 MongoDB 容器和数据"
    echo "  test          测试 MongoDB 连接"
    echo "  init          初始化数据库和用户"
    echo "  backup        备份数据库"
    echo "  restore       恢复数据库"
    echo "  help          显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 standalone          # 启动单机版"
    echo "  $0 replica             # 启动副本集"
    echo "  $0 status               # 查看状态"
    echo "  $0 test                 # 测试连接"
    echo "  $0 init                 # 初始化数据库"
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

# 启动单机版 MongoDB
start_standalone() {
    log_info "启动单机版 MongoDB..."
    cd "$DOCKER_DIR"
    
    # 检查是否已经运行
    if docker-compose ps | grep -q "mongodb-standalone.*Up"; then
        log_info "单机版 MongoDB 已经在运行中"
        log_info "MongoDB 端点: localhost:27017"
        log_info "MongoDB UI: http://localhost:8082"
        log_info "用户名: admin"
        log_info "密码: admin123"
        return 0
    fi
    
    # 使用 Docker Compose 命令（单节点默认不启动 Mongo Express）
    docker-compose up -d mongodb-standalone
    
    log_info "等待 MongoDB 启动..."
    sleep 3
    
    # 验证启动
    # 等待容器启动
    local retry_count=0
    while [ $retry_count -lt 3 ]; do
        if docker-compose ps | grep -q "mongodb-standalone.*Up"; then
            log_info "单机版 MongoDB 启动成功"
            log_info "MongoDB 端点: localhost:27017"
            log_info "Mongo Express 未启动，如需 Web UI 请在 docker/mongodb 目录执行:"
            log_info "  docker-compose up -d mongo-express"
            log_info "用户名: admin"
            log_info "密码: admin123"
            return 0
        fi
        log_info "等待 MongoDB 容器完全启动... (重试 $((retry_count + 1))/3)"
        sleep 5
        retry_count=$((retry_count + 1))
    done
    
    log_error "单机版 MongoDB 启动失败"
    docker-compose logs mongodb-standalone
    exit 1
}

# 启动副本集 MongoDB
start_replica() {
    log_info "启动副本集 MongoDB..."
    cd "$DOCKER_DIR"
    
    # 使用 Docker Compose 命令
    docker-compose up -d mongodb-1 mongodb-2 mongodb-3 mongo-express
    
    log_info "等待 MongoDB 副本集启动..."
    sleep 5
    
    # 验证启动
    if docker-compose ps | grep -c "mongodb-[1-3].*Up" | grep -q "3"; then
        log_info "副本集 MongoDB 启动成功"
        log_info "副本集端点:"
        log_info "  - mongodb-1: localhost:27018"
        log_info "  - mongodb-2: localhost:27019"
        log_info "  - mongodb-3: localhost:27020"
        log_info "MongoDB UI: http://localhost:8082"
        
        # 初始化副本集
        log_info "初始化副本集..."
        docker exec mongodb-1 mongosh --eval "$(cat replica-init.js)"
    else
        log_error "副本集 MongoDB 启动失败"
        docker-compose logs mongodb-1 mongodb-2 mongodb-3
        exit 1
    fi
}

# 启动分片集群 MongoDB
start_shard() {
    log_info "启动分片集群 MongoDB..."
    cd "$DOCKER_DIR"
    
    # 使用 Docker Compose 命令
    docker-compose up -d mongodb-1 mongodb-2 mongodb-3 mongodb-config-1 mongodb-config-2 mongodb-config-3 mongos mongo-express
    
    log_info "等待 MongoDB 分片集群启动..."
    sleep 8
    
    # 验证启动
    if docker-compose ps | grep -c "mongodb.*Up" | grep -q "7"; then
        log_info "分片集群 MongoDB 启动成功"
        log_info "分片集群端点:"
        log_info "  - mongos: localhost:27017"
        log_info "  - config servers: localhost:27018,27019,27020"
        log_info "  - shard servers: localhost:27021,27022,27023"
        log_info "MongoDB UI: http://localhost:8082"
    else
        log_error "分片集群 MongoDB 启动失败"
        docker-compose logs mongodb-1 mongodb-config-1 mongos
        exit 1
    fi
}

# 停止 MongoDB
stop_mongodb() {
    log_info "停止 MongoDB 容器..."
    cd "$DOCKER_DIR"
    
    # 使用 Docker Compose 命令
    docker-compose down
    
    log_info "MongoDB 容器已停止"
}

# 重启 MongoDB
restart_mongodb() {
    log_info "重启 MongoDB 容器..."
    stop_mongodb
    sleep 2
    start_standalone
}

# 查看状态
show_status() {
    log_info "MongoDB 容器状态:"
    echo ""
    
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep mongodb
}

# 查看日志
show_logs() {
    log_info "MongoDB 容器日志:"
    echo ""
    
    for container in $(docker ps --format "{{.Names}}" | grep mongodb); do
        log_info "$container 日志:"
        docker logs "$container" --tail 10
        echo ""
    done
}

# 清理容器和数据
clean_mongodb() {
    log_warn "这将删除所有 MongoDB 容器和数据，确定继续吗? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        log_info "清理 MongoDB 容器和数据..."
        cd "$DOCKER_DIR"
        
        # 使用 Docker Compose 命令
        docker-compose down -v
        
        docker rmi mongo:6 mongo-express:latest 2>/dev/null || true
        
        log_info "MongoDB 容器和数据已清理"
    else
        log_info "取消清理操作"
    fi
}

# 测试 MongoDB 连接
test_mongodb() {
    log_info "测试 MongoDB 连接..."
    
    # 检查是否有 MongoDB 容器运行
    if ! docker ps --format "{{.Names}}" | grep -q "mongodb"; then
        log_error "没有运行中的 MongoDB 容器"
        return 1
    fi
    
    # 测试单机版连接
    if docker ps -q -f name=mongodb-standalone | grep -q .; then
        log_info "测试单机版 MongoDB..."
        if docker exec mongodb-standalone mongosh --eval "db.runCommand('ping')" | grep -q "ok"; then
            log_info "单机版 MongoDB 连接正常"
        else
            log_error "单机版 MongoDB 连接失败"
        fi
    fi
    
    # 测试副本集连接
    if docker ps -q -f name=mongodb-1 | grep -q .; then
        log_info "测试副本集 MongoDB..."
        if docker exec mongodb-1 mongosh --eval "rs.status()" | grep -q "ok"; then
            log_info "副本集 MongoDB 连接正常"
        else
            log_error "副本集 MongoDB 连接失败"
        fi
    fi
}

# 初始化数据库
init_mongodb() {
    log_info "初始化 MongoDB 数据库..."
    
    # 检查是否有 MongoDB 容器运行
    if ! docker ps --format "{{.Names}}" | grep -q "mongodb"; then
        log_error "没有运行中的 MongoDB 容器"
        return 1
    fi
    
    # 初始化单机版
    if docker ps -q -f name=mongodb-standalone | grep -q .; then
        log_info "初始化单机版 MongoDB..."
        docker exec mongodb-standalone mongosh --eval "$(cat mongo-init.js)"
        log_info "单机版 MongoDB 初始化完成"
    fi
    
    # 初始化副本集
    if docker ps -q -f name=mongodb-1 | grep -q .; then
        log_info "初始化副本集 MongoDB..."
        docker exec mongodb-1 mongosh --eval "$(cat replica-init.js)"
        log_info "副本集 MongoDB 初始化完成"
    fi
}

# 备份数据库
backup_mongodb() {
    local backup_dir="${1:-./backup}"
    
    log_info "备份 MongoDB 数据库到: $backup_dir"
    
    # 创建备份目录
    mkdir -p "$backup_dir"
    
    # 备份单机版
    if docker ps -q -f name=mongodb-standalone | grep -q .; then
        log_info "备份单机版 MongoDB..."
        docker exec mongodb-standalone mongodump --out /tmp/backup
        docker cp mongodb-standalone:/tmp/backup "$backup_dir/standalone"
        docker exec mongodb-standalone rm -rf /tmp/backup
    fi
    
    # 备份副本集
    if docker ps -q -f name=mongodb-1 | grep -q .; then
        log_info "备份副本集 MongoDB..."
        docker exec mongodb-1 mongodump --out /tmp/backup
        docker cp mongodb-1:/tmp/backup "$backup_dir/replica"
        docker exec mongodb-1 rm -rf /tmp/backup
    fi
    
    log_info "MongoDB 备份完成"
}

# 恢复数据库
restore_mongodb() {
    local backup_dir="${1:-./backup}"
    
    log_info "从 $backup_dir 恢复 MongoDB 数据库"
    
    if [ ! -d "$backup_dir" ]; then
        log_error "备份目录不存在: $backup_dir"
        return 1
    fi
    
    # 恢复单机版
    if [ -d "$backup_dir/standalone" ] && docker ps -q -f name=mongodb-standalone | grep -q .; then
        log_info "恢复单机版 MongoDB..."
        docker cp "$backup_dir/standalone" mongodb-standalone:/tmp/backup
        docker exec mongodb-standalone mongorestore /tmp/backup
        docker exec mongodb-standalone rm -rf /tmp/backup
    fi
    
    # 恢复副本集
    if [ -d "$backup_dir/replica" ] && docker ps -q -f name=mongodb-1 | grep -q .; then
        log_info "恢复副本集 MongoDB..."
        docker cp "$backup_dir/replica" mongodb-1:/tmp/backup
        docker exec mongodb-1 mongorestore /tmp/backup
        docker exec mongodb-1 rm -rf /tmp/backup
    fi
    
    log_info "MongoDB 恢复完成"
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
        replica)
            start_replica
            ;;
        shard)
            start_shard
            ;;
        stop)
            stop_mongodb
            ;;
        restart)
            restart_mongodb
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        clean)
            clean_mongodb
            ;;
        test)
            test_mongodb
            ;;
        init)
            init_mongodb
            ;;
        backup)
            backup_mongodb "${2:-./backup}"
            ;;
        restore)
            restore_mongodb "${2:-./backup}"
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
