# Happy Billing 快速启动指南

## 🚀 一键启动（推荐）

### 全新环境启动

```bash
cd /Users/wb/Happy/happy-billing
bash scripts/start-docker.sh
```

该脚本会自动：
- ✅ 检查 Docker 环境
- ✅ 启动所有基础服务（MySQL, Redis, ClickHouse, Kafka, Jaeger, Loki, Grafana）
- ✅ 自动检测并初始化数据库
- ✅ 加载完整测试数据
- ✅ 显示访问地址和后续步骤

### 预期输出

```
=========================================
  Happy Billing 一键启动
=========================================

✅ Docker 环境检查通过
✅ .env 文件已存在

🚀 启动 Docker 服务...
⏳ 等待服务就绪...
📊 检查服务状态...

=========================================
  数据库初始化
=========================================

📋 检测到空数据库，将执行初始化...
🔧 执行数据库初始化脚本...

1️⃣  创建租户模块表...
2️⃣  创建产品模块表...
3️⃣  创建定价模块表...
...
✅ 数据库初始化成功！

📊 数据统计:
   - 数据表数量: 18
   - 测试租户数: 3
   - 测试产品数: 5
   - 测试订单数: 9
```

## 📊 验证数据库状态

```bash
cd /Users/wb/Happy/happy-billing
bash migrations/mysql/check.sh
```

输出示例：
```
✅ 数据库: happy_billing
📊 数据表数量: 18

📋 模块表统计:
   租户模块: 3 个租户、3 个组织、3 个项目、3 个用户
   产品模块: 8 个分类、2 个SPU、5 个SKU
   定价模块: 4 条价格规则、2 条折扣规则
   订单模块: 9 个订单、8 个订单项、7 张账单
   ...
```

## 🔧 手动初始化数据库

### 场景 1：只初始化数据库（不启动全部服务）

```bash
# 确保 MySQL 容器正在运行
docker compose up -d mysql

# 执行初始化
cd /Users/wb/Happy/happy-billing
bash migrations/mysql/init.sh
```

### 场景 2：重置数据库（清空并重新初始化）

```bash
# 方法一：使用初始化脚本（会提示确认）
bash migrations/mysql/init.sh

# 方法二：手动清空并重新初始化
docker compose exec -T mysql mysql -ubilling_user -pbilling_pass_2024 -e "DROP DATABASE IF EXISTS happy_billing; CREATE DATABASE happy_billing;"
bash migrations/mysql/init.sh
```

## 📍 访问地址

启动成功后，可以通过以下地址访问各个服务：

| 服务 | 地址 | 说明 |
|------|------|------|
| **前端 UI** | http://localhost:5173 | React 前端界面 |
| **API 服务** | http://localhost:8080 | 后端 API（需手动启动）|
| **Jaeger UI** | http://localhost:16686 | 分布式追踪界面 |
| **Grafana** | http://localhost:3000 | 日志/监控界面（admin/admin）|
| **MySQL** | localhost:3306 | 数据库（billing_user/billing_pass_2024）|
| **Redis** | localhost:6379 | 缓存服务 |

## 🎯 启动后端 API

```bash
# 方式一：使用 Makefile
cd /Users/wb/Happy/happy-billing
make run-api

# 方式二：直接运行
cd /Users/wb/Happy/happy-billing
go run cmd/api/main.go
```

## 🎨 启动前端开发服务

```bash
cd /Users/wb/Happy/happy-billing-frontend
npm run dev
```

访问 http://localhost:5173 查看前端界面

## 🧪 测试 API

### 健康检查

```bash
curl http://localhost:8080/health
```

### 查询租户列表

```bash
curl http://localhost:8080/api/v1/tenants?page=1&page_size=10
```

### 查询订单列表

```bash
curl "http://localhost:8080/api/v1/orders?tenant_id=tenant_demo_001&page=1&page_size=5"
```

## 🔍 查看服务日志

```bash
# 查看所有服务日志
docker compose logs -f

# 查看特定服务日志
docker compose logs -f mysql
docker compose logs -f redis
docker compose logs -f jaeger
docker compose logs -f loki
docker compose logs -f grafana
```

## 🛑 停止服务

```bash
# 停止所有服务（保留数据）
docker compose stop

# 停止并删除容器（保留数据卷）
docker compose down

# 完全清理（包括数据卷）
docker compose down -v
```

## 📦 测试数据说明

### 租户数据（3 个）

| tenant_id | 名称 | 类型 | 状态 |
|-----------|------|------|------|
| tenant_demo_001 | Demo 演示租户 | ENTERPRISE | 已认证 |
| tenant_a3f9b2c4d5 | 测试个人开发者 | INDIVIDUAL | 已认证 |
| tenant_e8d7c2a1f4 | 测试企业用户 | ENTERPRISE | 未认证 |

### 产品数据（5 个 SKU）

- **GPU 计算**：
  - A100 40GB（北京/上海）
  - A100 80GB（北京/上海）
  - V100 32GB（北京）

### 订单数据（9 个）

- 涵盖不同状态：待支付（PENDING）、已支付（PAID）、已取消（CANCELED）
- 关联测试租户 `tenant_demo_001`

### 测试用户凭证

| 用户名 | 密码 | 租户 | 角色 |
|--------|------|------|------|
| demouser | 123456 | tenant_demo_001 | 主账号 |
| testuser | 123456 | tenant_a3f9b2c4d5 | 主账号 |
| admin | 123456 | tenant_e8d7c2a1f4 | 主账号 |

⚠️ **安全提示**：测试密码仅用于开发环境，生产环境请务必修改！

## 🐛 故障排查

### 问题 1: Docker 服务无法启动

```bash
# 检查 Docker 状态
docker info

# 重启 Docker Desktop
# macOS: 点击 Docker 图标 -> Restart
```

### 问题 2: 端口已被占用

```bash
# 检查端口占用
lsof -i :3306  # MySQL
lsof -i :8080  # API
lsof -i :5173  # Frontend

# 停止占用端口的进程
kill -9 <PID>
```

### 问题 3: 数据库初始化失败

```bash
# 查看 MySQL 日志
docker compose logs mysql

# 手动连接数据库检查
docker compose exec mysql mysql -ubilling_user -pbilling_pass_2024 -e "SHOW DATABASES;"
```

### 问题 4: 服务健康检查失败

```bash
# 检查服务状态
docker compose ps

# 查看不健康服务的日志
docker compose logs <service_name>

# 重启特定服务
docker compose restart <service_name>
```

## 📚 相关文档

- [数据库迁移说明](migrations/README.md)
- [架构设计文档](docs/architecture.md)
- [API 接口文档](docs/api.md)
- [开发指南](docs/development.md)

## 🆘 获取帮助

如遇问题，请：
1. 查看本文档的故障排查部分
2. 检查服务日志：`docker compose logs -f`
3. 查看数据库状态：`bash migrations/mysql/check.sh`
4. 提交 Issue 到项目仓库

---

**Happy Coding! 🎉**
