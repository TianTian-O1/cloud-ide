#!/bin/bash

# 集群工作空间智能自动清理脚本
# 作者: Cloud-IDE Auto-Cleanup System
# 版本: v1.0

set -e

# 配置变量
NAMESPACE="cloud-ide-ws"
CONFIG_MAP="workspace-cleanup-config"
LOG_FILE="/var/log/workspace-cleanup.log"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

error() {
    echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1" | tee -a "$LOG_FILE"
}

warn() {
    echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] WARN:${NC} $1" | tee -a "$LOG_FILE"
}

info() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')] INFO:${NC} $1" | tee -a "$LOG_FILE"
}

# 读取配置
get_config() {
    kubectl get configmap $CONFIG_MAP -n $NAMESPACE -o jsonpath="{.data.$1}" 2>/dev/null || echo ""
}

# 获取节点内存使用率
get_memory_usage() {
    kubectl describe nodes | grep "memory.*Mi.*%" | sed 's/.*(\([0-9]*\)%).*/\1/' | head -1
}

# 获取工作空间总数
get_workspace_count() {
    kubectl get pods -n $NAMESPACE --no-headers | grep "^ws-" | wc -l
}

# 获取旧工作空间列表
get_old_workspaces() {
    local max_age_hours=$1
    local batch_size=$2
    local exclude_patterns=$3
    
    # 构建排除模式的grep参数
    local exclude_grep=""
    if [ -n "$exclude_patterns" ]; then
        IFS=',' read -ra PATTERNS <<< "$exclude_patterns"
        for pattern in "${PATTERNS[@]}"; do
            exclude_grep="$exclude_grep | grep -v '$pattern'"
        done
    fi
    
    # 获取旧工作空间
    eval "kubectl get pods -n $NAMESPACE --sort-by='.metadata.creationTimestamp' --no-headers | grep '^ws-' $exclude_grep | awk '\$5 ~ /[${max_age_hours}-9][0-9]*h|[1-9][0-9]h/ {print \$1}' | head -$batch_size"
}

# 执行清理
perform_cleanup() {
    local workspaces_to_clean=("$@")
    local cleaned_count=0
    
    for workspace in "${workspaces_to_clean[@]}"; do
        info "🗑️ 清理工作空间: $workspace"
        
        if kubectl delete workspace "$workspace" -n $NAMESPACE --timeout=30s 2>/dev/null; then
            log "✅ 成功删除: $workspace"
            ((cleaned_count++))
            sleep 2  # 避免API服务器过载
        else
            error "❌ 删除失败: $workspace"
        fi
    done
    
    return $cleaned_count
}

# 主清理逻辑
main() {
    log "🚀 启动集群自动清理检查..."
    
    # 读取配置
    local max_age=$(get_config "max_workspace_age_hours")
    local batch_size=$(get_config "cleanup_batch_size")
    local memory_threshold=$(get_config "memory_threshold_percent")
    local min_workspaces=$(get_config "min_workspaces_to_keep")
    local exclude_patterns=$(get_config "exclude_patterns")
    
    # 默认值
    max_age=${max_age:-8}
    batch_size=${batch_size:-10}
    memory_threshold=${memory_threshold:-85}
    min_workspaces=${min_workspaces:-50}
    
    info "📊 配置参数: 最大年龄=${max_age}h, 批次大小=${batch_size}, 内存阈值=${memory_threshold}%, 最小保留=${min_workspaces}"
    
    # 检查当前状态
    local current_memory=$(get_memory_usage)
    local current_workspaces=$(get_workspace_count)
    
    info "📈 当前状态: 内存使用率=${current_memory}%, 工作空间数量=${current_workspaces}"
    
    # 判断是否需要清理
    local need_cleanup=false
    local cleanup_reason=""
    
    if [ "$current_memory" -gt "$memory_threshold" ]; then
        need_cleanup=true
        cleanup_reason="内存使用率${current_memory}%超过阈值${memory_threshold}%"
    elif [ "$current_workspaces" -gt 120 ]; then
        need_cleanup=true
        cleanup_reason="工作空间数量${current_workspaces}超过建议值120"
    fi
    
    if [ "$need_cleanup" = false ]; then
        log "✅ 系统状态良好，无需清理: 内存${current_memory}%, 工作空间${current_workspaces}个"
        return 0
    fi
    
    warn "🚨 触发清理条件: $cleanup_reason"
    
    # 安全检查 - 不能清理太多
    if [ "$current_workspaces" -le "$min_workspaces" ]; then
        warn "⚠️ 工作空间数量${current_workspaces}已达到最小保留值${min_workspaces}，跳过清理"
        return 0
    fi
    
    # 获取需要清理的工作空间
    info "🔍 查找超过${max_age}小时的旧工作空间..."
    local old_workspaces=($(get_old_workspaces "$max_age" "$batch_size" "$exclude_patterns"))
    
    if [ ${#old_workspaces[@]} -eq 0 ]; then
        info "✅ 未发现需要清理的旧工作空间"
        return 0
    fi
    
    info "📋 发现${#old_workspaces[@]}个需要清理的工作空间"
    
    # 执行清理
    log "🗑️ 开始执行清理操作..."
    perform_cleanup "${old_workspaces[@]}"
    local cleaned=$?
    
    # 等待清理生效
    sleep 10
    
    # 检查清理效果
    local new_memory=$(get_memory_usage)
    local new_workspaces=$(get_workspace_count)
    
    log "🎉 清理完成! 清理了${cleaned}个工作空间"
    log "📊 清理效果: 内存使用率 ${current_memory}% → ${new_memory}%, 工作空间 ${current_workspaces} → ${new_workspaces}"
    
    # 发送通知 (可扩展)
    if command -v curl >/dev/null 2>&1; then
        # 这里可以添加webhook通知
        info "📬 清理报告已记录"
    fi
}

# 健康检查
health_check() {
    if ! kubectl get nodes >/dev/null 2>&1; then
        error "❌ Kubernetes集群连接失败"
        exit 1
    fi
    
    if ! kubectl get namespace $NAMESPACE >/dev/null 2>&1; then
        error "❌ 命名空间 $NAMESPACE 不存在"
        exit 1
    fi
    
    log "✅ 健康检查通过"
}

# 入口点
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    # 创建日志目录
    mkdir -p "$(dirname "$LOG_FILE")"
    
    case "${1:-main}" in
        "health")
            health_check
            ;;
        "main"|"")
            health_check
            main
            ;;
        *)
            echo "用法: $0 [main|health]"
            exit 1
            ;;
    esac
fi
