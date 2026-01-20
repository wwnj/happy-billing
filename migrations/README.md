# Happy Billing 数据库迁移脚本

## 📁 目录结构

```
migrations/
├── mysql/                          # MySQL 数据库相关
│   ├── init.sh                    # 数据库初始化脚本（自动化）
│   ├── init_all.sql               # 所有 SQL 的主入口（可选）
│   ├── 01_schema/                 # DDL 表结构（备用）
│   └── 02_data/                   # DML 测试数据（备用）
├── postgres/                       # PostgreSQL 数据库相关（预留）
├── timeseries/                     # ClickHouse 时序数据库相关（预留）
├── 20260117_create_tenant_tables.sql              # 租户模块表结构
├── 20260117_create_product_tables.sql             # 产品模块表结构
├── 20260117_create_pricing_tables.sql             # 定价模块表结构
├── 20260117_create_order_billing_payment_tables.sql # 订单/账单/支付表结构
├── 20240117_create_exchange_rates.sql             # 汇率表结构
├── 20240117_add_multi_currency_fields.sql         # 多币种字段扩展
├── fix_foreign_keys.sql                           # 外键约束修复
├── add_test_users_credentials.sql                 # 测试用户凭证数据
└── debug_and_test_data.sql                        # 完整测试数据
```

## 🚀 使用方法

### 方法一：使用启动脚本（推荐）

**一键启动所有服务并初始化数据库：**

```bash
cd /Users/wb/Happy/happy-billing
bash scripts/start-docker.sh
```

该脚本会：
1. ✅ 启动所有 Docker 服务（MySQL, ClickHouse, Redis, Kafka, Jaeger, Loki, Grafana）
2. ✅ 自动检测数据库是否需要初始化
3. ✅ 按正确顺序执行 DDL 和 DML 脚本
4. ✅ 显示初始化统计信息

### 方法二：手动执行初始化脚本

**仅初始化/重置数据库：**

```bash
cd /Users/wb/Happy/happy-billing
bash migrations/mysql/init.sh
```

### 方法三：使用 Docker 容器内执行

**通过 Docker 容器直接执行 SQL：**

```bash
# 进入 MySQL 容器
docker compose exec mysql bash

# 执行 SQL 文件
mysql -ubilling_user -pbilling_pass_2024 happy_billing < /path/to/script.sql
```

## 📋 SQL 脚本执行顺序

### DDL（数据定义语言 - 表结构）

执行顺序非常重要，因为存在外键依赖关系：

1. **20260117_create_tenant_tables.sql** - 租户/组织/项目/用户表
   - ⚠️ 基础表，无外键依赖，必须最先创建

2. **20260117_create_product_tables.sql** - 产品分类/SPU/SKU 表
   - 依赖：无

3. **20260117_create_pricing_tables.sql** - 价格规则/折扣规则表
   - 依赖：`product_sku`

4. **20260117_create_order_billing_payment_tables.sql** - 订单/账单/支付/余额表
   - 依赖：`tenants`, `product_sku`

5. **20240117_create_exchange_rates.sql** - 汇率表
   - 依赖：无

6. **20240117_add_multi_currency_fields.sql** - 多币种字段扩展
   - 依赖：已存在的表

7. **fix_foreign_keys.sql** - 外键约束修复
   - 依赖：所有表已创建

### DML（数据操作语言 - 测试数据）

8. **add_test_users_credentials.sql** - 测试用户凭证
   - 依赖：`users` 表已存在

9. **debug_and_test_data.sql** - 完整测试数据集
   - 依赖：所有表已创建
   - 包含：3个测试租户、完整产品目录、定价规则、订单数据

## 🔧 配置说明

### 环境变量

初始化脚本使用以下环境变量（可选，有默认值）：

```bash
MYSQL_HOST=127.0.0.1           # MySQL 主机地址
MYSQL_PORT=3306                 # MySQL 端口
MYSQL_USER=billing_user         # MySQL 用户名
MYSQL_PASSWORD=billing_pass_2024 # MySQL 密码
MYSQL_DATABASE=happy_billing    # 数据库名称
```

### 自定义执行

如果需要自定义执行特定脚本：

```bash
# 只执行 DDL
mysql -h127.0.0.1 -ubilling_user -pbilling_pass_2024 happy_billing < migrations/20260117_create_tenant_tables.sql

# 只插入测试数据
mysql -h127.0.0.1 -ubilling_user -pbilling_pass_2024 happy_billing < migrations/debug_and_test_data.sql
```

## 📊 测试数据说明

执行完成后会自动生成以下测试数据：

### 租户数据
- **3 个测试租户**：
  - `tenant_demo_001` - 个人开发者（已实名认证）
  - `tenant_demo_002` - 初创企业（未认证）
  - `tenant_demo_003` - 大型企业（企业认证）

### 产品数据
- **计算资源类**：GPU（A100/H100）、CPU、内存
- **存储资源类**：对象存储、块存储、NAS
- **网络资源类**：公网带宽、私网带宽

### 定价数据
- 按需定价规则（小时计费）
- 包年包月定价规则（月/年计费）
- 折扣规则（首购/续费/大客户）

### 订单数据
- 8+ 个测试订单，涵盖不同状态（待支付、已支付、已取消）

## ⚠️ 注意事项

1. **数据清空**：重新执行 `init.sh` 会清空现有数据并重新初始化
2. **外键约束**：DDL 脚本按照依赖顺序执行，不要打乱顺序
3. **字符集**：所有表使用 `utf8mb4` 字符集，支持 emoji 和多语言
4. **时区**：时间戳使用系统本地时区
5. **密码安全**：测试数据中的密码 hash 对应明文 `123456`，仅用于开发测试

## 🐛 故障排查

### 问题 1: MySQL 连接失败

```bash
# 检查 MySQL 容器状态
docker compose ps mysql

# 查看 MySQL 日志
docker compose logs mysql

# 测试连接
mysql -h127.0.0.1 -ubilling_user -pbilling_pass_2024 -e "SELECT 1"
```

### 问题 2: 表已存在错误

所有 DDL 脚本都使用 `CREATE TABLE IF NOT EXISTS`，不会因表存在而报错。
如需强制重建，请先删除数据库：

```bash
mysql -h127.0.0.1 -ubilling_user -pbilling_pass_2024 -e "DROP DATABASE IF EXISTS happy_billing; CREATE DATABASE happy_billing;"
```

### 问题 3: 外键约束错误

确保按照正确的顺序执行 DDL 脚本，或临时禁用外键检查：

```sql
SET FOREIGN_KEY_CHECKS = 0;
-- 执行 SQL
SET FOREIGN_KEY_CHECKS = 1;
```

## 📝 维护指南

### 添加新的迁移脚本

1. 在 `migrations/` 目录创建新的 SQL 文件
2. 使用日期前缀命名：`YYYYMMDD_description.sql`
3. 更新 `migrations/mysql/init.sh` 添加执行步骤
4. 更新本 README 文档

### 版本控制

- 所有 SQL 脚本都纳入 Git 版本控制
- 迁移脚本只增不改（除非修复严重 bug）
- 使用新脚本添加字段或表，不要修改旧脚本

## 🔗 相关文档

- [Happy Billing 架构设计](../docs/architecture.md)
- [数据库设计文档](../docs/database-design.md)
- [Docker 部署指南](../docs/docker-deployment.md)
