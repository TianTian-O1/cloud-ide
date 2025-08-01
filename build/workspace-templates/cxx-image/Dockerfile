FROM ubuntu:20.04

WORKDIR /.workspace

# 拷贝code-server
COPY code-server/code-server-4.9.0-linux-amd64.tar.gz .
# 拷贝字体文件
COPY fonts/SourceCodeVariable* ./
# 拷贝初始化脚本
COPY cxx-image/.init.sh .
# 拷贝zsh安装脚本
COPY cxx-image/install.sh .

ENV TZ=Asia/Shanghai
ENV DEBIAN_FRONTEND=noninteractive

# 安装必要的软件
RUN apt update && \
    apt install -y --no-install-recommends build-essential make cmake gdb valgrind vim wget curl netcat procps clangd git tar zip unzip fontconfig ca-certificates zsh && \
    apt install -y tzdata        && \
    ln -fs /usr/share/zoneinfo/$TZ /etc/localtime  && \
    echo $TZ > /etc/timezone && \
    mkdir -p ~/.fonts/source-code-pro        && \
    mv SourceCodeVariable* ~/.fonts/source-code-pro && \
    cd ~/.fonts/source-code-pro     &&\
    fc-cache -f -v           && \
    cd /.workspace           && \
    tar zxvf code-server-4.9.0-linux-amd64.tar.gz  && \
    mv code-server-4.9.0-linux-amd64 code-server   && \
    rm -f zxvf code-server-4.9.0-linux-amd64.tar.gz  && \
    apt clean && \
    apt autoremove -y && \
    rm -rf /var/cache/apk/* && \
    rm -rf /var/lib/apt/lists/*


# 用户工作空间目录
ENV USER_WORKSPACE /root/workspace
# 访问code-server默认打开目录
ENV OPEN_DIR /root/workspace
# code-server目录
ENV CODE_SERVER_DIR /.workspace/code-server
# code-server可执行文件加入PATH中
ENV PATH $CODE_SERVER_DIR/bin:$PATH

EXPOSE 9999

CMD ["/bin/bash", ".init.sh"]