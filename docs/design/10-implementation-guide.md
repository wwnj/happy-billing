# 实施指南

**文档版本**: v1.0  
**设计日期**: 2026-01-17

---

## 环境准备

### 前置条件

| 组件 | 版本要求 | 用途 |
|------|---------|------|
| Go | 1.25+ | 应用开发语言 |
| MySQL | 8.0+ | 业务数据存储 |
| ClickHouse | 23.x+ | 计量数据存储 |
| Redis | 7.x+ | 缓存与分布式锁 |
| Kafka | 3.x+ | 消息队列 |
| Docker | 20.x+ | 容器化部署 |
| Kubernetes | 1.28+ | 容器编排 |

---

## 开发环境搭建

### 1. 克隆项目

```bash
git clone https://github.com/wwnj/happy-billing.git
cd happy-billing
go mod download
```

### 2. 启动基础服务

使用 Docker Compose 快速启动所有依赖服务。

### 3. 初始化数据库

```bash
# MySQL 初始化
mysql -h 127.0.0.1 -u root -p < scripts/sql/mysql/schema.sql

# ClickHouse 初始化
clickhouse-client --host 127.0.0.1 < scripts/sql/clickhouse/metering.sql
```

---

## 生产环境部署

### Kubernetes 部署

```bash
# 创建命名空间
kubectl create namespace billing

# 部署服务
kubectl apply -f deployments/k8s/ -n billing

# 查看状态
kubectl get pods -n billing
```

### 高可用配置

- **MySQL**: 主从复制 + 半同步
- **ClickHouse**: 3分片 × 2副本
- **Redis**: Sentinel 哨兵模式
- **Kafka**: 3节点集群 + 副本因子3

---

## 监控与告警

### Prometheus 指标

- **业务指标**: 订单量、账单量、支付成功率
- **服务指标**: QPS、延迟P99、错误率
- **数据库指标**: 连接数、慢查询、复制延迟

### Grafana 面板

预置面板:
- Billing Overview
- Database Metrics
- ClickHouse Metrics

---

## 性能调优

### MySQL 优化

```ini
# my.cnf
innodb_buffer_pool_size = 16G
innodb_log_file_size = 2G
max_connections = 1000
```

### ClickHouse 优化

- 合理设置分区策略
- 使用物化视图加速聚合查询
- 定期清理过期数据

---

## 备份策略

- **MySQL**: 每日全量备份 + binlog增量备份
- **ClickHouse**: 每周全量备份
- **备份保留**: 30天

---

## 相关文档

- [系统架构](./01-architecture.md)
- [架构图](./09-architecture-diagrams.md)
