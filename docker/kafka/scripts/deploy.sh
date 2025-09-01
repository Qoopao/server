#!/bin/bash

# Kafka 部署管理脚本
# 支持 KRaft 模式和 ZooKeeper 模式

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
    echo "Kafka 部署管理脚本"
    echo ""
    echo "用法: $0 [命令] [选项]"
    echo ""
    echo "命令:"
    echo "  kraft         启动 KRaft 模式 Kafka (推荐)"
    echo "  zookeeper     启动 ZooKeeper 模式 Kafka"
    echo "  stop          停止所有 Kafka 容器"
    echo "  restart       重启所有 Kafka 容器"
    echo "  status        查看 Kafka 容器状态"
    echo "  logs          查看 Kafka 日志"
    echo "  clean         清理所有 Kafka 容器和数据"
    echo "  test          测试 Kafka 连接"
    echo "  create-topic  创建测试主题"
    echo "  produce       生产测试消息"
    echo "  consume       消费测试消息"
    echo "  help          显示此帮助信息"
    echo ""
    echo "选项:"
    echo "  -f, --file    指定docker-compose文件"
    echo "  -d, --detach  后台运行"
    echo ""
    echo "示例:"
    echo "  $0 kraft              # 启动 KRaft 模式"
    echo "  $0 zookeeper          # 启动 ZooKeeper 模式"
    echo "  $0 status             # 查看状态"
    echo "  $0 test               # 测试连接"
    echo "  $0 create-topic test  # 创建测试主题"
}

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        log_info "请先安装 Docker: https://docs.docker.com/get-docker/"
        exit 1
    fi

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

# 启动 KRaft 模式 Kafka
start_kraft() {
    log_info "启动 KRaft 模式 Kafka..."
    cd "$DOCKER_DIR"
    
    docker-compose -f docker-compose.kraft.yml up -d
    
    log_info "等待 Kafka 启动..."
    sleep 15
    
    # 验证启动
    if docker-compose -f docker-compose.kraft.yml ps | grep -c "Up" | grep -q "3"; then
        log_info "KRaft 模式 Kafka 启动成功"
        log_info "Kafka 端点:"
        log_info "  - kafka-1: localhost:9092"
        log_info "  - kafka-2: localhost:9094"
        log_info "  - kafka-3: localhost:9096"
        log_info "Kafka UI: http://localhost:8080"
    else
        log_error "KRaft 模式 Kafka 启动失败"
        docker-compose -f docker-compose.kraft.yml logs
        exit 1
    fi
}

# 启动 ZooKeeper 模式 Kafka
start_zookeeper() {
    log_info "启动 ZooKeeper 模式 Kafka..."
    cd "$DOCKER_DIR"
    
    # 检查是否有 ZooKeeper 配置文件
    if [ ! -f "docker-compose.zookeeper.yml" ]; then
        log_error "ZooKeeper 配置文件不存在"
        log_info "请先创建 docker-compose.zookeeper.yml 文件"
        exit 1
    fi
    
    docker-compose -f docker-compose.zookeeper.yml up -d
    
    log_info "等待 ZooKeeper 和 Kafka 启动..."
    sleep 20
    
    # 验证启动
    if docker-compose -f docker-compose.zookeeper.yml ps | grep -c "Up" | grep -q "6"; then
        log_info "ZooKeeper 模式 Kafka 启动成功"
        log_info "ZooKeeper 端点:"
        log_info "  - zookeeper-1: localhost:2181"
        log_info "  - zookeeper-2: localhost:2182"
        log_info "  - zookeeper-3: localhost:2183"
        log_info "Kafka 端点:"
        log_info "  - kafka-1: localhost:9092"
        log_info "  - kafka-2: localhost:9093"
        log_info "  - kafka-3: localhost:9094"
        log_info "Kafka UI: http://localhost:8080"
    else
        log_error "ZooKeeper 模式 Kafka 启动失败"
        docker-compose -f docker-compose.zookeeper.yml logs
        exit 1
    fi
}

