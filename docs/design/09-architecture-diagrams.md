# 系统架构图

**文档版本**: v1.0  
**设计日期**: 2026-01-17

---

## 1. 整体系统架构图

```mermaid
graph TB
    subgraph "客户端层"
        WebUI[Web 控制台]
        MobileApp[移动 App]
        API_Client[API 客户端]
    end

    subgraph "接入层"
        Gateway[API Gateway<br/>- 认证鉴权<br/>- 限流熔断<br/>- 路由转发]
    end

    subgraph "业务服务层"
        OrderSvc[订单服务<br/>Order Service]
        BillSvc[账单服务<br/>Bill Service]
        MeterSvc[计量服务<br/>Meter Service]
        PriceSvc[定价服务<br/>Price Service]
        PaySvc[支付服务<br/>Payment Service]
        SettleSvc[结算服务<br/>Settlement Service]
        AccountSvc[账户服务<br/>Account Service]
    end

    subgraph "基础服务层"
        AuthSvc[认证服务]
        NotifySvc[通知服务]
        ExchangeSvc[汇率服务]
    end

    subgraph "数据存储层"
        MySQL[(MySQL<br/>业务数据)]
        ClickHouse[(ClickHouse<br/>计量数据)]
        Redis[(Redis<br/>缓存)]
    end

    subgraph "消息队列"
        Kafka[Kafka<br/>事件总线]
    end

    subgraph "外部系统"
        ResourceMgr[资源管理系统<br/>GPU/Storage]
        PaymentGateway[第三方支付<br/>支付宝/微信]
        ExchangeAPI[汇率API]
    end

    WebUI --> Gateway
    MobileApp --> Gateway
    API_Client --> Gateway

    Gateway --> OrderSvc
    Gateway --> BillSvc
    Gateway --> AccountSvc
    Gateway --> PaySvc

    OrderSvc --> PriceSvc
    OrderSvc --> MySQL
    OrderSvc --> Kafka

    BillSvc --> MySQL
    BillSvc --> ClickHouse
    BillSvc --> Kafka

    MeterSvc --> ClickHouse
    MeterSvc --> Kafka

    PriceSvc --> MySQL
    PriceSvc --> Redis
    PriceSvc --> ExchangeSvc

    PaySvc --> MySQL
    PaySvc --> AccountSvc
    PaySvc --> PaymentGateway
    PaySvc --> Kafka

    SettleSvc --> MySQL
    SettleSvc --> BillSvc
    SettleSvc --> AccountSvc

    AccountSvc --> MySQL
    AccountSvc --> Redis

    ExchangeSvc --> MySQL
    ExchangeSvc --> ExchangeAPI

    Kafka --> MeterSvc
    Kafka --> BillSvc
    Kafka --> NotifySvc

    ResourceMgr -.计量上报.-> MeterSvc
```

---

## 2. 核心业务流程图

### 预付费下单流程

```mermaid
sequenceDiagram
    actor User as 用户
    participant UI as Web界面
    participant Order as 订单服务
    participant Price as 定价服务
    participant Bill as 账单服务
    participant Pay as 支付服务
    participant Account as 账户服务
    participant Resource as 资源管理

    User->>UI: 选择GPU包年
    UI->>Price: 查询价格
    Price-->>UI: 返回价格
    UI->>Order: 创建订单
    Order->>Price: 计算订单金额
    Price-->>Order: 返回总价
    Order->>Bill: 生成账单
    Bill-->>Order: 账单创建成功
    Order-->>UI: 订单创建成功
    
    User->>UI: 确认支付
    UI->>Pay: 发起支付
    Pay->>Account: 扣减余额
    Account-->>Pay: 扣减成功
    Pay->>Bill: 更新账单状态
    Pay->>Order: 更新订单状态
    Pay-->>UI: 支付成功
    
    Order->>Resource: 创建GPU实例
    Resource-->>Order: 实例创建成功
    Order-->>UI: 资源开通完成
```

### 后付费计量出账流程

```mermaid
sequenceDiagram
    participant Resource as GPU实例
    participant Meter as 计量服务
    participant ClickHouse as ClickHouse
    participant Bill as 账单服务
    participant Account as 账户服务
    participant Notify as 通知服务

    loop 每秒计量
        Resource->>Meter: 上报使用数据
        Meter->>ClickHouse: 写入计量记录
    end

    Note over ClickHouse: 物化视图自动聚合小时数据

    loop 每小时整点
        Bill->>ClickHouse: 查询上小时聚合数据
        ClickHouse-->>Bill: 返回汇总数据
        Bill->>Bill: 计算费用
        Bill->>Bill: 生成账单
        Bill->>Account: 扣减余额
        Account-->>Bill: 扣减成功
        Bill->>Notify: 发送账单通知
    end
```

