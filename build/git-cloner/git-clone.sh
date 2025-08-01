#!/bin/bash

# 从环境变量中获取仓库URL和本地路径
repo_url="$REPO_URL"
local_path="$LOCAL_PATH"

echo "$repo_url"
echo "$local_path"

# 检查本地仓库是否存在
if [ -d "$local_path" ]; then
    # 本地仓库不存在，执行克隆操作
    echo "Local repository already exists."
	exit 0
fi

# 尝试克隆仓库
git clone "$repo_url" "$local_path"

if [ $? -ne 0 ]; then
	echo "Failed to clone repository."
	exit 1
fi

echo "Repository cloned successfully."
exit 0
