#!/bin/bash

# 安全的Pod重启脚本 - 逐个重启以应用新内存配置
echo "🔄 开始安全重启Pod以应用新内存配置..."

# 获取所有需要更新的Pod（1Gi内存限制的）
PODS_TO_RESTART=$(kubectl get pods -n cloud-ide-ws -o json | jq -r '.items[] | select(.spec.containers[0].resources.limits.memory == "1Gi") | .metadata.name')

if [ -z "$PODS_TO_RESTART" ]; then
    echo "✅ 没有需要重启的Pod"
    exit 0
fi

TOTAL_PODS=$(echo "$PODS_TO_RESTART" | wc -l)
echo "需要重启的Pod数量: $TOTAL_PODS"
echo ""


COUNTER=1
for pod in $PODS_TO_RESTART; do
    echo "🔄 [$COUNTER/$TOTAL_PODS] 重启Pod: $pod"
    
    # 删除Pod，让控制器重新创建（会使用新的LimitRange配置）
    kubectl delete pod $pod -n cloud-ide-ws --timeout=30s
    
    # 等待Pod重新创建
    echo "⏳ 等待Pod重新创建..."
    sleep 5
    
    # 检查是否重新创建成功
    if kubectl get pod $pod -n cloud-ide-ws &>/dev/null; then
        echo "✅ $pod 重启成功"
    else
        echo "⚠️  $pod 可能需要手动检查"
    fi
    
    ((COUNTER++))
    echo ""
    
    # 每重启5个Pod暂停一下，避免系统过载
    if [ $((COUNTER % 5)) -eq 1 ] && [ $COUNTER -gt 1 ]; then
        echo "💤 暂停10秒避免系统过载..."
        sleep 10
    fi
done

echo "✅ 所有Pod重启完成！"
echo "🔍 检查新配置是否生效..."
kubectl get pods -n cloud-ide-ws -o json | jq -r '.items[] | select(.spec.containers[0].resources.limits.memory == "2Gi") | .metadata.name' | wc -l
echo "个Pod现在使用2Gi内存配置"
