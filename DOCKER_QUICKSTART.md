# 🎉 Docker Compose 一键部署 - 立即体验！

## 🚀 3 步快速启动

### 步骤 1: 启动所有依赖服务（30秒）

```bash
cd /Users/bobbowu/Happy/happy-billing

# 一键启动
./scripts/start-docker.sh

# 或手动执行
docker-compose up -d
```

**启动的服务：**
- ✅ MySQL 8.0 (端口 3306)
- ✅ Redis 7.x (端口 6379)
- ✅ ClickHouse 23.x (端口 8123, 9000)
- ✅ Jaeger (端口 16686, 14268)

### 步骤 2: 启动 API 服务（5秒）

```bash
# 已编译好的二进制文件
./bin/api

# 或使用 go run
go run cmd/api/main.go
```

**预期输出：**
```json
{"level":"info","msg":"MySQL connected successfully"}
{"level":"info","msg":"Redis connected successfully"}
{"level":"info","msg":"ClickHouse connected successfully"}
{"level":"info","msg":"Tracing enabled"}
{"level":"info","msg":"HTTP Server listening on 0.0.0.0:8080"}
```

### 步骤 3: 验证服务

```bash
# 方式一：使用测试脚本
./scripts/test-services.sh

# 方式二：手动测试
curl http://localhost:8080/ping
curl http://localhost:8080/health | jq .
```

---

## ✅ 验证清单

打开新的终端窗口，执行以下命令验证：

### 1. 检查 Docker 服务

```bash
docker-compose ps
```

**预期输出：**
```
NAME                        STATUS              PORTS
happy-billing-mysql         Up (healthy)        0.0.0.0:3306->3306/tcp
happy-billing-redis         Up (healthy)        0.0.0.0:6379->6379/tcp
happy-billing-clickhouse    Up (healthy)        0.0.0.0:8123->8123/tcp
happy-billing-jaeger        Up (healthy)        0.0.0.0:16686->16686/tcp
```

### 2. 测试 API 健康检查

```bash
curl http://localhost:8080/health | jq .
```

**预期响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok",
    "timestamp": "2026-01-17T...",
    "services": {
      "clickhouse": {"status": "ok"},
      "mysql": {"status": "ok"},
      "redis": {"status": "ok"}
    }
  }
}
```

### 3. 体验链路追踪

```bash
# 发送带业务上下文的请求
curl -H "X-Tenant-ID: T20240117001" \
     -H "X-User-ID: U20240117001" \
     -H "X-Org-ID: ORG20240117001" \
     http://localhost:8080/health | jq .

# 查看响应头中的 Trace ID
# X-Trace-ID: ...
# X-Span-ID: ...
# X-Request-ID: ...
```

### 4. 在 Jaeger UI 中查看追踪

```bash
# 打开 Jaeger UI
open http://localhost:16686

# 在 UI 中:
# 1. Service: 选择 "happy-billing-api"
# 2. Tags: 输入 "tenant.id=T20240117001"
# 3. 点击 "Find Traces"
# 4. 查看完整的请求链路（HTTP → MySQL → Redis）
```

---

## 📊 访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| API 服务 | http://localhost:8080 | 健康检查、Ping |
| Jaeger UI | http://localhost:16686 | 链路追踪查询 |
| MySQL | localhost:3306 | 业务数据库 |
| Redis | localhost:6379 | 缓存 |
| ClickHouse | localhost:8123 | 计量数据库 |

**数据库凭据：**
- 用户名: `billing_user`
- 密码: `billing_pass_2024`
- 数据库: `happy_billing`

---

## 🎯 体验功能

### 1. 基础功能测试

```bash
# Ping 测试
curl http://localhost:8080/ping

# 健康检查
curl http://localhost:8080/health
```

### 2. 业务上下文染色测试

```bash
# 发送带租户ID和用户ID的请求
curl -v \
  -H "X-Tenant-ID: T20240117001" \
  -H "X-User-ID: U20240117001" \
  -H "X-Org-ID: ORG20240117001" \
  -H "X-Project-ID: PRJ20240117001" \
  http://localhost:8080/health

