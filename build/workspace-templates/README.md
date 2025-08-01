# 生成Workspace模板镜像
build/workspace-templates目录用于构建Workspace的模板镜像，比如C++、Go、Java等的模板镜像
但是在这里只介绍C++和Go的构建方式，其它类似

## 下载依赖文件
build/workspace-templates/install.sh用于下载所有镜像构建依赖的程序

比如：code-server、Source Code Pro字体文件（个人认为这种字体写代码很舒服，因此如果不需要也可以不下载）

首先下载依赖的程序：
```shell
# 在build/workspace-template目录下执行
# 如果因为网络问题下载不了的，需要自己解决，也可以在windows上下载然后拷贝过来
./install.sh
```

## C++ Workspace镜像构建
cxx-image目录为构建C++工作空间镜像的文件夹

### step1: 构建镜像
```shell
# 在build/workspace-template目录下执行
docker build -t code-server-cxx:alpha -f cxx-image/Dockerfile .
```

### step2: 运行容器并对vscode进行配置
2.1 运行vscode容器
```shell
docker run -id -p 49999:9999 --name code-server-cxx-container -v ~/cxx_workspace:/root code-server-cxx:alpha
```
2.2 在浏览器中访问 yourip:49999，然后对vscode进行配置（字体、字体大小、外观、插件）,根据你自己的喜好安装插件

2.3 停止容器,将宿主机中~/cxx_workspace下的所有数据打包
```shell
docker stop code-server-cxx-container
cd ~/cxx_workspace
rm -rf workspace .zsh_history
tar zcvf code-server-config.tar.gz .
# 将配置文件移动到/root目录暂存
mv code-server-config.tar.gz ..
```

2.4 启动容器，并且将code-server-config.tar.gz拷贝到容器中
```shell
docker start code-server-cxx-container
docker cp ~/code-server-config.tar.gz code-server-cxx-container:/.workspace
```

2.5 停止容器，并且将容器提交为镜像
```shell
docker stop code-server-cxx-container
docker commit code-server-cxx-container code-server-cxx:v1.0
docker rm -f code-server-cxx-container
```
&nbsp;

经过以上的步骤，已经生成了一个C++工作空间的Docker镜像，通过这种方式生成的镜像的体积会小一点

## Go Workspace镜像构建
go-image目录为构建Go工作空间镜像的文件夹

### Step1: 构建镜像
```shell
# 在build/workspace-template目录下执行
docker build -t code-server-go:alpha -f go-image/Dockerfile .
```

### Step2: 运行容器并对vscode进行配置
2.1 运行容器
```shell
docker run -id -p 49999:9999 --name code-server-go-container -v ~/goxx_workspace:/root code-server-go:alpha
```

2.2 在浏览器中访问 yourip:49999，然后对vscode进行配置（字体、字体大小、外观、插件）,根据你自己的喜好安装插件

2.3 停止容器,将宿主机中~/goxx_workspace下的所有数据打包
```shell
docker stop code-server-go-container
cd ~/goxx_workspace
rm -rf workspace .zsh_history go/pkg/* .cache/*
tar zcvf code-server-config.tar.gz .
# 将配置文件移动到/root目录暂存
mv code-server-config.tar.gz ..
```

2.4 启动容器，并且将code-server-config.tar.gz拷贝到容器中
```shell
docker start code-server-go-container
docker cp ~/code-server-config.tar.gz code-server-go-container:/.workspace
```

2.5 停止容器，并且将容器提交为镜像
```shell
docker stop code-server-go-container
docker commit code-server-go-container code-server-go:v1.21
docker rm -f code-server-go-container
```