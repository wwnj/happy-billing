# OpenTelemetry 链路追踪 + 业务上下文染色实现报告

## ✅ 已完成的功能

### 1. OpenTelemetry 核心追踪模块 (`pkg/tracing/`)

**文件：** `pkg/tracing/tracing.go`

**功能：**
- ✅ 基于 OpenTelemetry SDK 的追踪系统
- ✅ Jaeger Exporter 集成
- ✅ 灵活的采样率配置 (0.0-1.0)
- ✅ Trace Context 跨服务传播
- ✅ 优雅的初始化和关闭

**核心API：**
```go
// 初始化追踪系统
tracing.Init(&tracing.Config{
    Enabled:     true,
    ServiceName: "happy-billing-api",
    Endpoint:    "http://localhost:14268/api/traces",
    SampleRate:  1.0,
})

// 创建 Span
ctx, span := tracing.StartSpan(ctx, "operation-name")
defer span.End()

// 获取 Trace ID
traceID := tracing.GetTraceID(ctx)
```

---

### 2. 业务上下文染色 (`pkg/tracing/context.go`)

**支持的业务上下文：**
- ✅ **租户ID** (Tenant ID)
- ✅ **用户ID** (User ID)
- ✅ **组织ID** (Organization ID)
- ✅ **项目ID** (Project ID)
- ✅ **订单ID** (Order ID)
- ✅ **账单ID** (Bill ID)
- ✅ **请求ID** (Request ID)

**HTTP Header 映射：**
```
X-Tenant-ID   → tenant_id
X-User-ID     → user_id
X-Org-ID      → org_id
X-Project-ID  → project_id
X-Request-ID  → request_id
```

**Span 属性映射：**
```
tenant.id   → 租户ID
user.id     → 用户ID
org.id      → 组织ID
project.id  → 项目ID
order.id    → 订单ID
bill.id     → 账单ID
request.id  → 请求ID
```

**核心API：**
```go
// 设置租户ID到上下文和Span
ctx = tracing.SetTenantID(ctx, "T20240117001")

// 获取租户ID
tenantID := tracing.GetTenantID(ctx)

// 批量设置业务上下文
ctx = tracing.SetBusinessContext(ctx, &tracing.BusinessContext{
    TenantID:  "T20240117001",
    UserID:    "U20240117001",
    ProjectID: "PRJ20240117001",
})
```

---

### 3. Gin HTTP 追踪中间件 (`internal/api/middleware/tracing.go`)

**自动追踪功能：**
- ✅ 自动创建 HTTP Span
- ✅ 从 Header 提取业务上下文并染色
- ✅ 生成或提取 Request ID
- ✅ 记录 HTTP 元信息（Method、Route、StatusCode、UserAgent、ClientIP）
- ✅ 自动注入 Trace ID 和 Span ID 到响应头
- ✅ 错误自动标记和记录

**响应头：**
```
X-Trace-ID    → Trace ID (用于日志关联)
X-Span-ID     → Span ID
X-Request-ID  → Request ID (自动生成或来自请求头)
```

**使用方式：**
```go
r.Use(middleware.Tracing())
```

---

### 4. GORM 数据库追踪插件 (`pkg/tracing/gorm.go`)

**追踪能力：**
- ✅ 自动追踪所有数据库操作 (Create/Query/Update/Delete/Raw)
- ✅ 记录 SQL 语句
- ✅ 记录表名
- ✅ 记录影响行数
- ✅ 继承业务上下文（租户ID、用户ID等）
- ✅ 错误自动记录

**Span 属性：**
```
db.system      → mysql
db.operation   → query/create/update/delete
db.sql         → SQL语句
db.sql.table   → 表名
db.rows_affected → 影响行数
tenant.id      → 租户ID (继承)
user.id        → 用户ID (继承)
```

**使用方式：**
```go
// 自动安装（在 InitMySQL 中）
tracingPlugin := tracing.NewGormTracingPlugin(tracing.DBSystemMySQL)
db.Use(tracingPlugin)

// 使用时传递上下文
db.WithContext(ctx).Where("id = ?", 1).First(&user)
```

---

### 5. Redis 追踪 Hook (`pkg/tracing/redis.go`)

**追踪能力：**
- ✅ 自动追踪所有 Redis 命令
- ✅ 记录命令名称和参数
- ✅ 支持 Pipeline 批量操作追踪
- ✅ 继承业务上下文
- ✅ 错误自动记录

**Span 属性：**
```
db.system         → redis
db.operation      → GET/SET/HGET等
db.statement      → 完整命令（参数限长100字符）
net.peer.name     → Redis地址
db.pipeline.size  → Pipeline命令数量（Pipeline模式）
tenant.id         → 租户ID (继承)
```

**使用方式：**
```go
// 自动安装（在 InitRedis 中）
tracing.InstallRedisHook(client, "127.0.0.1:6379")

// Redis 操作会自动追踪
client.Get(ctx, "key")
```

