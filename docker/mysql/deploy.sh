#!/bin/bash

# MySQL 部署管理脚本
# 支持单机版和主从复制模式

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

MYSQL_ROOT_PASSWORD="mysql123"
REPL_PASSWORD="repl123"

# 显示帮助信息
show_help() {
    echo "MySQL 部署管理脚本"
    echo ""
    echo "用法: $0 [命令] [选项]"
    echo ""
    echo "命令:"
    echo "  standalone    启动单机版 MySQL"
    echo "  replication   启动主从复制 MySQL"
    echo "  stop          停止所有 MySQL 容器"
    echo "  restart       重启所有 MySQL 容器"
    echo "  status        查看 MySQL 容器状态"
    echo "  logs          查看 MySQL 日志"
    echo "  clean         清理所有 MySQL 容器和数据"
    echo "  test          测试 MySQL 连接"
    echo "  info          查看 MySQL 信息"
    echo "  init          初始化主从复制"
    echo "  help          显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 standalone          # 启动单机版"
    echo "  $0 replication         # 启动主从复制"
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

# 等待 MySQL 就绪
wait_for_mysql() {
    local container=$1
    local retry_count=0

    while [ $retry_count -lt 30 ]; do
        if docker exec "$container" mysqladmin ping -h localhost -uroot -p"$MYSQL_ROOT_PASSWORD" --silent &> /dev/null; then
            return 0
        fi
        log_info "等待 $container 就绪... (重试 $((retry_count + 1))/30)"
        sleep 2
        retry_count=$((retry_count + 1))
    done

    return 1
}

# 启动单机版 MySQL
start_standalone() {
    log_info "启动单机版 MySQL..."
    cd "$DOCKER_DIR"

    # 检查是否已经运行
    if docker-compose ps | grep -q "mysql-standalone.*Up"; then
        log_info "单机版 MySQL 已经在运行中"
        log_info "MySQL 端点: localhost:3306"
        log_info "Adminer UI: http://localhost:8084"
        log_info "Root 密码: $MYSQL_ROOT_PASSWORD"
        log_info "应用用户: app_user / app123"
        return 0
    fi

    docker-compose up -d mysql-standalone adminer

    log_info "等待 MySQL 启动..."
    if wait_for_mysql mysql-standalone; then
        log_info "单机版 MySQL 启动成功"
        log_info "MySQL 端点: localhost:3306"
        log_info "Adminer UI: http://localhost:8084"
        log_info "Root 密码: $MYSQL_ROOT_PASSWORD"
        log_info "应用用户: app_user / app123"
        log_info "数据库: app"
    else
        log_error "单机版 MySQL 启动失败"
        docker-compose logs mysql-standalone
        exit 1
    fi
}

# 配置从库复制
setup_slave_replication() {
    local slave_container=$1

    log_info "配置 $slave_container 复制..."

    docker exec "$slave_container" mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -e "
        STOP REPLICA;
        CHANGE REPLICATION SOURCE TO
            SOURCE_HOST='mysql-master',
            SOURCE_USER='repl',
            SOURCE_PASSWORD='$REPL_PASSWORD',
            SOURCE_AUTO_POSITION=1;
        START REPLICA;
    "
}

# 初始化主从复制
init_replication() {
    log_info "初始化主从复制..."

    if ! docker ps -q -f name=mysql-master | grep -q .; then
        log_error "主库 mysql-master 未运行"
        return 1
    fi

    for slave in mysql-slave-1 mysql-slave-2; do
        if docker ps -q -f name="$slave" | grep -q .; then
            if wait_for_mysql "$slave"; then
                setup_slave_replication "$slave"
            else
                log_error "$slave 未就绪"
                return 1
            fi
        fi
    done

    log_info "主从复制初始化完成"
}

# 启动主从复制 MySQL
start_replication() {
    log_info "启动主从复制 MySQL..."
    cd "$DOCKER_DIR"

    # 检查是否已经运行
    if docker-compose ps | grep -c "mysql.*Up" | grep -q "3"; then
        log_info "主从复制 MySQL 已经在运行中"
        log_info "MySQL 主从端点:"
        log_info "  - mysql-master: localhost:3307"
        log_info "  - mysql-slave-1: localhost:3308"
        log_info "  - mysql-slave-2: localhost:3309"
        return 0
    fi

    docker-compose up -d mysql-master mysql-slave-1 mysql-slave-2

    log_info "等待 MySQL 主从启动..."
    if wait_for_mysql mysql-master; then
        init_replication

        log_info "主从复制 MySQL 启动成功"
        log_info "MySQL 主从端点:"
        log_info "  - mysql-master: localhost:3307"
        log_info "  - mysql-slave-1: localhost:3308"
        log_info "  - mysql-slave-2: localhost:3309"
        log_info "Root 密码: $MYSQL_ROOT_PASSWORD"
        log_info "复制用户: repl / $REPL_PASSWORD"
    else
        log_error "主从复制 MySQL 启动失败"
        docker-compose logs mysql-master mysql-slave-1 mysql-slave-2
        exit 1
    fi
}

# 停止 MySQL
stop_mysql() {
    log_info "停止 MySQL 容器..."
    cd "$DOCKER_DIR"

    docker-compose down

    log_info "MySQL 容器已停止"
}

# 重启 MySQL
restart_mysql() {
    log_info "重启 MySQL 容器..."
    stop_mysql
    sleep 1
    start_standalone
}

# 查看状态
show_status() {
    log_info "MySQL 容器状态:"
    echo ""

    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "mysql|adminer" || true
}

# 查看日志
show_logs() {
    log_info "MySQL 容器日志:"
    echo ""

    for container in $(docker ps --format "{{.Names}}" | grep mysql); do
        log_info "$container 日志:"
        docker logs "$container" --tail 10
        echo ""
    done
}

# 清理容器和数据
clean_mysql() {
    log_warn "这将删除所有 MySQL 容器和数据，确定继续吗? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        log_info "清理 MySQL 容器和数据..."
        cd "$DOCKER_DIR"

        docker-compose down -v
        docker rmi mysql:8.0 adminer:latest 2>/dev/null || true

        log_info "MySQL 容器和数据已清理"
    else
        log_info "取消清理操作"
    fi
}

# 测试 MySQL 连接
test_mysql() {
    log_info "测试 MySQL 连接..."

    # 检查是否有 MySQL 容器运行
    if ! docker ps --format "{{.Names}}" | grep -q "mysql"; then
        log_error "没有运行中的 MySQL 容器"
        return 1
    fi

    # 测试单机版
    if docker ps -q -f name=mysql-standalone | grep -q .; then
        log_info "测试单机版 MySQL..."
        if docker exec mysql-standalone mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -e "SELECT 1" &> /dev/null; then
            log_info "单机版 MySQL 连接正常"
            docker exec mysql-standalone mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -e "SHOW DATABASES;"
        else
            log_error "单机版 MySQL 连接失败"
        fi
    fi

    # 测试主从复制
    if docker ps -q -f name=mysql-master | grep -q .; then
        log_info "测试主库 MySQL..."
        if docker exec mysql-master mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -e "SELECT 1" &> /dev/null; then
            log_info "主库 MySQL 连接正常"
        else
            log_error "主库 MySQL 连接失败"
        fi

        for slave in mysql-slave-1 mysql-slave-2; do
            if docker ps -q -f name="$slave" | grep -q .; then
                log_info "测试 $slave 复制状态..."
                docker exec "$slave" mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -e "SHOW REPLICA STATUS\G" | grep -E "Replica_IO_Running|Replica_SQL_Running|Last_Error" || true
            fi
        done
    fi
}

# 查看 MySQL 信息
show_info() {
    log_info "MySQL 信息:"
    echo ""

    if docker ps -q -f name=mysql-standalone | grep -q .; then
        log_info "单机版 MySQL 信息:"
        docker exec mysql-standalone mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -e "SHOW VARIABLES LIKE 'version%';"
        docker exec mysql-standalone mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -e "SHOW VARIABLES LIKE 'character_set%';"
        docker exec mysql-standalone mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -e "SHOW TABLE STATUS FROM app;"
    fi

    if docker ps -q -f name=mysql-master | grep -q .; then
        log_info "主库 MySQL 信息:"
        docker exec mysql-master mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -e "SHOW MASTER STATUS\G"
        docker exec mysql-master mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -e "SHOW TABLE STATUS FROM app;"
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
        replication)
            start_replication
            ;;
        stop)
            stop_mysql
            ;;
        restart)
            restart_mysql
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        clean)
            clean_mysql
            ;;
        test)
            test_mysql
            ;;
        info)
            show_info
            ;;
        init)
            init_replication
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
