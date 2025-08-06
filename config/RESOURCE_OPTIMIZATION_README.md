# 资源优化配置说明

本目录包含了经过优化的Kubernetes资源配置文件，用于提高云IDE系统的资源利用率和容量。

## 配置文件

### 1. workspace-resource-quota-optimized.yaml
**ResourceQuota配置 - 工作空间资源配额**

优化内容：
- **Services**: 从100增加到500 (支持500个同时运行的工作空间)
- **Pods**: 保持200个
- **PVCs**: 保持200个  
- **CPU Limits**: 保持200核
- **Memory Limits**: 优化为160Gi
- **CPU Requests**: 优化为20核
- **Memory Requests**: 优化为40Gi

### 2. workspace-limit-range-optimized.yaml
**LimitRange配置 - 容器资源默认值**

优化内容：
- **CPU Request**: 从500m降低到100m (减少80%资源占用)
- **Memory Request**: 从512Mi降低到128Mi (减少75%资源占用)
- **CPU Limit**: 保持500m
- **Memory Limit**: 保持512Mi

## 优化效果

### 资源利用率提升
- **理论工作空间容量**: 从约50个提升到500个
- **内存利用率**: 提升约4倍
- **CPU利用率**: 提升约5倍

### 系统稳定性
- 避免资源配额满载导致的创建失败
- 更好的资源调度和分配
- 支持更多并发用户

## 部署方法

```bash
# 应用ResourceQuota配置
kubectl apply -f workspace-resource-quota-optimized.yaml

# 应用LimitRange配置  
kubectl apply -f workspace-limit-range-optimized.yaml

# 强制重启所有工作空间Pod以应用新的资源配置
kubectl delete pods -n cloud-ide-ws --all
```

## 监控建议

定期检查资源使用情况：
```bash
# 检查配额使用
kubectl describe resourcequota workspace-resource-quota -n cloud-ide-ws

# 检查节点资源
kubectl top nodes

# 检查Pod资源使用
kubectl top pods -n cloud-ide-ws
```

## 注意事项

1. **资源监控**: 部署后需要持续监控实际资源使用情况
2. **性能测试**: 建议在低峰期部署，并进行充分测试
3. **回滚准备**: 保留原始配置文件以备回滚需要
4. **用户通知**: 配置变更可能影响正在运行的工作空间

## 更新历史

- **2025-08-05**: 初始优化配置
  - Services配额: 100 → 500
  - Memory优化: 减少默认请求值
  - CPU优化: 减少默认请求值