---

### 6. 日志与追踪关联 (`pkg/logger/context.go`)

**自动注入字段：**
- ✅ **trace_id** - Trace ID
- ✅ **span_id** - Span ID
- ✅ **tenant_id** - 租户ID
- ✅ **user_id** - 用户ID
- ✅ **org_id** - 组织ID
- ✅ **project_id** - 项目ID
- ✅ **order_id** - 订单ID
- ✅ **bill_id** - 账单ID
- ✅ **request_id** - 请求ID

**使用方式：**
```go
// 自动包含 Trace ID 和业务上下文
logger.InfoCtx(ctx, "处理订单", zap.String("order_id", "ORD20240117001"))

// 日志输出示例：
{
  "level": "info",
  "time": "2026-01-17T16:30:00+08:00",
  "msg": "处理订单",
  "trace_id": "a1b2c3d4e5f6g7h8",
  "span_id": "i9j0k1l2m3n4",
  "tenant_id": "T20240117001",
  "user_id": "U20240117001",
  "order_id": "ORD20240117001",
  "request_id": "req-uuid-12345"
}
```

---

### 7. 配置支持 (`pkg/config/config.go`, `config/config.example.yaml`)

**配置项：**
```yaml
tracing:
  enabled: true                                         # 是否启用追踪
  service_name: happy-billing-api                       # 服务名称
  endpoint: http://localhost:14268/api/traces           # Jaeger Collector 端点
  sample_rate: 1.0                                      # 采样率 (0.0-1.0)
```

**默认值：**
- Service Name: `happy-billing-api`
- Endpoint: `http://localhost:14268/api/traces`
- Sample Rate: `1.0` (100% 采样)

---

## 🎯 完整的追踪链路

### HTTP 请求追踪示例

```
HTTP Request
  ↓
[Tracing Middleware] - 创建 HTTP Span
  ├─ 提取 Header: X-Tenant-ID, X-User-ID, X-Org-ID
  ├─ 染色到 Span Attributes: tenant.id, user.id, org.id
  ├─ 染色到 Context
  ├─ 生成 Request ID
  └─ 注入响应头: X-Trace-ID, X-Span-ID, X-Request-ID
  ↓
[Business Handler]
  ↓
[Database Operation] - 创建 DB Span (GORM Plugin)
  ├─ 记录 SQL: SELECT * FROM users WHERE tenant_id = ?
  ├─ 继承业务上下文: tenant.id, user.id
  └─ 父 Span: HTTP Span
  ↓
[Redis Operation] - 创建 Redis Span (Redis Hook)
  ├─ 记录命令: GET user:cache:T20240117001:U20240117001
  ├─ 继承业务上下文: tenant.id, user.id
  └─ 父 Span: HTTP Span
  ↓
[Logger] - 日志自动关联
  ├─ trace_id: a1b2c3d4e5f6g7h8
  ├─ span_id: i9j0k1l2m3n4
  ├─ tenant_id: T20240117001
  ├─ user_id: U20240117001
  └─ request_id: req-uuid-12345
```

---

## 📊 Jaeger UI 查看效果

### Span 属性示例

**HTTP Span:**
```
span.name: GET /api/v1/orders
http.method: GET
http.route: /api/v1/orders
http.status_code: 200
http.user_agent: Mozilla/5.0 ...
http.client_ip: 192.168.1.100
tenant.id: T20240117001
user.id: U20240117001
org.id: ORG20240117001
project.id: PRJ20240117001
request.id: req-uuid-12345
```

**Database Span:**
```
span.name: gorm:query
db.system: mysql
db.operation: query
db.sql: SELECT * FROM orders WHERE tenant_id = ? AND id = ?
db.sql.table: orders
db.rows_affected: 1
tenant.id: T20240117001
user.id: U20240117001
```

**Redis Span:**
```
span.name: redis:GET
db.system: redis
db.operation: GET
db.statement: GET order:cache:ORD20240117001
net.peer.name: 127.0.0.1:6379
tenant.id: T20240117001
user.id: U20240117001
```

---

## 🚀 使用指南

### 1. 启动 Jaeger (本地测试)

```bash
# 使用 Docker 启动 Jaeger All-in-One
docker run -d --name jaeger \
  -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 14250:14250 \
  -p 9411:9411 \
  jaegertracing/all-in-one:latest

# 访问 Jaeger UI
open http://localhost:16686
```

### 2. 启动 API 服务

```bash
# 确保配置文件中启用了追踪
vi config/config.yaml

# 启动服务
./bin/api
# 或
go run cmd/api/main.go
```

### 3. 发送测试请求

