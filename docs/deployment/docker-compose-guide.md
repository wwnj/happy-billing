# 🐳 Docker Compose 一键部署指南

## 📋 快速开始

### 1. 前置条件

确保已安装：
- Docker (>= 20.x)
- Docker Compose (>= 2.x)

```bash
# 检查版本
docker --version
docker-compose --version
```

---

## 🚀 快速启动（3 步）

### 步骤 1: 启动所有基础服务

```bash
# 在项目根目录执行
docker-compose up -d

# 查看服务状态
docker-compose ps
```

**启动的服务：**
- ✅ MySQL 8.0 (端口 3306)
- ✅ Redis 7.x (端口 6379)
- ✅ ClickHouse 23.x (端口 8123, 9000)
- ✅ Jaeger (端口 16686)

### 步骤 2: 等待服务就绪

```bash
# 查看服务日志
docker-compose logs -f

# 等待所有服务健康检查通过（约30秒）
# 看到类似输出表示就绪：
# mysql        | ready for connections
# redis        | Ready to accept connections
# clickhouse   | Ready for connections
# jaeger       | Server listening
```

### 步骤 3: 启动 API 服务

```bash
# 方式一：直接运行
./bin/api

# 方式二：使用 go run
go run cmd/api/main.go

# 方式三：使用配置文件
./bin/api --config=config/config.docker.yaml
```

---

## ✅ 验证服务

### 1. 验证基础服务

```bash
# MySQL
docker exec -it happy-billing-mysql mysql -ubilling_user -pbilling_pass_2024 -e "SELECT 'MySQL OK' as status;"

# Redis
docker exec -it happy-billing-redis redis-cli ping

# ClickHouse
curl http://localhost:8123/ping

# Jaeger UI
open http://localhost:16686
```

### 2. 验证 API 服务

```bash
# Ping 检查
curl http://localhost:8080/ping

# 健康检查
curl http://localhost:8080/health

# 带业务上下文的请求
curl -H "X-Tenant-ID: T20240117001" \
     -H "X-User-ID: U20240117001" \
     http://localhost:8080/health
```

**预期响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok",
    "timestamp": "2026-01-17T16:30:00+08:00",
    "services": {
      "mysql": {"status": "ok"},
      "redis": {"status": "ok"},
      "clickhouse": {"status": "ok"}
    }
  }
}
```

---

## 🔧 高级用法

### 启动可选服务

#### 1. 启动 Kafka（消息队列）

```bash
docker-compose --profile with-kafka up -d
```

**包含服务：**
- Zookeeper
- Kafka (端口 29092)

#### 2. 启动管理工具

```bash
docker-compose --profile with-admin-tools up -d
```

**包含服务：**
- Adminer (数据库管理) - http://localhost:8080
- RedisInsight (Redis 管理) - http://localhost:8001

#### 3. 启动所有服务

```bash
docker-compose --profile with-kafka --profile with-admin-tools up -d
```

---

## 📊 访问管理界面

| 服务 | 访问地址 | 说明 |
|------|---------|------|
| Jaeger UI | http://localhost:16686 | 链路追踪查询 |
| Adminer | http://localhost:8080 | 数据库管理（需启用 profile） |
| RedisInsight | http://localhost:8001 | Redis 管理（需启用 profile） |

---

## 🛠️ 常用命令

### 服务管理

```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose down

# 重启服务
docker-compose restart

# 查看服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f [service_name]

# 进入容器
docker exec -it happy-billing-mysql bash
```

### 数据管理

```bash
# 清理所有数据（危险操作！）
docker-compose down -v

# 仅停止服务，保留数据
docker-compose stop

# 备份 MySQL 数据
docker exec happy-billing-mysql mysqldump -ubilling_user -pbilling_pass_2024 happy_billing > backup.sql

# 恢复 MySQL 数据
docker exec -i happy-billing-mysql mysql -ubilling_user -pbilling_pass_2024 happy_billing < backup.sql
```

---

## 🔐 安全配置

### 生产环境建议

1. **修改默认密码**

编辑 `.env` 文件：
```bash
cp .env.example .env
vi .env

# 修改以下配置
MYSQL_ROOT_PASSWORD=your_strong_password
MYSQL_PASSWORD=your_strong_password
CLICKHOUSE_PASSWORD=your_strong_password
```

2. **Redis 启用密码**

编辑 `docker/redis/redis.conf`：
```conf
requirepass your_redis_password
```

3. **限制端口暴露**

编辑 `docker-compose.yml`，移除不必要的端口映射。

---

## 🐛 故障排查

### 问题 1: MySQL 连接失败

```bash
# 检查 MySQL 是否启动
docker-compose ps mysql

# 查看 MySQL 日志
docker-compose logs mysql

# 测试连接
docker exec -it happy-billing-mysql mysql -uroot -phappy_billing_2024
```

### 问题 2: Redis 连接超时

```bash
# 检查 Redis 状态
docker exec -it happy-billing-redis redis-cli ping