---

## 3. 数据流转架构

```mermaid
graph LR
    A[资源使用] -->|秒级上报| B[Kafka]
    B --> C[计量服务]
    C --> D[ClickHouse<br/>秒级明细]
    D -->|物化视图| E[ClickHouse<br/>小时聚合]
    E -->|整点触发| F[账单服务]
    F --> G[MySQL<br/>账单表]
    G --> H[支付服务]
    H --> I[账户余额]
```

---

## 4. 部署架构图

```mermaid
graph TB
    subgraph "负载均衡层"
        LB[Load Balancer<br/>Nginx/HAProxy]
    end

    subgraph "API网关层"
        GW1[API Gateway 1]
        GW2[API Gateway 2]
        GW3[API Gateway N]
    end

    subgraph "业务服务层 (K8s)"
        subgraph "Order Service"
            O1[Pod 1]
            O2[Pod 2]
        end
        subgraph "Bill Service"
            B1[Pod 1]
            B2[Pod 2]
        end
        subgraph "Meter Service"
            M1[Pod 1]
            M2[Pod 2]
        end
    end

    subgraph "数据层"
        subgraph "MySQL集群"
            MySQL_M[Master]
            MySQL_S1[Slave 1]
            MySQL_S2[Slave 2]
        end
        subgraph "ClickHouse集群"
            CH1[Shard 1]
            CH2[Shard 2]
            CH3[Shard N]
        end
        subgraph "Redis集群"
            R1[Master 1]
            R2[Master 2]
        end
    end

    LB --> GW1
    LB --> GW2
    LB --> GW3

    GW1 --> O1
    GW1 --> B1
    GW1 --> M1

    O1 --> MySQL_M
    B1 --> MySQL_M
    M1 --> CH1

    MySQL_M -.复制.-> MySQL_S1
    MySQL_M -.复制.-> MySQL_S2
```

---

## 5. 技术栈组件图

```mermaid
mindmap
  root((Happy Billing))
    后端服务
      Go 1.25
        Gin/Echo框架
        GORM
        gRPC
      微服务
        API Gateway
        Service Mesh
    数据存储
      MySQL 8.0
        主从复制
        分库分表
      ClickHouse 23.x
        分布式表
        物化视图
      Redis 7.x
        缓存
        分布式锁
      Kafka 3.x
        事件总线
        计量数据流
    监控运维
      Prometheus
      Grafana
      ELK Stack
      Jaeger追踪
    基础设施
      Kubernetes
      Docker
      Helm Charts
      CI/CD
```

---

## 6. 数据库分库分表策略

```mermaid
graph TB
    subgraph "应用层"
        App[应用服务]
    end

    subgraph "Sharding-Proxy"
        Proxy[ShardingSphere Proxy]
    end

    subgraph "MySQL分库 (按tenant_id哈希)"
        DB0[(billing_db_0<br/>租户ID % 4 = 0)]
        DB1[(billing_db_1<br/>租户ID % 4 = 1)]
        DB2[(billing_db_2<br/>租户ID % 4 = 2)]
        DB3[(billing_db_3<br/>租户ID % 4 = 3)]
    end

    subgraph "ClickHouse分布式表"
        CH[(metering_records_dist<br/>按月分区)]
    end

    App --> Proxy
    Proxy --> DB0
    Proxy --> DB1
    Proxy --> DB2
    Proxy --> DB3
    App --> CH
```

---

## 7. 高可用设计

### MySQL高可用

```
MySQL Master (写)
    ↓ 异步复制
MySQL Slave 1 (读)
MySQL Slave 2 (读)
    ↓ 主库故障
自动切换 (MHA/MGR)
    ↓
Slave 1 提升为 Master
```

### ClickHouse高可用

```
Shard 1: Replica 1, Replica 2
Shard 2: Replica 1, Replica 2
Shard 3: Replica 1, Replica 2

ZooKeeper集群协调
```

### Redis高可用

```
Sentinel模式:
  Master (写)
  Slave 1 (读)
  Slave 2 (读)
  Sentinel监控自动故障转移
```

---

## 相关文档

- [系统架构](./01-architecture.md)
- [账单模型](./06-billing-models.md)
