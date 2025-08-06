#!/bin/bash

# 自动清理系统管理工具
# 用于管理、监控和配置自动清理系统

NAMESPACE="cloud-ide-ws"
CONFIG_MAP="workspace-cleanup-config"
CRONJOB_NAME="workspace-auto-cleanup"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# 显示状态
show_status() {
    echo -e "${BLUE}🔍 自动清理系统状态${NC}"
    echo "===================="
    
    echo -e "\n${GREEN}📅 CronJob状态:${NC}"
    kubectl get cronjob $CRONJOB_NAME -n $NAMESPACE
    
    echo -e "\n${GREEN}📊 最近的清理任务:${NC}"
    kubectl get jobs -n $NAMESPACE -l app=workspace-cleanup --sort-by='.metadata.creationTimestamp' | tail -5
    
    echo -e "\n${GREEN}⚙️ 当前配置:${NC}"
    kubectl get configmap $CONFIG_MAP -n $NAMESPACE -o yaml | grep -A 10 "data:"
    
    echo -e "\n${GREEN}💾 系统资源状态:${NC}"
    local memory_usage=$(kubectl describe nodes | grep "memory.*Mi.*%" | sed 's/.*(\([0-9]*\)%).*/\1/' | head -1)
    local workspace_count=$(kubectl get pods -n $NAMESPACE --no-headers | grep "^ws-" | wc -l)
    echo "   内存使用率: ${memory_usage}%"
    echo "   工作空间数量: ${workspace_count}个"
}

# 查看日志
show_logs() {
    echo -e "${BLUE}📝 最近的清理日志${NC}"
    echo "================"
    
    local latest_job=$(kubectl get jobs -n $NAMESPACE -l app=workspace-cleanup --sort-by='.metadata.creationTimestamp' -o jsonpath='{.items[-1].metadata.name}' 2>/dev/null)
    
    if [ -n "$latest_job" ]; then
        echo -e "\n${GREEN}最新任务: $latest_job${NC}"
        kubectl logs job/$latest_job -n $NAMESPACE
    else
        echo "暂无清理任务记录"
    fi
}

# 手动触发清理
trigger_cleanup() {
    echo -e "${YELLOW}🚀 手动触发清理任务${NC}"
    local job_name="manual-cleanup-$(date +%s)"
    kubectl create job $job_name --from=cronjob/$CRONJOB_NAME -n $NAMESPACE
    
    echo "清理任务已创建: $job_name"
    echo "等待任务完成..."
    
    # 等待任务完成
    kubectl wait --for=condition=complete job/$job_name -n $NAMESPACE --timeout=300s
    
    echo -e "\n${GREEN}📊 清理结果:${NC}"
    kubectl logs job/$job_name -n $NAMESPACE
}

# 更新配置
update_config() {
    echo -e "${PURPLE}⚙️ 更新清理配置${NC}"
    echo "==================="
    
    read -p "最大工作空间年龄(小时,当前:8): " max_age
    read -p "每次清理批次大小(当前:10): " batch_size
    read -p "内存阈值百分比(当前:85): " memory_threshold
    read -p "最小保留工作空间数(当前:50): " min_workspaces
    
    # 使用默认值
    max_age=${max_age:-8}
    batch_size=${batch_size:-10}
    memory_threshold=${memory_threshold:-85}
    min_workspaces=${min_workspaces:-50}
    
    # 更新ConfigMap
    kubectl patch configmap $CONFIG_MAP -n $NAMESPACE --type merge -p="{\"data\":{\"max_workspace_age_hours\":\"$max_age\",\"cleanup_batch_size\":\"$batch_size\",\"memory_threshold_percent\":\"$memory_threshold\",\"min_workspaces_to_keep\":\"$min_workspaces\"}}"
    
    echo -e "${GREEN}✅ 配置已更新${NC}"
    show_status
}

# 暂停/恢复自动清理
toggle_cleanup() {
    local current_suspend=$(kubectl get cronjob $CRONJOB_NAME -n $NAMESPACE -o jsonpath='{.spec.suspend}')
    
    if [ "$current_suspend" = "true" ]; then
        kubectl patch cronjob $CRONJOB_NAME -n $NAMESPACE -p '{"spec":{"suspend":false}}'
        echo -e "${GREEN}✅ 自动清理已恢复${NC}"
    else
        kubectl patch cronjob $CRONJOB_NAME -n $NAMESPACE -p '{"spec":{"suspend":true}}'
        echo -e "${YELLOW}⏸️ 自动清理已暂停${NC}"
    fi
}

# 清理历史任务
cleanup_history() {
    echo -e "${BLUE}🧹 清理历史任务记录${NC}"
    
    # 删除完成的任务(保留最近3个)
    kubectl get jobs -n $NAMESPACE -l app=workspace-cleanup --sort-by='.status.completionTime' -o jsonpath='{.items[:-3].metadata.name}' | xargs -r kubectl delete job -n $NAMESPACE
    
    echo -e "${GREEN}✅ 历史任务已清理${NC}"
}

# 卸载自动清理系统
uninstall() {
    echo -e "${RED}🗑️ 卸载自动清理系统${NC}"
    read -p "确认卸载自动清理系统? (y/N): " confirm
    
    if [[ $confirm =~ ^[Yy]$ ]]; then
        kubectl delete cronjob $CRONJOB_NAME -n $NAMESPACE
        kubectl delete configmap $CONFIG_MAP -n $NAMESPACE
        kubectl delete serviceaccount workspace-cleanup-sa -n $NAMESPACE
        kubectl delete clusterrole workspace-cleanup-role
        kubectl delete clusterrolebinding workspace-cleanup-binding
        
        echo -e "${GREEN}✅ 自动清理系统已卸载${NC}"
    else
        echo "取消卸载"
    fi
}

# 主菜单
show_menu() {
    echo -e "${PURPLE}🤖 工作空间自动清理管理器${NC}"
    echo "============================"
    echo "1. 查看系统状态"
    echo "2. 查看清理日志"
    echo "3. 手动触发清理"
    echo "4. 更新清理配置"
    echo "5. 暂停/恢复自动清理"
    echo "6. 清理历史任务"
    echo "7. 卸载自动清理系统"
    echo "8. 退出"
    echo ""
    read -p "请选择操作 (1-8): " choice
    
    case $choice in
        1) show_status ;;
        2) show_logs ;;
        3) trigger_cleanup ;;
        4) update_config ;;
        5) toggle_cleanup ;;
        6) cleanup_history ;;
        7) uninstall ;;
        8) echo "再见!" ; exit 0 ;;
        *) echo "无效选择" ;;
    esac
}

# 主程序
if [ $# -eq 0 ]; then
    while true; do
        show_menu
        echo ""
        read -p "按回车继续..."
        clear
    done
else
    case $1 in
        "status") show_status ;;
        "logs") show_logs ;;
        "trigger") trigger_cleanup ;;
        "config") update_config ;;
        "toggle") toggle_cleanup ;;
        "cleanup") cleanup_history ;;
        "uninstall") uninstall ;;
        *) echo "用法: $0 [status|logs|trigger|config|toggle|cleanup|uninstall]" ;;
    esac
fi