# 查看 Redis 日志
docker-compose logs redis
```

### 问题 3: ClickHouse 启动失败

```bash
# 检查日志
docker-compose logs clickhouse

# 可能原因：ulimit 限制
# 解决方法：修改 docker-compose.yml 中的 ulimits 配置
```

### 问题 4: 端口冲突

```bash
# 检查端口占用
lsof -i :3306  # MySQL
lsof -i :6379  # Redis
lsof -i :8123  # ClickHouse HTTP
lsof -i :9000  # ClickHouse Native

# 修改端口：编辑 .env 文件
```

---

## 📝 配置文件说明

### 主要配置文件

```
docker-compose.yml           # Docker Compose 编排文件
.env.example                 # 环境变量示例
config/config.docker.yaml    # API 服务配置（Docker 环境）

docker/
├── mysql/
│   ├── init.sql            # MySQL 初始化脚本
│   └── my.cnf              # MySQL 配置
├── clickhouse/
│   └── init.sql            # ClickHouse 初始化脚本
└── redis/
    └── redis.conf          # Redis 配置
```

### 环境变量优先级

1. `.env` 文件（如果存在）
2. `.env.example` 默认值
3. `docker-compose.yml` 中的默认值

---

## 📈 性能调优

### MySQL 调优

编辑 `docker/mysql/my.cnf`：
```ini
innodb_buffer_pool_size=1G      # 根据内存调整
max_connections=500             # 根据需求调整
```

### ClickHouse 调优

```bash
# 增加 ClickHouse 内存限制
# 编辑 docker-compose.yml
clickhouse:
  deploy:
    resources:
      limits:
        memory: 4G
```

### Redis 调优

编辑 `docker/redis/redis.conf`：
```conf
maxmemory 1gb                   # 根据内存调整
maxmemory-policy allkeys-lru
```

---

## 🧪 测试环境

### 快速创建测试数据

```bash
# MySQL
docker exec -it happy-billing-mysql mysql -ubilling_user -pbilling_pass_2024 happy_billing -e "
CREATE TABLE test_table (id INT PRIMARY KEY, name VARCHAR(100));
INSERT INTO test_table VALUES (1, 'test');
SELECT * FROM test_table;
"

# Redis
docker exec -it happy-billing-redis redis-cli SET test_key test_value
docker exec -it happy-billing-redis redis-cli GET test_key

# ClickHouse
curl 'http://localhost:8123/?query=SELECT%201'
```

---

## 🔄 升级服务

```bash
# 拉取最新镜像
docker-compose pull

# 重新创建容器
docker-compose up -d --force-recreate

# 清理旧镜像
docker image prune -f
```

---

## 📦 完整示例：从零启动

```bash
# 1. 克隆项目（或进入项目目录）
cd happy-billing

# 2. 复制环境变量配置
cp .env.example .env

# 3. 启动所有基础服务
docker-compose up -d

# 4. 等待服务就绪（约30秒）
docker-compose logs -f

# 5. 编译 API 服务
go build -o bin/api cmd/api/main.go

# 6. 启动 API 服务
./bin/api

# 7. 验证服务
curl http://localhost:8080/health

# 8. 查看 Jaeger UI
open http://localhost:16686

# 9. 发送测试请求
curl -H "X-Tenant-ID: T20240117001" \
     -H "X-User-ID: U20240117001" \
     http://localhost:8080/health

# 10. 在 Jaeger UI 中搜索 Trace
# Service: happy-billing-api
# Tags: tenant.id=T20240117001
```

---

## ✨ 成功标志

当你看到以下输出，说明一切就绪：

1. **Docker Compose 输出：**
```
✔ Container happy-billing-mysql       Started
✔ Container happy-billing-redis       Started
✔ Container happy-billing-clickhouse  Started
✔ Container happy-billing-jaeger      Started
```

2. **API 服务输出：**
```json
{"level":"info","msg":"MySQL connected successfully"}
{"level":"info","msg":"Redis connected successfully"}
{"level":"info","msg":"ClickHouse connected successfully"}
{"level":"info","msg":"Tracing enabled"}
{"level":"info","msg":"HTTP Server listening on 0.0.0.0:8080"}
```

3. **健康检查响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok",
    "services": {
      "mysql": {"status": "ok"},
      "redis": {"status": "ok"},
      "clickhouse": {"status": "ok"}
    }
  }
}
```

---

## 🎉 开始体验

现在你已经成功部署了 Happy Billing 的所有依赖服务！

**下一步：**
- 查看 [API 文档](../api/)
- 阅读 [开发指南](./code-standards.md)
- 开始 [租户模块开发](./phase2-tenant-module.md)

**遇到问题？**
- 查看 [故障排查](#-故障排查)
- 提交 [Issue](https://github.com/wwnj/happy-billing/issues)

Happy Coding! 🚀
