#!/bin/bash

# LinuxDo OAuth 配置测试脚本
# 用于验证 OAuth 配置是否正确

echo "=== LinuxDo OAuth 配置测试 ==="
echo

# 检查环境变量
echo "1. 检查环境变量配置..."
if [ -z "$LINUXDO_CLIENT_ID" ]; then
    echo "❌ LINUXDO_CLIENT_ID 未设置"
else
    echo "✅ LINUXDO_CLIENT_ID 已设置: ${LINUXDO_CLIENT_ID:0:10}..."
fi

if [ -z "$LINUXDO_CLIENT_SECRET" ]; then
    echo "❌ LINUXDO_CLIENT_SECRET 未设置"
else
    echo "✅ LINUXDO_CLIENT_SECRET 已设置: ${LINUXDO_CLIENT_SECRET:0:10}..."
fi

if [ -z "$LINUXDO_REDIRECT_URL" ]; then
    echo "❌ LINUXDO_REDIRECT_URL 未设置"
else
    echo "✅ LINUXDO_REDIRECT_URL 已设置: $LINUXDO_REDIRECT_URL"
fi

if [ -z "$LINUXDO_BASE_URL" ]; then
    echo "❌ LINUXDO_BASE_URL 未设置"
else
    echo "✅ LINUXDO_BASE_URL 已设置: $LINUXDO_BASE_URL"
fi

echo

# 检查配置文件
echo "2. 检查配置文件..."
CONFIG_FILE="config/webserver.yaml"
if [ -f "$CONFIG_FILE" ]; then
    echo "✅ 配置文件存在: $CONFIG_FILE"
    if grep -q "oauth:" "$CONFIG_FILE"; then
        echo "✅ 配置文件包含 OAuth 配置"
    else
        echo "⚠️  配置文件不包含 OAuth 配置"
    fi
else
    echo "⚠️  配置文件不存在: $CONFIG_FILE"
fi

echo

# 检查数据库表结构
echo "3. 检查数据库表结构..."
if command -v mysql >/dev/null 2>&1; then
    echo "✅ MySQL 客户端可用"
    echo "请手动运行以下命令检查表结构："
    echo "mysql -u username -p -e 'DESCRIBE t_user;' database_name"
else
    echo "⚠️  MySQL 客户端不可用，请手动检查数据库表结构"
fi

echo

# 检查服务状态
echo "4. 检查服务状态..."
if [ -n "$1" ]; then
    SERVER_URL="$1"
    echo "测试服务器: $SERVER_URL"
    
    # 测试 OAuth 状态端点
    if command -v curl >/dev/null 2>&1; then
        echo "测试 OAuth 状态端点..."
        RESPONSE=$(curl -s "${SERVER_URL}/auth/oauth/status" 2>/dev/null)
        if [ $? -eq 0 ]; then
            echo "✅ OAuth 状态端点响应: $RESPONSE"
        else
            echo "❌ OAuth 状态端点无响应"
        fi
    else
        echo "⚠️  curl 不可用，无法测试 API 端点"
    fi
else
    echo "⚠️  未提供服务器 URL，跳过 API 测试"
    echo "使用方法: $0 <server_url>"
    echo "例如: $0 https://your-domain.com"
fi

echo
echo "=== 测试完成 ==="
echo
echo "如果所有检查都通过，OAuth 配置应该可以正常工作。"
echo "如果有问题，请参考 docs/OAUTH_SETUP.md 文档。"