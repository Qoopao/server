#!/bin/bash

# etcd 服务监控脚本
# 用于查看 etcd 中注册的所有服务实例

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

# 默认配置
ETCD_ENDPOINT="${ETCD_ENDPOINT:-localhost:2379}"
NAMESPACE="${ETCD_NAMESPACE:-/services}"

# 检查 etcd 容器
get_etcd_container() {
    if docker ps --format "{{.Names}}" | grep -qE "etcd-standalone|etcd-1"; then
        docker ps --format "{{.Names}}" | grep -E "etcd-standalone|etcd-1" | head -1
    else
        echo ""
    fi
}

# 执行 etcdctl 命令
run_etcdctl() {
    local CMD="$1"
    local CONTAINER=$(get_etcd_container)
    local OUTPUT=""
    local EXIT_CODE=0
    
    if [ -n "$CONTAINER" ]; then
        OUTPUT=$(docker exec "$CONTAINER" etcdctl $CMD 2>&1)
        EXIT_CODE=$?
    elif command -v etcdctl &> /dev/null; then
        OUTPUT=$(etcdctl --endpoints=$ETCD_ENDPOINT $CMD 2>&1)
        EXIT_CODE=$?
    else
        echo -e "${RED}❌ 未找到 etcdctl 或运行中的 etcd 容器${NC}"
        echo -e "${YELLOW}提示: 请确保 etcd 容器正在运行${NC}"
        exit 1
    fi
    
    # 检查是否有错误
    if [ $EXIT_CODE -ne 0 ]; then
        if echo "$OUTPUT" | grep -q "permission denied\|connect.*refused\|No such container"; then
            echo -e "${RED}❌ 无法连接到 etcd${NC}" >&2
            echo -e "${YELLOW}错误信息: $OUTPUT${NC}" >&2
            return 1
        fi
    fi
    
    echo "$OUTPUT"
    return $EXIT_CODE
}

# 扫描所有命名空间
scan_all_namespaces() {
    echo -e "${BLUE}🔍 扫描所有命名空间...${NC}"
    echo ""
    
    # 检查 etcd 连接
    local CONTAINER=$(get_etcd_container)
    if [ -z "$CONTAINER" ]; then
        echo -e "${RED}❌ 未找到运行中的 etcd 容器${NC}"
        echo -e "${YELLOW}提示: 请先启动 etcd 容器${NC}"
        echo "  cd docker/etcd && docker-compose up -d etcd-standalone"
        return 1
    fi
    
    # 获取所有键
    ALL_KEYS=$(run_etcdctl "get \"\" --prefix --keys-only" 2>/dev/null | grep -v "^$" || echo "")
    
    if [ -z "$ALL_KEYS" ]; then
        echo -e "${YELLOW}⚠️  etcd 中没有任何数据${NC}"
        echo ""
        echo -e "${CYAN}可能的原因:${NC}"
        echo "  1. 服务还没有注册到 etcd"
        echo "  2. 服务注册在不同的命名空间"
        echo ""
        echo -e "${CYAN}建议:${NC}"
        echo "  - 检查服务是否已启动: docker ps | grep -E 'backbon|message|sequence'"
        echo "  - 查看服务日志确认是否注册成功"
        echo "  - 尝试使用 Web UI: http://localhost:8084"
        return
    fi
    
    # 提取所有命名空间
    NAMESPACES=$(echo "$ALL_KEYS" | cut -d'/' -f1-2 | sort -u | sed 's|^|/|')
    
    if [ -z "$NAMESPACES" ]; then
        echo -e "${YELLOW}⚠️  未找到任何命名空间${NC}"
        return
    fi
    
    echo -e "${GREEN}✅ 找到以下命名空间:${NC}"
    for NS in $NAMESPACES; do
        echo -e "  ${CYAN}$NS${NC}"
    done
    echo ""
    
    # 遍历每个命名空间
    for NS in $NAMESPACES; do
        if [ "$NS" = "/" ]; then
            continue
        fi
        
        echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}命名空间: $NS${NC}"
        echo ""
        
        # 获取该命名空间下的所有键
        NS_KEYS=$(echo "$ALL_KEYS" | grep "^$NS/" || echo "")
        
        if [ -z "$NS_KEYS" ]; then
            echo -e "  ${YELLOW}⚠️  该命名空间下没有服务${NC}"
            echo ""
            continue
        fi
        
        # 提取服务名称
        SERVICE_NAMES=$(echo "$NS_KEYS" | sed "s|$NS/||" | cut -d'/' -f1 | sort -u)
        
        for SERVICE_NAME in $SERVICE_NAMES; do
            echo -e "  ${CYAN}服务: $SERVICE_NAME${NC}"
            OLD_NAMESPACE=$NAMESPACE
            NAMESPACE=$NS
            show_service_instances "$SERVICE_NAME"
            NAMESPACE=$OLD_NAMESPACE
            echo ""
        done
    done
}

