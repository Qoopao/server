#!/bin/bash

# 停止所有服务的脚本

echo "停止所有服务..."

# 停止函数
stop_service() {
    local service_name=$1
    local pid_file="logs/${service_name}.pid"
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if kill -0 $pid 2>/dev/null; then
            echo "🛑 停止 ${service_name} (PID: ${pid})"
            kill $pid
            sleep 1
            
            # 如果还在运行，强制杀死
            if kill -0 $pid 2>/dev/null; then
                echo "⚠️  强制停止 ${service_name}"
                kill -9 $pid
            fi
        else
            echo "⚪ ${service_name} 未运行"
        fi
        rm -f "$pid_file"
    else
        echo "⚪ ${service_name} PID 文件不存在"
    fi
}

# 停止所有服务
stop_service "msggateway"
stop_service "msg-service"
stop_service "conversation-service"
stop_service "backservice"
stop_service "longconnection"
stop_service "msgconsumer"

# 等待端口完全释放
echo ""
echo "⏳ 等待端口释放..."
sleep 2

# 强制清理端口（如果还被占用）
for port in 10010 10100 10200 10300 6060 8956; do
    if lsof -ti:$port >/dev/null 2>&1; then
        echo "⚠️  强制释放端口 $port"
        lsof -ti:$port | xargs -r kill -9
    fi
done

echo "✅ 所有服务已停止，端口已释放"

