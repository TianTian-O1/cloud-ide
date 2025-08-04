FROM ubuntu:20.04

WORKDIR /root

# 拷贝必要文件
COPY code-server/code-server-4.101.2-linux-amd64.tar.gz .
COPY claude-ai-image/.init.sh /opt/.init.sh
COPY claude-ai-image/install.sh /opt/install.sh
RUN chmod +x /opt/.init.sh /opt/install.sh

ENV TZ=Asia/Shanghai
ENV DEBIAN_FRONTEND=noninteractive

# 只安装必要的软件包 - 大幅减少安装项
RUN apt update && \
    apt install -y --no-install-recommends curl ca-certificates zsh git && \
    # 安装Node.js 18.x (Claude AI工具需要)
    curl -fsSL https://deb.nodesource.com/setup_18.x | bash - && \
    apt install -y nodejs && \
    # 设置时区
    ln -fs /usr/share/zoneinfo/$TZ /etc/localtime && \
    echo $TZ > /etc/timezone && \
    # 解压code-server
    tar zxf code-server-4.101.2-linux-amd64.tar.gz && \
    mv code-server-4.101.2-linux-amd64 /opt/code-server && \
    rm -f code-server-4.101.2-linux-amd64.tar.gz && \
    # 创建工作目录
    mkdir -p /root/workspace && \
    # 清理缓存
    apt clean && \
    apt autoremove -y && \
    rm -rf /var/lib/apt/lists/*

# 预装Claude AI工具 - 包含完整工具链
RUN npm config set registry https://registry.npmmirror.com && \
    npm install -g @anthropic-ai/claude-code @musistudio/claude-code-router && \
    npm cache clean --force

# 最小化环境变量
ENV USER_WORKSPACE=/root/workspace
ENV OPEN_DIR=/root/workspace
ENV CODE_SERVER_DIR=/opt/code-server
ENV PATH=$CODE_SERVER_DIR/bin:$PATH

EXPOSE 9999

CMD ["/bin/bash", "/opt/.init.sh"]