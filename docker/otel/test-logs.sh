#!/bin/bash

echo "🚀 测试 OpenTelemetry Collector 日志导出..."

# 检查 Collector 是否运行
echo "📋 检查 Collector 状态..."
docker ps | grep otel-collector

if [ $? -eq 0 ]; then
    echo "✅ Collector 正在运行"
    
    # 查看 Collector 日志
    echo "📝 查看 Collector 控制台日志..."
    echo "按 Ctrl+C 停止查看"
    docker logs -f otel-collector
else
    echo "❌ Collector 未运行，请先启动服务："
    echo "docker-compose up -d"
fi