# 显示所有服务
show_all_services() {
    # 如果命名空间是 "/" 或 "all"，扫描所有命名空间
    if [ "$NAMESPACE" = "/" ] || [ "$NAMESPACE" = "all" ]; then
        scan_all_namespaces
        return
    fi
    
    echo -e "${BLUE}📋 查看所有注册的服务...${NC}"
    echo -e "${CYAN}命名空间: $NAMESPACE${NC}"
    echo ""
    
    # 获取所有服务键
    KEYS=$(run_etcdctl "get \"$NAMESPACE\" --prefix --keys-only" 2>/dev/null | grep -v "^$" || echo "")
    
    if [ -z "$KEYS" ]; then
        echo -e "${YELLOW}⚠️  未找到任何注册的服务${NC}"
        echo -e "${YELLOW}   命名空间: $NAMESPACE${NC}"
        echo ""
        echo -e "${CYAN}提示: 尝试扫描所有命名空间:${NC}"
        echo "  $0 all"
        echo "  或指定其他命名空间:"
        echo "  ETCD_NAMESPACE=/long-connection-service $0"
        return
    fi
    
    # 提取服务名称
    SERVICE_NAMES=$(echo "$KEYS" | sed "s|$NAMESPACE/||" | cut -d'/' -f1 | sort -u)
    
    if [ -z "$SERVICE_NAMES" ]; then
        echo -e "${YELLOW}⚠️  未找到任何服务${NC}"
        return
    fi
    
    echo -e "${GREEN}✅ 找到以下服务:${NC}"
    echo ""
    
    for SERVICE_NAME in $SERVICE_NAMES; do
        echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}服务名称: $SERVICE_NAME${NC}"
        show_service_instances "$SERVICE_NAME"
        echo ""
    done
}

