#!/bin/bash

# 消息服务部署脚本
# 用于部署多个消息服务实例

set -e

# 配置参数
SERVICE_NAME="msg-service"
BASE_PORT=10100
INSTANCE_COUNT=3
ETCD_ENDPOINTS=${ETCD_ENDPOINTS:-"localhost:2379"}
LOG_LEVEL=${LOG_LEVEL:-"info"}

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查 Etcd 是否运行
    if ! curl -s ${ETCD_ENDPOINTS}/health > /dev/null 2>&1; then
        log_error "Etcd 服务未运行或不可访问: ${ETCD_ENDPOINTS}"
        log_info "请先启动 Etcd 服务："
        echo "docker run -d --name etcd -p 2379:2379 -p 2380:2380 quay.io/coreos/etcd:v3.5.0 etcd --advertise-client-urls http://0.0.0.0:2379 --listen-client-urls http://0.0.0.0:2379 --listen-peer-urls http://0.0.0.0:2380 --initial-advertise-peer-urls http://0.0.0.0:2380 --initial-cluster default=http://0.0.0.0:2380"
        exit 1
    fi
    
    # 检查端口是否可用
    for i in $(seq 0 $((INSTANCE_COUNT-1))); do
        port=$((BASE_PORT + i))
        if lsof -i:$port > /dev/null 2>&1; then
            log_warn "端口 $port 已被占用"
        fi
    done
    
    log_info "依赖检查完成"
}

# 构建服务
build_service() {
    log_info "构建消息服务..."
    go build -o bin/msg-service main.go
    if [ $? -eq 0 ]; then
        log_info "构建成功"
    else
        log_error "构建失败"
        exit 1
    fi
}

# 启动服务实例
start_instance() {
    local instance_id=$1
    local port=$((BASE_PORT + instance_id))
    
    log_info "启动实例 ${instance_id} (端口: ${port})"
    
    # 设置环境变量
    export ETCD_ENDPOINTS=$ETCD_ENDPOINTS
    export MSG_SERVICE_PORT=$port
    export LOG_LEVEL=$LOG_LEVEL
    
    # 创建日志目录
    mkdir -p logs
    
    # 启动服务
    nohup ./bin/msg-service > logs/msg-service-${instance_id}.log 2>&1 &
    local pid=$!
    
    # 保存 PID
    echo $pid > logs/msg-service-${instance_id}.pid
    
    log_info "实例 ${instance_id} 已启动 (PID: ${pid})"
    
    # 等待服务启动
    sleep 2
    
    # 检查服务是否正常启动
    if kill -0 $pid 2>/dev/null; then
        log_info "实例 ${instance_id} 启动成功"
    else
        log_error "实例 ${instance_id} 启动失败"
        return 1
    fi
}

# 停止所有实例
stop_all_instances() {
    log_info "停止所有消息服务实例..."
    
    for pid_file in logs/msg-service-*.pid; do
        if [ -f "$pid_file" ]; then
            pid=$(cat "$pid_file")
            if kill -0 $pid 2>/dev/null; then
                log_info "停止进程 $pid"
                kill $pid
                sleep 1
                # 强制杀死
                if kill -0 $pid 2>/dev/null; then
                    kill -9 $pid
                fi
            fi
            rm -f "$pid_file"
        fi
    done
    
    log_info "所有实例已停止"
}

# 查看服务状态
show_status() {
    log_info "服务状态:"
    
    for i in $(seq 0 $((INSTANCE_COUNT-1))); do
        pid_file="logs/msg-service-${i}.pid"
        port=$((BASE_PORT + i))
        
        if [ -f "$pid_file" ]; then
            pid=$(cat "$pid_file")
            if kill -0 $pid 2>/dev/null; then
                echo -e "  实例 ${i}: ${GREEN}运行中${NC} (PID: ${pid}, 端口: ${port})"
            else
                echo -e "  实例 ${i}: ${RED}已停止${NC} (端口: ${port})"
            fi
        else
            echo -e "  实例 ${i}: ${RED}未启动${NC} (端口: ${port})"
        fi
    done
}

# 查看日志
show_logs() {
    local instance_id=${1:-0}
    local log_file="logs/msg-service-${instance_id}.log"
    
    if [ -f "$log_file" ]; then
        log_info "显示实例 ${instance_id} 的日志:"
        tail -f "$log_file"
    else
        log_error "日志文件不存在: $log_file"
    fi
}

# 主函数
main() {
    case "${1:-start}" in
        "start")
            check_dependencies
            build_service
            
            log_info "启动 ${INSTANCE_COUNT} 个消息服务实例..."
            for i in $(seq 0 $((INSTANCE_COUNT-1))); do
                start_instance $i
            done
            
            log_info "所有实例启动完成"
            show_status
            ;;
            
        "stop")
            stop_all_instances
            ;;
            
        "restart")
            stop_all_instances
            sleep 2
            main start
            ;;
            
        "status")
            show_status
            ;;
            
        "logs")
            show_logs $2
            ;;
            
        "build")
            build_service
            ;;
            
        *)
            echo "用法: $0 {start|stop|restart|status|logs [instance_id]|build}"
            echo ""
            echo "命令说明:"
            echo "  start    - 启动所有服务实例"
            echo "  stop     - 停止所有服务实例"
            echo "  restart  - 重启所有服务实例"
            echo "  status   - 查看服务状态"
            echo "  logs     - 查看日志 (可指定实例ID)"
            echo "  build    - 构建服务"
            echo ""
            echo "环境变量:"
            echo "  ETCD_ENDPOINTS  - Etcd 服务地址 (默认: localhost:2379)"
            echo "  LOG_LEVEL       - 日志级别 (默认: info)"
            exit 1
            ;;
    esac
}

# 创建必要的目录
mkdir -p bin logs

# 执行主函数
main "$@"
