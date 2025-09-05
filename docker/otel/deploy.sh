#!/bin/bash

# OpenTelemetry 部署管理脚本
# 用于管理 OpenTelemetry Collector, Jaeger, Grafana 等服务的部署

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$SCRIPT_DIR"

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 Docker 和 Docker Compose
check_dependencies() {
    log_info "检查依赖..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# 启动所有服务
start_all() {
    log_info "启动 OpenTelemetry 监控栈..."
    cd "$PROJECT_DIR"
    
    # 创建必要的目录
    mkdir -p grafana/dashboards
    
    # 启动服务
    docker-compose up -d
    
    log_success "所有服务已启动"
    show_status
}

# 启动简化版本（仅核心组件）
start_simple() {
    log_info "启动简化版 OpenTelemetry 监控栈..."
    cd "$PROJECT_DIR"
    
    # 创建必要的目录
    mkdir -p grafana/dashboards
    
    # 使用简化配置启动服务
    docker-compose up -d
    
    log_success "简化版服务已启动"
    show_status
}

# 停止所有服务
stop_all() {
    log_info "停止 OpenTelemetry 监控栈..."
    cd "$PROJECT_DIR"
    
    docker-compose down
    
    log_success "所有服务已停止"
}

# 重启所有服务
restart_all() {
    log_info "重启 OpenTelemetry 监控栈..."
    stop_all
    sleep 2
    start_all
}

# 查看服务状态
show_status() {
    log_info "服务状态："
    cd "$PROJECT_DIR"
    
    docker-compose ps
    
    echo ""
    log_info "访问地址："
    echo "  - Grafana:     http://localhost:3000 (admin/admin123)"
    echo "  - Jaeger:      http://localhost:16686"
    echo "  - Prometheus:  http://localhost:9090"
    echo "  - Loki:        http://localhost:3100"
    echo "  - OTEL Collector Health: http://localhost:13133"
}

# 查看日志
show_logs() {
    local service=${1:-""}
    cd "$PROJECT_DIR"
    
    if [ -z "$service" ]; then
        log_info "显示所有服务日志..."
        docker-compose logs -f
    else
        log_info "显示 $service 服务日志..."
        docker-compose logs -f "$service"
    fi
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    # 检查 OpenTelemetry Collector
    if curl -s http://localhost:13133 > /dev/null; then
        log_success "OpenTelemetry Collector 健康检查通过"
    else
        log_error "OpenTelemetry Collector 健康检查失败"
    fi
    
    # 检查 Jaeger
    if curl -s http://localhost:16686 > /dev/null; then
        log_success "Jaeger 健康检查通过"
    else
        log_error "Jaeger 健康检查失败"
    fi
    
    # 检查 Grafana
    if curl -s http://localhost:3000 > /dev/null; then
        log_success "Grafana 健康检查通过"
    else
        log_error "Grafana 健康检查失败"
    fi
    
    # 检查 Prometheus
    if curl -s http://localhost:9090 > /dev/null; then
        log_success "Prometheus 健康检查通过"
    else
        log_error "Prometheus 健康检查失败"
    fi
}

# 清理数据
clean_data() {
    log_warning "这将删除所有持久化数据，包括 Grafana 仪表板、Prometheus 数据等"
    read -p "确定要继续吗？(y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "清理数据..."
        cd "$PROJECT_DIR"
        
        docker-compose down -v
        
        # 删除卷
        docker volume rm otel_grafana_data otel_prometheus_data otel_loki_data 2>/dev/null || true
        
        log_success "数据清理完成"
    else
        log_info "取消清理操作"
    fi
}

# 更新服务
update_services() {
    log_info "更新服务镜像..."
    cd "$PROJECT_DIR"
    
    docker-compose pull
    docker-compose up -d
    
    log_success "服务更新完成"
}

# 显示帮助信息
show_help() {
    echo "OpenTelemetry 监控栈管理脚本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  start        启动所有服务"
    echo "  start-simple 启动简化版服务（仅核心组件）"
    echo "  stop         停止所有服务"
    echo "  restart      重启所有服务"
    echo "  status       显示服务状态"
    echo "  logs         显示服务日志 [服务名]"
    echo "  health       执行健康检查"
    echo "  clean        清理所有数据"
    echo "  update       更新服务镜像"
    echo "  help         显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 start                    # 启动所有服务"
    echo "  $0 logs grafana            # 查看 Grafana 日志"
    echo "  $0 health                  # 执行健康检查"
}

# 主函数
main() {
    case "${1:-help}" in
        start)
            check_dependencies
            start_all
            ;;
        start-simple)
            check_dependencies
            start_simple
            ;;
        stop)
            stop_all
            ;;
        restart)
            check_dependencies
            restart_all
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs "$2"
            ;;
        health)
            health_check
            ;;
        clean)
            clean_data
            ;;
        update)
            check_dependencies
            update_services
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