# 注意响应头中的追踪信息:
# X-Trace-ID: ...
# X-Span-ID: ...
# X-Request-ID: ...
```

### 3. 链路追踪可视化

1. 访问 Jaeger UI: http://localhost:16686
2. 在 "Service" 下拉框选择 `happy-billing-api`
3. 在 "Tags" 输入框输入 `tenant.id=T20240117001`
4. 点击 "Find Traces" 按钮
5. 点击任意 Trace 查看详细信息

**可以看到：**
- HTTP Span (包含租户ID、用户ID等业务上下文)
- MySQL Span (数据库查询，继承业务上下文)
- Redis Span (缓存操作，继承业务上下文)

### 4. 日志追踪关联

```bash
# 查看 API 服务日志（包含 trace_id 和业务上下文）
# 在运行 API 的终端中可以看到类似输出:
{
  "level": "info",
  "time": "...",
  "msg": "...",
  "trace_id": "abc123...",
  "tenant_id": "T20240117001",
  "user_id": "U20240117001"
}
```

---

## 🛠️ 常用命令

### 服务管理

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose stop

# 重启服务
docker-compose restart

# 查看日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f mysql
docker-compose logs -f redis

# 停止并删除服务（保留数据）
docker-compose down

# 停止并删除服务和数据（危险！）
docker-compose down -v
```

### 数据库管理

```bash
# MySQL 命令行
docker exec -it happy-billing-mysql \
  mysql -ubilling_user -pbilling_pass_2024 happy_billing

# Redis 命令行
docker exec -it happy-billing-redis redis-cli

# ClickHouse 客户端
docker exec -it happy-billing-clickhouse clickhouse-client
```

---

## 🐛 故障排查

### 问题：Docker 服务启动失败

```bash
# 查看日志
docker-compose logs

# 重新启动
docker-compose down
docker-compose up -d
```

### 问题：API 连接数据库失败

```bash
# 检查 Docker 服务状态
docker-compose ps

# 确认所有服务都是 "Up (healthy)"

# 检查配置文件
cat config/config.yaml

# 确认数据库连接信息正确
```

### 问题：Jaeger 看不到追踪数据

```bash
# 1. 确认追踪已启用
grep "tracing" config/config.yaml

# 2. 确认 API 日志显示 "Tracing enabled"

# 3. 发送测试请求
curl http://localhost:8080/health

# 4. 刷新 Jaeger UI
```

---

## 📚 完整文档

- **快速启动**: 当前文件
- **详细部署指南**: [docker-compose-guide.md](./docker-compose-guide.md)
- **部署完成报告**: [DOCKER_COMPLETE.md](./DOCKER_COMPLETE.md)
- **链路追踪使用**: [../development/tracing-implementation.md](../development/tracing-implementation.md)

---

## 🎓 已实现的功能

### 基础设施层 ✅
- ✅ 配置管理 (Viper)
- ✅ 日志系统 (Zap)
- ✅ 错误处理 (统一错误码)
- ✅ 数据库连接池 (MySQL + ClickHouse + Redis)
- ✅ 工具函数 (ID生成、时间、金额)
- ✅ 基础模型定义

### 链路追踪 ✅
- ✅ OpenTelemetry + Jaeger 集成
- ✅ HTTP 请求自动追踪
- ✅ 数据库操作追踪 (GORM)
- ✅ Redis 操作追踪
- ✅ 业务上下文染色 (租户ID、用户ID等)
- ✅ 日志与追踪关联

### API 服务 ✅
- ✅ HTTP 服务器 (Gin)
- ✅ 健康检查接口
- ✅ 追踪中间件
- ✅ CORS 中间件
- ✅ 优雅启停

---

## 🎉 恭喜！

你已经成功部署了 Happy Billing 的完整开发环境！

**下一步：**
1. 体验链路追踪功能
2. 查看日志中的 trace_id 关联
3. 开始租户模块开发

**遇到问题？**
- 查看 [故障排查](#-故障排查)
- 阅读 [完整文档](#-完整文档)
- 提交 [GitHub Issue](https://github.com/wwnj/happy-billing/issues)

Happy Coding! 🚀
