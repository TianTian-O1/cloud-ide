#!/bin/bash

# 批量重启脚本 - 快速但会同时影响多个用户
echo "⚡ 批量重启所有旧配置Pod..."

# 获取所有需要更新的Pod
PODS_TO_RESTART=$(kubectl get pods -n cloud-ide-ws -o json | jq -r '.items[] | select(.spec.containers[0].resources.limits.memory == "1Gi") | .metadata.name')

if [ -z "$PODS_TO_RESTART" ]; then
    echo "✅ 没有需要重启的Pod"
    exit 0
fi

TOTAL_PODS=$(echo "$PODS_TO_RESTART" | wc -l)
echo "将要重启 $TOTAL_PODS 个Pod"
echo ""

read -p "⚠️  警告：这将同时重启所有旧配置Pod，是否继续？(y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "操作已取消"
    exit 0
fi

echo "🔄 开始批量重启..."
echo "$PODS_TO_RESTART" | xargs -I {} kubectl delete pod {} -n cloud-ide-ws

echo "⏳ 等待所有Pod重新创建..."
sleep 15

echo "✅ 批量重启完成！"
echo "🔍 检查结果..."
kubectl get pods -n cloud-ide-ws -o json | jq -r '.items[] | select(.spec.containers[0].resources.limits.memory == "2Gi") | .metadata.name' | wc -l
echo "个Pod现在使用2Gi内存配置"
