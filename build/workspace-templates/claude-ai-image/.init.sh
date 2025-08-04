#!/bin/bash 

function graceful_exit () {
    echo "receive SIGTERM, exiting..."
    pkill -f code-server
    exit 0
}

trap graceful_exit SIGTERM

# 确保工作空间目录存在
mkdir -p /root/workspace

# 一次性设置（只在首次运行时）
if [ ! -f "/root/.claude_setup_done" ]; then
    echo "Quick setup..."
    
    # 设置zsh为默认shell
    chsh -s /bin/zsh
    
    # 创建Claude包装脚本来隐藏敏感信息
    cat > /usr/local/bin/claude-clean << 'CLAUDE_WRAPPER_EOF'
#!/bin/bash
# Claude包装脚本 - 隐藏API URL等敏感信息

# 临时重定向stderr来过滤敏感信息
exec 3>&2
exec 2> >(grep -v -E "(API Base URL|Overrides.*env|─────────────)" >&3)

# 运行原始claude命令
/usr/bin/claude "$@"

# 恢复stderr
exec 2>&3
exec 3>&-
CLAUDE_WRAPPER_EOF

    chmod +x /usr/local/bin/claude-clean
    
    # 创建Claude配置目录和配置文件
    mkdir -p /root/.claude
    
    # 配置Claude默认使用Claude 4 Sonnet模型
    cat > /root/.claude/settings.json << 'CLAUDE_CONFIG_EOF'
{
  "model": "claude-sonnet-4-20250514",
  "env": {
    "ANTHROPIC_MODEL": "claude-sonnet-4-20250514",
    "ANTHROPIC_SMALL_FAST_MODEL": "claude-sonnet-4-20250514"
  }
}
CLAUDE_CONFIG_EOF
    
    # 基础zsh配置
    cat > /root/.zshrc << 'EOF'
export PATH=/opt/code-server/bin:$PATH
export USER_WORKSPACE=/root/workspace
# Claude模型配置 - 默认使用Claude 4 Sonnet
export ANTHROPIC_MODEL="claude-sonnet-4-20250514"
export ANTHROPIC_SMALL_FAST_MODEL="claude-sonnet-4-20250514"
alias ll='ls -la'
alias la='ls -A'
alias l='ls -CF'
alias claude='claude-clean'  # 使用包装脚本
cd /root/workspace
echo "🚀 Claude AI Development Environment Ready!"
echo "📁 Workspace: /root/workspace"
echo "🔧 Tools: claude (clean output), claude-code-router available"
echo "🤖 Default Model: Claude 4 Sonnet (both primary and fast models)"
echo "💡 Usage: Run 'claude' to start Claude AI assistant (API info hidden)"
EOF
    
    touch /root/.claude_setup_done
    echo "Setup completed!"
fi

# 快速启动code-server
echo "Starting code-server..."
exec /opt/code-server/bin/code-server \
    --port 9999 \
    --host 0.0.0.0 \
    --auth none \
    --disable-update-check \
    --locale zh-cn \
    --open /root/workspace