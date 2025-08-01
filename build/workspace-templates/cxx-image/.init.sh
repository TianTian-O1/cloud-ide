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
        # 生成zsh配置
        chsh -s /bin/zsh
        sh -c ./install.sh
        rm -f install.sh

        sed -i 's/BUILD=true/BUILD=false/' .init.sh
    else
        # 第一次启动工作空间,拷贝code-server的数据
        if [ ! -f "/root/.do_not_delete" ]; then
            echo "copy code-server"
            mv /.workspace/code-server-config.tar.gz /root
            cd /root
            tar zxvf code-server-config.tar.gz > /dev/null
            rm -f code-server-config.tar.gz
            touch /root/.do_not_delete
            cd -
        fi
    fi

    # 启动code-server
    node_id=$(ps aux | grep -E "/.workspace/code-server/lib/node /.workspace/code-server --port 9999 --host 0.0.0.0 --auth none" | grep -v grep)
    if [ -z "$node_id" ]; then
        nohup code-server --port 9999 --host 0.0.0.0 \
        --auth none --disable-update-check  --locale zh-cn \
        --open "$OPEN_DIR" &

        echo "start code-server success"
    fi

    sleep 5
done