# 显示指定服务的实例
show_service_instances() {
    local SERVICE_NAME=$1
    local PREFIX="$NAMESPACE/$SERVICE_NAME/"
    
    # 获取所有实例
    OUTPUT=$(run_etcdctl "get \"$PREFIX\" --prefix" 2>/dev/null || echo "")
    
    if [ -z "$OUTPUT" ]; then
        echo -e "  ${YELLOW}⚠️  未找到实例${NC}"
        return
    fi
    
    # 解析输出（格式：key\nvalue\nkey\nvalue...）
    echo "$OUTPUT" | {
        local current_key=""
        local instance_count=0
        
        while IFS= read -r line; do
            if [[ $line == $PREFIX* ]]; then
                # 这是 key 行
                if [ -n "$current_key" ]; then
                    instance_count=$((instance_count + 1))
                fi
                current_key="$line"
                INSTANCE_ID=$(echo "$line" | sed "s|$PREFIX||")
            elif [[ $line =~ ^\{.*\} ]] || [[ $line =~ ^\" ]]; then
                # 这是 value 行（JSON）
                instance_count=$((instance_count + 1))
                echo -e "  ${CYAN}实例 #$instance_count:${NC}"
                echo -e "    ${YELLOW}InstanceID:${NC} $INSTANCE_ID"
                
                # 尝试解析 JSON（如果 python3 可用）
                if command -v python3 &> /dev/null; then
                    echo "$line" | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    print(f\"    ${YELLOW}ServiceName:${NC} {data.get('ServiceName', 'N/A')}\")
    print(f\"    ${YELLOW}Host:${NC} {data.get('Host', 'N/A')}\")
    print(f\"    ${YELLOW}Port:${NC} {data.get('Port', 'N/A')}\")
    print(f\"    ${YELLOW}Status:${NC} {data.get('Status', 'N/A')}\")
    print(f\"    ${YELLOW}Weight:${NC} {data.get('Weight', 'N/A')}\")
    if 'Version' in data:
        print(f\"    ${YELLOW}Version:${NC} {data.get('Version', 'N/A')}\")
except:
    print(f\"    ${YELLOW}Raw:${NC} {sys.stdin.read()[:100]}\")
" 2>/dev/null || echo -e "    ${YELLOW}Raw:${NC} ${line:0:100}..."
                else
                    # 简单显示
                    echo -e "    ${YELLOW}Data:${NC} ${line:0:100}..."
                fi
                echo ""
            fi
        done
        
        if [ $instance_count -eq 0 ]; then
            echo -e "  ${YELLOW}⚠️  未找到实例${NC}"
        fi
    }
}

# 显示指定服务
show_service() {
    local SERVICE_NAME=$1
    echo -e "${BLUE}📋 查看服务: $SERVICE_NAME${NC}"
    echo -e "${CYAN}命名空间: $NAMESPACE${NC}"
    echo ""
    show_service_instances "$SERVICE_NAME"
}

# 主函数
main() {
    case "${1:-all}" in
        all|list)
            show_all_services
            ;;
        scan)
            # 扫描所有命名空间
            NAMESPACE="/"
            scan_all_namespaces
            ;;
        check|diagnose)
            # 诊断 etcd 连接和服务状态
            echo -e "${BLUE}🔍 诊断 etcd 连接和服务状态...${NC}"
            echo ""
            
            # 检查容器
            local CONTAINER=$(get_etcd_container)
            if [ -z "$CONTAINER" ]; then
                echo -e "${RED}❌ etcd 容器未运行${NC}"
                echo -e "${YELLOW}解决方案:${NC}"
                echo "  cd docker/etcd && docker-compose up -d etcd-standalone"
                exit 1
            fi
            echo -e "${GREEN}✅ etcd 容器运行中: $CONTAINER${NC}"
            
            # 检查健康状态
            echo -e "${BLUE}检查 etcd 健康状态...${NC}"
            HEALTH=$(docker exec "$CONTAINER" etcdctl endpoint health 2>&1)
            if [ $? -eq 0 ]; then
                echo -e "${GREEN}✅ etcd 健康: $HEALTH${NC}"
            else
                echo -e "${RED}❌ etcd 不健康: $HEALTH${NC}"
            fi
            echo ""
            
            # 检查所有键
            echo -e "${BLUE}检查 etcd 中的数据...${NC}"
            ALL_KEYS=$(docker exec "$CONTAINER" etcdctl get "" --prefix --keys-only 2>&1 | grep -v "^$" || echo "")
            if [ -z "$ALL_KEYS" ]; then
                echo -e "${YELLOW}⚠️  etcd 中没有任何数据${NC}"
                echo ""
                echo -e "${CYAN}可能的原因:${NC}"
                echo "  1. 服务还没有启动并注册"
                echo "  2. 服务启动失败，未能注册到 etcd"
                echo ""
                echo -e "${CYAN}检查服务进程:${NC}"
                if command -v ps &> /dev/null; then
                    SERVICE_PROCESSES=$(ps aux 2>/dev/null | grep -E 'all-services|backbon|message|sequence|conversation' | grep -v grep || echo "")
                    if [ -n "$SERVICE_PROCESSES" ]; then
                        echo -e "${GREEN}✅ 找到运行中的服务进程:${NC}"
                        echo "$SERVICE_PROCESSES" | head -5 | sed 's/^/  /'
                    else
                        echo -e "${YELLOW}⚠️  未找到运行中的服务进程${NC}"
                    fi
                fi
                echo ""
                echo -e "${CYAN}建议:${NC}"
                echo "  1. 启动服务:"
                echo "     cd roc-im-server/cmd/all-services"
                echo "     make run"
                echo ""
                echo "  2. 等待服务启动完成（通常需要 5-10 秒）"
                echo ""
                echo "  3. 再次运行此命令检查:"
                echo "     $0 check"
                echo ""
                echo "  4. 或使用 Web UI 查看:"
                echo "     docker-compose up -d etcd-ui"
                echo "     访问 http://localhost:8084"
            else
                KEY_COUNT=$(echo "$ALL_KEYS" | wc -l)
                echo -e "${GREEN}✅ 找到 $KEY_COUNT 个键${NC}"
                echo ""
                echo -e "${CYAN}前 10 个键:${NC}"
                echo "$ALL_KEYS" | head -10 | sed 's/^/  /'
                if [ $KEY_COUNT -gt 10 ]; then
                    echo -e "  ${YELLOW}... 还有 $((KEY_COUNT - 10)) 个键${NC}"
                fi
                echo ""
                echo -e "${CYAN}提示: 运行 '$0 scan' 查看所有服务${NC}"
            fi
            ;;
        service|s)
            if [ -z "$2" ]; then
                echo -e "${RED}❌ 请指定服务名称${NC}"
                echo "用法: $0 service <服务名称>"
                exit 1
            fi
            show_service "$2"
            ;;
        help|--help|-h)
            echo "etcd 服务监控工具"
            echo ""
            echo "用法:"
            echo "  $0 [命令] [参数]"
            echo ""
            echo "命令:"
            echo "  all, list         显示指定命名空间的所有服务（默认: /services）"
            echo "  scan               扫描所有命名空间并显示所有服务"
            echo "  check, diagnose    诊断 etcd 连接和服务状态"
            echo "  service, s <名称>  显示指定服务的实例"
            echo "  help               显示帮助信息"
            echo ""
            echo "环境变量:"
            echo "  ETCD_ENDPOINT     etcd 端点（默认: localhost:2379）"
            echo "  ETCD_NAMESPACE    命名空间（默认: /services）"
            echo "                     使用 '/' 或 'all' 扫描所有命名空间"
            echo ""
            echo "示例:"
            echo "  $0                                    # 查看 /services 命名空间的服务"
            echo "  $0 scan                               # 扫描所有命名空间"
            echo "  $0 check                              # 诊断 etcd 连接状态"
            echo "  $0 all                                # 查看当前命名空间的服务"
            echo "  $0 service backbon-service             # 查看 backbon-service"
            echo "  ETCD_NAMESPACE=/long-connection-service $0  # 使用不同的命名空间"
            echo ""
            echo "Web UI:"
            echo "  访问 http://localhost:8084 使用 etcd-ui 管理界面"
            echo ""
            echo "故障排除:"
            echo "  1. 检查 etcd 容器是否运行: docker ps | grep etcd"
            echo "  2. 检查 etcd 健康状态: docker exec etcd-standalone etcdctl endpoint health"
            echo "  3. 查看所有键: docker exec etcd-standalone etcdctl get \"\" --prefix --keys-only"
            ;;
        *)
            echo -e "${RED}❌ 未知命令: $1${NC}"
            echo "使用 '$0 help' 查看帮助"
            exit 1
            ;;
    esac
}

main "$@"