```bash
# 带业务上下文的请求
curl -H "X-Tenant-ID: T20240117001" \
     -H "X-User-ID: U20240117001" \
     -H "X-Org-ID: ORG20240117001" \
     http://localhost:8080/health

# 查看响应头中的 Trace ID
# X-Trace-ID: a1b2c3d4e5f6g7h8
# X-Span-ID: i9j0k1l2m3n4
# X-Request-ID: req-uuid-12345
```

### 4. 在 Jaeger UI 中查看

1. 打开 `http://localhost:16686`
2. Service 选择: `happy-billing-api`
3. 点击 "Find Traces"
4. 查看 Trace 详情，可以看到：
   - HTTP Span (带业务上下文)
   - MySQL Span (如果有数据库操作)
   - Redis Span (如果有缓存操作)
5. 在 Tags 中搜索: `tenant.id=T20240117001`

---

## 💡 业务场景示例

### 场景1：订单创建追踪

```go
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    // 1. 从上下文获取租户ID
    tenantID := tracing.GetTenantID(ctx)
    userID := tracing.GetUserID(ctx)

    // 2. 创建订单（自动追踪数据库操作）
    order := &Order{
        TenantID: tenantID,
        UserID:   userID,
        ...
    }
    if err := database.GetMySQL().WithContext(ctx).Create(order).Error; err != nil {
        return nil, err
    }

    // 3. 设置订单ID到上下文（染色）
    ctx = tracing.SetOrderID(ctx, order.OrderID)

    // 4. 记录日志（自动包含 trace_id, tenant_id, user_id, order_id）
    logger.InfoCtx(ctx, "订单创建成功",
        zap.String("order_id", order.OrderID),
        zap.Int64("amount", order.Amount))

    // 5. 缓存订单（自动追踪 Redis 操作）
    cacheKey := fmt.Sprintf("order:%s", order.OrderID)
    database.GetRedis().Set(ctx, cacheKey, order, 1*time.Hour)

    return order, nil
}
```

**Jaeger 中可以看到：**
- HTTP Span: `POST /api/v1/orders`
  - tenant.id: T20240117001
  - user.id: U20240117001
- DB Span: `gorm:create`
  - db.sql: `INSERT INTO orders (...)`
  - tenant.id: T20240117001 (继承)
  - user.id: U20240117001 (继承)
- Redis Span: `redis:SET`
  - db.statement: `SET order:ORD20240117001 ...`
  - tenant.id: T20240117001 (继承)
  - order.id: ORD20240117001 (新增)

**日志输出：**
```json
{
  "level": "info",
  "time": "2026-01-17T16:30:00+08:00",
  "msg": "订单创建成功",
  "trace_id": "a1b2c3d4e5f6g7h8",
  "span_id": "i9j0k1l2m3n4",
  "tenant_id": "T20240117001",
  "user_id": "U20240117001",
  "order_id": "ORD20240117001",
  "request_id": "req-uuid-12345",
  "amount": 99900
}
```

---

## 📈 性能影响

### 采样率建议

- **开发环境**: 1.0 (100% 采样)
- **测试环境**: 1.0 (100% 采样)
- **生产环境**: 0.1 - 0.01 (10% - 1% 采样)

### 开销评估

- **Span 创建**: ~1-5μs
- **Attribute 设置**: ~0.5μs/attribute
- **网络传输**: 异步批量发送，不阻塞请求

---

## ✅ 验收标准

全部完成：
- ✅ 编译通过
- ✅ 服务能正常启动
- ✅ HTTP 请求自动创建 Span
- ✅ 业务上下文自动染色
- ✅ 数据库操作自动追踪
- ✅ Redis 操作自动追踪
- ✅ 日志自动关联 Trace ID 和业务上下文
- ✅ Jaeger UI 能正常查看 Trace

---

## 📁 新增文件清单

```
pkg/tracing/
  ├── tracing.go         # OpenTelemetry 核心模块
  ├── context.go         # 业务上下文染色
  ├── gorm.go            # GORM 追踪插件
  └── redis.go           # Redis 追踪 Hook

internal/api/middleware/
  └── tracing.go         # Gin HTTP 追踪中间件

pkg/logger/
  └── context.go         # 日志上下文增强

pkg/config/config.go     # 添加 TracingConfig
config/config.example.yaml  # 添加 tracing 配置
```

---

## 🎓 总结

OpenTelemetry 链路追踪 + 业务上下文染色已完整实现！

**核心价值：**
1. **全链路可观测** - HTTP → Database → Redis 全程追踪
2. **业务上下文关联** - 租户ID、用户ID等自动传播
3. **日志追踪关联** - Trace ID 自动注入日志
4. **零侵入集成** - 中间件和插件自动处理
5. **灵活配置** - 支持启用/禁用和采样率控制

**下一步建议：**
- 在 Kubernetes 中部署 Jaeger
- 配置长期存储（Elasticsearch/Cassandra）
- 设置告警规则（慢查询、高错误率）
- 接入 Grafana 进行可视化分析

准备好使用链路追踪进行问题排查了！🎉
