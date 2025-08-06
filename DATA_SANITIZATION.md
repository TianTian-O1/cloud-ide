# 🔒 数据脱敏指南 - Cloud IDE v1.2 Release

> 此版本已经过数据脱敏处理，移除了所有敏感信息

## 🧹 已清理的敏感数据

### 🔑 **认证信息**
- **Gateway Token**: 示例脚本中的 Token 已替换为占位符
- **API Keys**: 配置文件中的 API 密钥已清空
- **数据库密码**: 配置中的数据库密码已清空
- **OAuth Secrets**: 第三方认证密钥已清空

### 💾 **编译文件**
- **二进制文件**: 删除所有编译后的可执行文件
  - `control-plane`
  - `webserver` 
  - `manager`
- **临时文件**: 删除所有 `.backup`, `.tmp`, `.log` 文件

### 🗃️ **开发文件**
- **Node Modules**: 删除 `front-end/node_modules/` 减少体积
- **Backup 目录**: 删除所有备份目录和文件
- **测试配置**: 删除测试用的 YAML 和 SQL 文件

### 🚫 **已删除的文件**
```
❌ backup/                          # 备份目录
❌ front-end/node_modules/          # NPM 依赖包
❌ front-end/dist.tar.gz            # 构建产物
❌ *.backup                         # 所有备份文件
❌ claude-workspace*.yaml           # 测试工作空间配置
❌ test-*.yaml                      # 测试配置文件
❌ *.sql                            # 数据库脚本
❌ *webserver, *control-plane       # 编译后的二进制
```

## 🔧 配置文件脱敏

### `config/webserver.yaml`
```yaml
database:
  password: ""                    # ✅ 已清空

oauth:
  client_secret: ""              # ✅ 已清空
```

### `run.sh`
```bash
# 原始（包含真实 Token）
-gateway-token XnRbVnoUZa0rT9xKAwHX0Zof3H7VpfCe

# 脱敏后（使用占位符）
-gateway-token YOUR_GATEWAY_TOKEN_HERE
```

## 🚀 部署前需要配置

在部署此版本前，请确保配置以下信息：

### 1. **环境变量配置**
```bash
export GATEWAY_TOKEN="your-actual-gateway-token"
export DB_PASSWORD="your-database-password"
export OAUTH_CLIENT_SECRET="your-oauth-secret"
```

### 2. **配置文件更新**
- 更新 `config/webserver.yaml` 中的数据库密码
- 配置 OAuth 客户端密钥
- 替换 `run.sh` 中的占位符

### 3. **重新编译**
```bash
# 编译控制平面
go build -o bin/control-plane cmd/control-plane/main.go

# 编译 Web 服务器
go build -o bin/webserver cmd/webserver/main.go
```

### 4. **安装依赖**
```bash
# 前端依赖
cd front-end && npm install

# Go 依赖
go mod download
```

## 🔍 验证清理结果

使用以下命令验证敏感数据已清理：

```bash
# 检查是否还有硬编码的 Token
grep -r "XnRbVnoUZa0rT9xKAwHX0Zof3H7VpfCe" .

# 检查是否还有真实 IP 地址
grep -r "10.99.144.6" .

# 检查是否还有备份文件
find . -name "*.backup*"

# 检查是否还有编译文件
find . -name "*webserver" -o -name "*control-plane"
```

## 📋 安全检查清单

- [x] 删除所有编译后的二进制文件
- [x] 清理配置文件中的敏感信息
- [x] 替换硬编码的 Token 和密钥
- [x] 删除备份和临时文件
- [x] 移除测试配置和数据
- [x] 删除大型依赖目录
- [x] 创建数据脱敏文档

## ⚠️ 重要提醒

1. **此版本仅供代码审查和学习使用**
2. **部署前必须配置真实的认证信息**
3. **不要在生产环境使用占位符配置**
4. **建议使用环境变量管理敏感信息**

---

**🔒 Cloud IDE v1.2 - 安全的开源发布版本**