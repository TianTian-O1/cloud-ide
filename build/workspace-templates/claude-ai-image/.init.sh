#!/bin/bash 

BUILD=true

function graceful_exit () {
    echo "receive SIGTERM, exiting..."
    pid=$(ps -ef | grep code-server | awk '{print $2}')
    kill -SIGTERM "$pid"
    exit 0
}

trap graceful_exit SIGTERM

if [ -z "$OPEN_DIR" ]; then
	OPEN_DIR="$USER_WORKSPACE" 
fi

while :;do

    # 创建用户工作空间目录
    if [ ! -d "$USER_WORKSPACE" ]; then
        echo "create $USER_WORKSPACE"
        mkdir -p $USER_WORKSPACE
    fi

    # 构建阶段
    if [ "$BUILD" = true ]; then
        echo "Setting up Claude AI development environment..."
        
        # 生成zsh配置
        chsh -s /bin/zsh
        sh -c ./install.sh
        rm -f install.sh

        # 创建Go工作空间
        mkdir -p $GOPATH/src $GOPATH/bin $GOPATH/pkg

        # 验证Claude AI工具安装
        echo "Verifying Claude AI tools installation..."
        if command -v claude-code > /dev/null; then
            echo "✓ @anthropic-ai/claude-code installed successfully"
        else
            echo "⚠ @anthropic-ai/claude-code installation may have failed"
        fi

        if npm list -g @musistudio/claude-code-router > /dev/null 2>&1; then
            echo "✓ @musistudio/claude-code-router installed successfully"
        else
            echo "⚠ @musistudio/claude-code-router installation may have failed"
        fi

        # 创建示例配置文件
        cat > $USER_WORKSPACE/claude-ai-config.md << 'EOL'
# Claude AI Development Environment

## 已安装的Claude AI工具:

### 1. @anthropic-ai/claude-code
- Claude AI代码助手工具
- 使用: `claude-code --help`

### 2. @musistudio/claude-code-router  
- Claude代码路由工具
- 使用: `claude-code-router --help`

## 开发环境信息:

- Code-Server版本: 4.101.2
- Node.js版本: 22.x
- Go版本: 1.23.5
- Python版本: 3.10+

## 使用说明:

1. 打开VS Code后，可以直接使用Claude AI扩展
2. 支持多种编程语言开发
3. 预装了常用开发工具和库

## 快速开始:

```bash
# 验证安装
node --version
go version
python3 --version

# 创建新项目
mkdir my-project
cd my-project
```

欢迎使用Claude AI增强的云端IDE！
EOL

        sed -i 's/BUILD=true/BUILD=false/' .init.sh
        echo "Setup completed!"
    else
        # 第一次启动工作空间,拷贝code-server的数据
        if [ ! -f "/root/.do_not_delete" ]; then
            echo "Initializing code-server configuration..."
            
            # 如果存在配置文件，则恢复
            if [ -f "/.workspace/code-server-config.tar.gz" ]; then
                echo "Restoring code-server configuration..."
                mv /.workspace/code-server-config.tar.gz /root
                cd /root
                tar zxvf code-server-config.tar.gz > /dev/null
                rm -f code-server-config.tar.gz
                cd -
            fi
            
            touch /root/.do_not_delete
        fi
    fi

    # 启动code-server
    code_server_running=$(ps aux | grep -E "code-server.*--port 9999" | grep -v grep)
    if [ -z "$code_server_running" ]; then
        echo "Starting code-server with Claude AI tools..."
        nohup code-server \
            --port 9999 \
            --host 0.0.0.0 \
            --auth none \
            --disable-update-check \
            --locale zh-cn \
            --open "$OPEN_DIR" \
            --install-extension ms-python.python \
            --install-extension golang.go \
            --install-extension ms-vscode.vscode-typescript-next &

        echo "Code-server started successfully!"
        echo "Access your Claude AI development environment at: http://localhost:9999"
        echo "Workspace directory: $USER_WORKSPACE"
    fi

    sleep 5
done
