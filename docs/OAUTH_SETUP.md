# LinuxDo OAuth 集成配置指南

本文档介绍如何为 Cloud Code 配置 LinuxDo OAuth 登录功能。

## 概述

LinuxDo OAuth 集成允许用户使用他们的 LinuxDo 账号登录 Cloud Code，无需单独注册账号。

## 功能特性

- ✅ 使用 LinuxDo 账号一键登录
- ✅ 自动创建新用户或关联现有用户
- ✅ 支持用户头像和昵称同步
- ✅ 安全的 OAuth 2.0 授权流程
- ✅ 状态验证防止 CSRF 攻击

## 前置条件

1. 在 LinuxDo 平台注册并创建 OAuth 应用
2. 获取 Client ID 和 Client Secret
3. 配置正确的回调地址

## LinuxDo OAuth 应用配置

### 1. 创建 OAuth 应用

1. 访问 [LinuxDo 开发者页面](https://connect.linux.do)
2. 登录您的 LinuxDo 账号
3. 创建新的 OAuth 应用
4. 填写应用信息：
   - **应用名称**: Cloud Code
   - **应用描述**: 云端代码编辑器
   - **回调地址**: `https://your-domain.com/auth/oauth/linuxdo/callback`

### 2. 获取凭据

创建应用后，您将获得：
- **Client ID**: 应用的公开标识符
- **Client Secret**: 应用的私密密钥（请妥善保管）

## 服务端配置

### 方式1: 环境变量配置（推荐）

```bash
# LinuxDo OAuth 配置
export LINUXDO_CLIENT_ID="your_client_id_here"
export LINUXDO_CLIENT_SECRET="your_client_secret_here"
export LINUXDO_REDIRECT_URL="https://your-domain.com/auth/oauth/linuxdo/callback"
export LINUXDO_BASE_URL="https://connect.linux.do"
```

### 方式2: 配置文件配置

编辑 `config/webserver.yaml`:

```yaml
oauth:
  linuxdo:
    client_id: "your_client_id_here"
    client_secret: "your_client_secret_here"
    redirect_url: "https://your-domain.com/auth/oauth/linuxdo/callback"
    base_url: "https://connect.linux.do"
```

## 数据库迁移

运行以下 SQL 脚本为用户表添加 OAuth 支持字段：

```sql
-- 添加LinuxDo OAuth字段
ALTER TABLE t_user 
ADD COLUMN linuxdo_id INT NULL COMMENT 'LinuxDo用户ID',
ADD COLUMN linuxdo_username VARCHAR(100) NULL COMMENT 'LinuxDo用户名';

-- 添加索引
CREATE UNIQUE INDEX idx_linuxdo_id ON t_user (linuxdo_id);
CREATE INDEX idx_linuxdo_username ON t_user (linuxdo_username);
```

或者运行提供的迁移脚本：

```bash
mysql -u username -p database_name < sql/add_oauth_fields.sql
```

## 部署配置

### Docker 环境变量

如果使用 Docker 部署，在 `docker-compose.yml` 中添加环境变量：

```yaml
services:
  cloud-ide-web:
    environment:
      - LINUXDO_CLIENT_ID=your_client_id_here
      - LINUXDO_CLIENT_SECRET=your_client_secret_here
      - LINUXDO_REDIRECT_URL=https://your-domain.com/auth/oauth/linuxdo/callback
      - LINUXDO_BASE_URL=https://connect.linux.do
```

### Kubernetes ConfigMap

创建 ConfigMap 和 Secret：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: oauth-config
data:
  LINUXDO_BASE_URL: "https://connect.linux.do"
  LINUXDO_REDIRECT_URL: "https://your-domain.com/auth/oauth/linuxdo/callback"

---
apiVersion: v1
kind: Secret
metadata:
  name: oauth-secrets
type: Opaque
stringData:
  LINUXDO_CLIENT_ID: "your_client_id_here"
  LINUXDO_CLIENT_SECRET: "your_client_secret_here"
```

## 验证配置

### 1. 检查 OAuth 状态

访问 `/auth/oauth/status` 端点检查配置状态：

```bash
curl https://your-domain.com/auth/oauth/status
```

预期响应：
```json
{
  "status": 0,
  "data": {
    "linuxdo_enabled": true
  }
}
```

### 2. 测试登录流程

1. 访问登录页面
2. 确认显示 "使用 LinuxDo 登录" 按钮
3. 点击按钮，应该跳转到 LinuxDo 授权页面
4. 授权后应该成功登录并跳转到主页

## API 端点

### OAuth 相关端点

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/auth/oauth/status` | 获取 OAuth 配置状态 |
| GET | `/auth/oauth/linuxdo/login` | 发起 LinuxDo 登录 |
| GET | `/auth/oauth/linuxdo/callback` | LinuxDo 回调处理 |
| POST | `/auth/oauth/linuxdo/callback/api` | API 版本的回调处理 |

### 登录流程

1. **发起登录**: `GET /auth/oauth/linuxdo/login`
   - 返回授权 URL 和 state 参数
   - 前端跳转到授权 URL

2. **授权回调**: `GET /auth/oauth/linuxdo/callback?code=xxx&state=xxx`
   - 验证 state 参数
   - 使用 code 交换访问令牌
   - 获取用户信息
   - 登录或创建用户
   - 重定向到成功页面

## 用户关联逻辑

系统会按以下优先级处理用户登录：

1. **LinuxDo ID 匹配**: 如果数据库中存在相同 LinuxDo ID 的用户，直接登录
2. **邮箱匹配**: 如果邮箱匹配现有用户，绑定 LinuxDo 账号并登录
3. **创建新用户**: 如果都不匹配，创建新用户账号

## 安全注意事项

### 1. 保护敏感信息
- 永远不要在前端代码中暴露 Client Secret
- 使用环境变量或安全的配置管理工具存储凭据
- 定期轮换 Client Secret

### 2. 验证回调地址
- 确保回调地址使用 HTTPS
- 在 LinuxDo 应用配置中精确匹配回调地址
- 验证 state 参数防止 CSRF 攻击

### 3. 错误处理
- 不要在错误信息中暴露敏感信息
- 记录详细的错误日志用于调试
- 为用户提供友好的错误提示

## 故障排除

### 常见问题

1. **OAuth 按钮不显示**
   - 检查 `/auth/oauth/status` 返回的状态
   - 确认 Client ID 和 Client Secret 已正确配置

2. **授权后跳转失败**
   - 检查回调地址配置是否正确
   - 确认 LinuxDo 应用中的回调地址设置

3. **Token 交换失败**
   - 验证 Client Secret 是否正确
   - 检查网络连接到 LinuxDo 服务

4. **用户信息获取失败**
   - 确认授权范围包含用户信息读取权限
   - 检查 LinuxDo API 的可用性

### 调试日志

启用详细日志记录：

```yaml
logger:
  level: "DEBUG"
  toFile: true
```

关键日志关键词：
- `OAuth login successful`
- `Failed to exchange OAuth token`
- `Failed to get OAuth user info`

## 更新和维护

### 版本兼容性
- 当前实现基于 LinuxDo OAuth 2.0 标准
- 如果 LinuxDo API 发生变化，需要相应更新代码

### 监控建议
- 监控 OAuth 登录成功率
- 跟踪 Token 交换失败率
- 监控用户信息获取延迟

## 支持

如果遇到配置问题或需要技术支持，请：

1. 检查本文档的故障排除部分
2. 查看应用日志获取详细错误信息
3. 确认 LinuxDo 平台的服务状态
4. 联系开发团队获取支持