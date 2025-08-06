#!/bin/bash

# 云IDE资源优化配置部署脚本
# 用途：应用优化的ResourceQuota和LimitRange配置

set -e

echo "🚀 开始部署云IDE资源优化配置..."
echo ""

# 检查必要文件
if [[ ! -f "workspace-resource-quota-optimized.yaml" ]]; then
    echo "❌ 错误：找不到 workspace-resource-quota-optimized.yaml"
    exit 1
fi

if [[ ! -f "workspace-limit-range-optimized.yaml" ]]; then
    echo "❌ 错误：找不到 workspace-limit-range-optimized.yaml"
    exit 1
fi

echo "📋 应用ResourceQuota配置..."
kubectl apply -f workspace-resource-quota-optimized.yaml

echo "📋 应用LimitRange配置..."
kubectl apply -f workspace-limit-range-optimized.yaml

echo ""
echo "✅ 配置应用完成！"
echo ""

echo "📊 当前配额状态："
kubectl get resourcequota workspace-resource-quota -n cloud-ide-ws -o custom-columns="RESOURCE:.spec.hard,USED:.status.used" --no-headers

echo ""
echo "🔍 LimitRange配置："
kubectl describe limitrange workspace-limit-range -n cloud-ide-ws | grep -A 10 "Container"

echo ""
echo "⚠️  注意："
echo "  - 新配置已应用，现有Pod将在重启后使用新的资源配置"
echo "  - 建议监控资源使用情况"
echo "  - 如需强制应用到现有Pod，可运行："
echo "    kubectl delete pods -n cloud-ide-ws --all"
echo ""
echo "🎯 部署完成！"
