#!/bin/bash
# 测试CloudBoot服务器启动

echo "🚀 启动CloudBoot服务器..."
/tmp/cloudboot-test &
PID=$!

echo "⏳ 等待服务器启动 (PID: $PID)..."
sleep 3

echo "🔍 测试健康检查端点..."
curl -s http://localhost:8080/health

echo ""
echo "🔍 测试主页..."
curl -s http://localhost:8080/ | head -20

echo ""
echo "✋ 停止服务器..."
kill $PID

echo "✅ 测试完成"
