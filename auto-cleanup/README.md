# 工作空间自动清理系统

## 🎯 功能特性

- **智能清理**: 自动识别长时间运行的工作空间并清理
- **资源监控**: 监控内存使用率，超过阈值时触发清理
- **安全保护**: 保证最小工作空间数量，防止过度清理
- **灵活配置**: 支持动态调整清理策略
- **定时调度**: 每30分钟自动检查一次
- **完整日志**: 详细的清理记录和状态报告

## 📋 系统组件

1. **ConfigMap**: `workspace-cleanup-config` - 清理策略配置
2. **CronJob**: `workspace-auto-cleanup` - 定时清理任务
3. **ServiceAccount**: `workspace-cleanup-sa` - 权限管理
4. **管理工具**: `cleanup-manager.sh` - 系统管理界面

## ⚙️ 配置参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| max_workspace_age_hours | 8 | 工作空间最大运行时间(小时) |
| cleanup_batch_size | 10 | 每次清理的最大数量 |
| memory_threshold_percent | 85 | 内存使用率阈值(%) |
| min_workspaces_to_keep | 50 | 最小保留工作空间数量 |
| cleanup_interval_minutes | 30 | 清理检查间隔(分钟) |

## 🚀 使用方法

### 查看系统状态
```bash
./cleanup-manager.sh status
```

### 手动触发清理
```bash
./cleanup-manager.sh trigger
```

### 查看清理日志
```bash
./cleanup-manager.sh logs
```

### 更新配置
```bash
./cleanup-manager.sh config
```

### 暂停/恢复自动清理
```bash
./cleanup-manager.sh toggle
```

## 📊 监控指标

系统会监控以下指标：
- 节点内存使用率
- 工作空间总数量
- 工作空间运行时间
- 清理成功率

## 🔒 安全特性

- **最小保留**: 确保至少保留50个工作空间
- **渐进清理**: 每次只清理有限数量，避免系统冲击
- **权限控制**: 使用专用ServiceAccount，最小权限原则
- **超时保护**: 任务10分钟超时，防止hang住

## 📈 清理策略

1. **触发条件**:
   - 内存使用率 > 85%
   - 工作空间数量 > 120个

2. **清理优先级**:
   - 运行时间最长的工作空间优先
   - 排除特殊命名的工作空间(admin-, test-, demo-)

3. **清理限制**:
   - 每次最多清理10个工作空间
   - 确保至少保留50个工作空间

## 🛠️ 故障排除

### 查看CronJob状态
```bash
kubectl get cronjob workspace-auto-cleanup -n cloud-ide-ws
```

### 查看最近的任务
```bash
kubectl get jobs -n cloud-ide-ws -l app=workspace-cleanup
```

### 查看详细日志
```bash
kubectl logs job/[job-name] -n cloud-ide-ws
```

### 手动清理历史任务
```bash
./cleanup-manager.sh cleanup
```

## ⚠️ 注意事项

1. 自动清理会删除长时间运行的工作空间，请确保用户了解此政策
2. 建议在业务低峰期进行清理操作
3. 定期检查清理日志，确认系统正常运行
4. 根据实际使用情况调整清理参数

## 📞 支持

如需帮助或报告问题，请检查：
1. 系统日志: `kubectl logs`
2. 配置状态: `./cleanup-manager.sh status`
3. 资源使用情况: `kubectl top nodes`
