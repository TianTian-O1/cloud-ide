FROM bitnami/git:latest

# 将脚本文件复制到镜像中
COPY git-clone.sh /git-clone.sh

# 设置脚本文件的执行权限
RUN chmod +x git-clone.sh


# 在容器启动时克隆
CMD ["/git-clone.sh"]
