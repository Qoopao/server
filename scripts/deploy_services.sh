#!/bin/bash

# 分布式服务部署脚本
# 用于部署消息服务和会话服务

set -e

# 配置参数
MSG_SERVICE_NAME="msg-service"
CONVERSATION_SERVICE_NAME="conversation-service"
MSG_BASE_PORT=10100
CONVERSATION_BASE_PORT=10200
MSG_INSTANCE_COUNT=2
CONVERSATION_INSTANCE_COUNT=2
ETCD_ENDPOINTS=${ETCD_ENDPOINTS:-"localhost:2379"}
LOG_LEVEL=${LOG_LEVEL:-"info"}

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
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
    log_debug "检查端口占用情况..."
    for i in $(seq 0 $((MSG_INSTANCE_COUNT-1))); do
        port=$((MSG_BASE_PORT + i))
        if lsof -i:$port > /dev/null 2>&1; then
            log_warn "消息服务端口 $port 已被占用"
        fi
    done
    
    for i in $(seq 0 $((CONVERSATION_INSTANCE_COUNT-1))); do
        port=$((CONVERSATION_BASE_PORT + i))
        if lsof -i:$port > /dev/null 2>&1; then
            log_warn "会话服务端口 $port 已被占用"
        fi
    done
    
    log_info "依赖检查完成"
}

# 构建服务
build_services() {
    log_info "构建服务..."
    go build -o bin/msg-service main.go
    if [ $? -eq 0 ]; then
        log_info "消息服务构建成功"
    else
        log_error "消息服务构建失败"
        exit 1
    fi
}

# 启动消息服务实例
start_msg_instance() {
    local instance_id=$1
    local port=$((MSG_BASE_PORT + instance_id))
    
    log_info "启动消息服务实例 ${instance_id} (端口: ${port})"
    
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
    
    log_info "消息服务实例 ${instance_id} 已启动 (PID: ${pid})"
    
    # 等待服务启动
    sleep 3
    
    # 检查服务是否正常启动
    if kill -0 $pid 2>/dev/null; then
        log_info "消息服务实例 ${instance_id} 启动成功"
    else
        log_error "消息服务实例 ${instance_id} 启动失败"
        return 1
    fi
}

# 启动会话服务实例
start_conversation_instance() {
    local instance_id=$1
    local port=$((CONVERSATION_BASE_PORT + instance_id))
    
    log_info "启动会话服务实例 ${instance_id} (端口: ${port})"
    
    # 设置环境变量
    export ETCD_ENDPOINTS=$ETCD_ENDPOINTS
    export CONVERSATION_SERVICE_PORT=$port
    export LOG_LEVEL=$LOG_LEVEL
    
    # 创建日志目录
    mkdir -p logs
    
    # 启动服务
    nohup ./bin/msg-service > logs/conversation-service-${instance_id}.log 2>&1 &
    local pid=$!
    
    # 保存 PID
    echo $pid > logs/conversation-service-${instance_id}.pid
    
    log_info "会话服务实例 ${instance_id} 已启动 (PID: ${pid})"
    
    # 等待服务启动
    sleep 3
    
    # 检查服务是否正常启动
    if kill -0 $pid 2>/dev/null; then
        log_info "会话服务实例 ${instance_id} 启动成功"
    else
        log_error "会话服务实例 ${instance_id} 启动失败"
        return 1
    fi
}

# 停止所有实例
stop_all_instances() {
    log_info "停止所有服务实例..."
    
    # 停止消息服务实例
    for pid_file in logs/msg-service-*.pid; do
        if [ -f "$pid_file" ]; then
            pid=$(cat "$pid_file")
            if kill -0 $pid 2>/dev/null; then
                log_info "停止消息服务进程 $pid"
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
    
    # 停止会话服务实例
    for pid_file in logs/conversation-service-*.pid; do
        if [ -f "$pid_file" ]; then
            pid=$(cat "$pid_file")
            if kill -0 $pid 2>/dev/null; then
                log_info "停止会话服务进程 $pid"
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
    
    echo "消息服务:"
    for i in $(seq 0 $((MSG_INSTANCE_COUNT-1))); do
        pid_file="logs/msg-service-${i}.pid"
        port=$((MSG_BASE_PORT + i))
        
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
    
    echo "会话服务:"
    for i in $(seq 0 $((CONVERSATION_INSTANCE_COUNT-1))); do
        pid_file="logs/conversation-service-${i}.pid"
        port=$((CONVERSATION_BASE_PORT + i))
        
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
    local service_type=${1:-"msg"}
    local instance_id=${2:-0}
    local log_file="logs/${service_type}-service-${instance_id}.log"
    
    if [ -f "$log_file" ]; then
        log_info "显示 ${service_type} 服务实例 ${instance_id} 的日志:"
        tail -f "$log_file"
    else
        log_error "日志文件不存在: $log_file"
    fi
}

# 查看注册的服务
show_registered_services() {
    log_info "查看注册的服务实例:"
    
    # 检查 etcdctl 是否可用
    if command -v etcdctl >/dev/null 2>&1; then
        echo "消息服务实例:"
        etcdctl get --prefix /roc-im-server/services/msg-service/ 2>/dev/null || echo "  无实例"
        
        echo "会话服务实例:"
        etcdctl get --prefix /roc-im-server/services/conversation-service/ 2>/dev/null || echo "  无实例"
    else
        log_warn "etcdctl 命令不可用，无法查看注册的服务"
    fi
}

# 主函数
main() {
    case "${1:-start}" in
        "start")
            check_dependencies
            build_services
            
            log_info "启动 ${MSG_INSTANCE_COUNT} 个消息服务实例..."
            for i in $(seq 0 $((MSG_INSTANCE_COUNT-1))); do
                start_msg_instance $i
            done
            
            log_info "启动 ${CONVERSATION_INSTANCE_COUNT} 个会话服务实例..."
            for i in $(seq 0 $((CONVERSATION_INSTANCE_COUNT-1))); do
                start_conversation_instance $i
            done
            
            log_info "所有实例启动完成"
            show_status
            show_registered_services
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
            show_logs $2 $3
            ;;
            
        "services")
            show_registered_services
            ;;
            
        "build")
            build_services
            ;;
            
        *)
            echo "用法: $0 {start|stop|restart|status|logs [service_type] [instance_id]|services|build}"
            echo ""
            echo "命令说明:"
            echo "  start    - 启动所有服务实例"
            echo "  stop     - 停止所有服务实例"
            echo "  restart  - 重启所有服务实例"
            echo "  status   - 查看服务状态"
            echo "  logs     - 查看日志 (service_type: msg/conversation, instance_id: 0,1,2...)"
            echo "  services - 查看注册的服务实例"
            echo "  build    - 构建服务"
            echo ""
            echo "环境变量:"
            echo "  ETCD_ENDPOINTS           - Etcd 服务地址 (默认: localhost:2379)"
            echo "  LOG_LEVEL               - 日志级别 (默认: info)"
            echo "  MSG_INSTANCE_COUNT      - 消息服务实例数量 (默认: 2)"
            echo "  CONVERSATION_INSTANCE_COUNT - 会话服务实例数量 (默认: 2)"
            echo ""
            echo "示例:"
            echo "  $0 start                    # 启动所有服务"
            echo "  $0 logs msg 0               # 查看消息服务实例0的日志"
            echo "  $0 logs conversation 1     # 查看会话服务实例1的日志"
            exit 1
            ;;
    esac
}

# 创建必要的目录
mkdir -p bin logs

# 执行主函数
main "$@"