# 停止 Kafka
stop_kafka() {
    log_info "停止 Kafka 容器..."
    cd "$DOCKER_DIR"
    
    # 停止所有 Kafka 相关容器
    docker-compose -f docker-compose.kraft.yml down 2>/dev/null || true
    docker-compose -f docker-compose.zookeeper.yml down 2>/dev/null || true
    
    log_info "Kafka 容器已停止"
}

# 重启 Kafka
restart_kafka() {
    log_info "重启 Kafka 容器..."
    stop_kafka
    sleep 2
    start_kraft
}

# 查看状态
show_status() {
    log_info "Kafka 容器状态:"
    echo ""
    
    # 检查 KRaft 模式
    if docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -q "kafka-[1-3]"; then
        log_info "KRaft 模式 Kafka:"
        docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep "kafka-[1-3]"
    fi
    
    # 检查 ZooKeeper 模式
    if docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -q "zookeeper"; then
        log_info "ZooKeeper:"
        docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep "zookeeper"
        log_info "Kafka (ZooKeeper 模式):"
        docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep "kafka-[1-3]"
    fi
    
    echo ""
    log_info "Kafka 网络:"
    docker network ls | grep kafka || log_warn "未找到 Kafka 网络"
}

# 查看日志
show_logs() {
    log_info "Kafka 容器日志:"
    echo ""
    
    # 检查 KRaft 模式
    if docker ps -q -f name=kafka-1 | grep -q .; then
        log_info "KRaft 模式 Kafka 日志:"
        for i in {1..3}; do
            if docker ps -q -f name=kafka-$i | grep -q .; then
                log_info "kafka-$i 日志:"
                docker logs kafka-$i --tail 10
                echo ""
            fi
        done
    fi
    
    # 检查 ZooKeeper 模式
    if docker ps -q -f name=zookeeper-1 | grep -q .; then
        log_info "ZooKeeper 日志:"
        for i in {1..3}; do
            if docker ps -q -f name=zookeeper-$i | grep -q .; then
                log_info "zookeeper-$i 日志:"
                docker logs zookeeper-$i --tail 10
                echo ""
            fi
        done
    fi
}

# 清理容器和数据
clean_kafka() {
    log_warn "这将删除所有 Kafka 容器和数据，确定继续吗? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        log_info "清理 Kafka 容器和数据..."
        cd "$DOCKER_DIR"
        
        # 停止并删除容器
        docker-compose -f docker-compose.kraft.yml down -v 2>/dev/null || true
        docker-compose -f docker-compose.zookeeper.yml down -v 2>/dev/null || true
        
        # 删除 Kafka 相关镜像
        docker rmi confluentinc/cp-kafka:7.4.0 2>/dev/null || true
        docker rmi confluentinc/cp-zookeeper:7.4.0 2>/dev/null || true
        docker rmi provectuslabs/kafka-ui:latest 2>/dev/null || true
        
        # 删除 Kafka 相关网络
        docker network rm docker_kafka-network 2>/dev/null || true
        
        log_info "Kafka 容器和数据已清理"
    else
        log_info "取消清理操作"
    fi
}

# 测试 Kafka 连接
test_kafka() {
    log_info "测试 Kafka 连接..."
    
    # 检查是否有 Kafka 容器运行
    if ! docker ps --format "{{.Names}}" | grep -q "kafka"; then
        log_error "没有运行中的 Kafka 容器"
        return 1
    fi
    
    # 测试 KRaft 模式
    if docker ps -q -f name=kafka-1 | grep -q .; then
        log_info "测试 KRaft 模式 Kafka..."
        
        # 创建测试主题
        docker exec kafka-1 kafka-topics --bootstrap-server kafka-1:29092 --create --topic test-topic --partitions 3 --replication-factor 3 2>/dev/null || true
        
        # 列出主题
        log_info "Kafka 主题列表:"
        docker exec kafka-1 kafka-topics --bootstrap-server kafka-1:29092 --list
        
        # 检查集群状态
        log_info "Kafka 集群状态:"
        docker exec kafka-1 kafka-broker-api-versions --bootstrap-server kafka-1:29092
        
        log_info "KRaft 模式 Kafka 连接正常"
    fi
    
    # 测试 ZooKeeper 模式
    if docker ps -q -f name=zookeeper-1 | grep -q .; then
        log_info "测试 ZooKeeper 模式 Kafka..."
        
        # 检查 ZooKeeper 连接
        echo stat | docker exec -i zookeeper-1 nc localhost 2181
        
        # 检查 Kafka 连接
        docker exec kafka-1 kafka-topics --bootstrap-server kafka-1:29092 --list
        
        log_info "ZooKeeper 模式 Kafka 连接正常"
    fi
}

# 创建主题
create_topic() {
    local topic_name="${1:-test-topic}"
    local partitions="${2:-3}"
    local replication="${3:-3}"
    
    log_info "创建主题: $topic_name (分区: $partitions, 副本: $replication)"
    
    # 检查是否有 Kafka 容器运行
    if ! docker ps --format "{{.Names}}" | grep -q "kafka"; then
        log_error "没有运行中的 Kafka 容器"
        return 1
    fi
    
    # 创建主题
    if docker ps -q -f name=kafka-1 | grep -q .; then
        docker exec kafka-1 kafka-topics --bootstrap-server kafka-1:29092 --create --topic "$topic_name" --partitions "$partitions" --replication-factor "$replication"
        log_info "主题 $topic_name 创建成功"
    else
        log_error "无法找到 Kafka 容器"
    fi
}

# 生产消息
produce_message() {
    local topic_name="${1:-test-topic}"
    local message="${2:-Hello Kafka!}"
    
    log_info "生产消息到主题: $topic_name"
    log_info "消息内容: $message"
    
    # 检查是否有 Kafka 容器运行
    if ! docker ps --format "{{.Names}}" | grep -q "kafka"; then
        log_error "没有运行中的 Kafka 容器"
        return 1
    fi
    
    # 生产消息
    if docker ps -q -f name=kafka-1 | grep -q .; then
        echo "$message" | docker exec -i kafka-1 kafka-console-producer --bootstrap-server kafka-1:29092 --topic "$topic_name"
        log_info "消息生产成功"
    else
        log_error "无法找到 Kafka 容器"
    fi
}

# 消费消息
consume_message() {
    local topic_name="${1:-test-topic}"
    local max_messages="${2:-10}"
    
    log_info "消费消息从主题: $topic_name (最多 $max_messages 条)"
    
    # 检查是否有 Kafka 容器运行
    if ! docker ps --format "{{.Names}}" | grep -q "kafka"; then
        log_error "没有运行中的 Kafka 容器"
        return 1
    fi
    
    # 消费消息
    if docker ps -q -f name=kafka-1 | grep -q .; then
        docker exec kafka-1 kafka-console-consumer --bootstrap-server kafka-1:29092 --topic "$topic_name" --from-beginning --max-messages "$max_messages"
        log_info "消息消费完成"
    else
        log_error "无法找到 Kafka 容器"
    fi
}

# 主函数
main() {
    # 检查Docker
    check_docker
    
    # 解析参数
    case "${1:-help}" in
        kraft)
            start_kraft
            ;;
        zookeeper)
            start_zookeeper
            ;;
        stop)
            stop_kafka
            ;;
        restart)
            restart_kafka
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        clean)
            clean_kafka
            ;;
        test)
            test_kafka
            ;;
        create-topic)
            create_topic "${2:-test-topic}" "${3:-3}" "${4:-3}"
            ;;
        produce)
            produce_message "${2:-test-topic}" "${3:-Hello Kafka!}"
            ;;
        consume)
            consume_message "${2:-test-topic}" "${3:-10}"
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
