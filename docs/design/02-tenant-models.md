# 租户与组织模型设计

**文档版本**: v1.0  
**设计日期**: 2026-01-17

---

## 模型概述

Happy Billing 采用完整的多租户体系,支持企业用户和个人开发者:

```
Tenant (租户) - 顶层隔离单位
  └─ Organization (组织) - 企业主体,支持树形结构
       └─ Project (项目) - 成本中心/工作区
            └─ User (用户) - 实际使用者
```

---

## 核心表设计

### 1. 租户表

```sql
CREATE TABLE tenants (
  id                BIGINT PRIMARY KEY AUTO_INCREMENT,  -- 物理主键
  tenant_id         VARCHAR(64) UNIQUE NOT NULL,        -- 业务主键: T20240117001
  tenant_code       VARCHAR(64) UNIQUE NOT NULL,
  name              VARCHAR(255) NOT NULL,
  tenant_type       VARCHAR(32) NOT NULL,               -- 'ENTERPRISE'/'INDIVIDUAL'
  preferred_currency VARCHAR(8) DEFAULT 'CNY',
  
  verified          TINYINT DEFAULT 0,
  verified_type     VARCHAR(32),
  verified_at       TIMESTAMP,
  verified_info     JSON,
  
  status            TINYINT DEFAULT 1,
  created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  INDEX idx_tenant_id (tenant_id),
  INDEX idx_type (tenant_type)
);
```

**业务ID生成**: `T + YYYYMMDD + 序号`  
示例: `T20240117001`, `T20240117002`

### 2. 组织表

```sql
CREATE TABLE organizations (
  id                    BIGINT PRIMARY KEY AUTO_INCREMENT,
  organization_id       VARCHAR(64) UNIQUE NOT NULL,        -- ORG20240117001
  
  tenant_id             VARCHAR(64) NOT NULL,
  parent_organization_id VARCHAR(64),                       -- 树形结构
  
  org_code              VARCHAR(64) UNIQUE NOT NULL,
  name                  VARCHAR(255) NOT NULL,
  org_type              VARCHAR(32),
  
  created_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  INDEX idx_tenant (tenant_id),
  INDEX idx_parent (parent_organization_id)
);
```

### 3. 项目表

```sql
CREATE TABLE projects (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  project_id      VARCHAR(64) UNIQUE NOT NULL,        -- PRJ20240117001
  
  tenant_id       VARCHAR(64) NOT NULL,
  organization_id VARCHAR(64) NOT NULL,
  
  project_code    VARCHAR(64) UNIQUE NOT NULL,
  name            VARCHAR(255) NOT NULL,
  status          TINYINT DEFAULT 1,
  
  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  INDEX idx_project_id (project_id),
  INDEX idx_tenant (tenant_id),
  INDEX idx_org (organization_id)
);
```

### 4. 用户表

```sql
CREATE TABLE users (
  id          BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id     VARCHAR(64) UNIQUE NOT NULL,        -- USR20240117000001
  
  tenant_id   VARCHAR(64) NOT NULL,
  
  is_primary  TINYINT DEFAULT 0,
  real_name   VARCHAR(128),
  id_card     VARCHAR(64),
  
  email       VARCHAR(255),
  phone       VARCHAR(32),
  
  created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  INDEX idx_user_id (user_id),
  INDEX idx_tenant (tenant_id)
);
```

---

## 个人开发者支持

### 注册流程

```
个人开发者注册
    ↓
自动创建租户 (tenant_type='INDIVIDUAL')
    ↓
自动创建默认组织 ('个人工作区')
    ↓
自动创建默认项目 ('默认项目')
    ↓
创建用户账号 (is_primary=1)
    ↓
实名认证 (下单前置条件)
```

### 代码示例

```go
func RegisterIndividual(name, email, phone string) error {
    tx.Begin()
    
    // 1. 创建租户
    tenant := Tenant{
        TenantID: generateTenantID(),  // T20240117001
        Name: name,
        TenantType: "INDIVIDUAL",
    }
    tenantID := db.Insert(tenant)
    
    // 2. 自动创建默认组织
    org := Organization{
        OrganizationID: generateOrganizationID(),
        TenantID: tenant.TenantID,
        Name: "个人工作区",
    }
    db.Insert(org)
    
    // 3. 自动创建默认项目
    project := Project{
        ProjectID: generateProjectID(),
        TenantID: tenant.TenantID,
        OrganizationID: org.OrganizationID,
        Name: "默认项目",
    }
    db.Insert(project)
    
    // 4. 创建用户
    user := User{
        UserID: generateUserID(),
        TenantID: tenant.TenantID,
        IsPrimary: 1,
        Email: email,
    }
    db.Insert(user)
    
    tx.Commit()
}
```

---

## 实名认证

### 认证表设计

```sql
CREATE TABLE verifications (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id       VARCHAR(64) NOT NULL,
  user_id         VARCHAR(64) NOT NULL,
  
  verify_type     VARCHAR(32) NOT NULL,            -- 'INDIVIDUAL'/'ENTERPRISE'
  
  -- 个人认证
  real_name       VARCHAR(128),
  id_card         VARCHAR(64),
  id_card_front   VARCHAR(512),
  id_card_back    VARCHAR(512),
  
  -- 企业认证
  company_name    VARCHAR(255),
  credit_code     VARCHAR(64),
  license_url     VARCHAR(512),
  
  status          VARCHAR(32) NOT NULL,            -- 'PENDING'/'APPROVED'/'REJECTED'
  reject_reason   TEXT,
  verified_at     TIMESTAMP,
  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  INDEX idx_tenant (tenant_id),
  INDEX idx_status (status)
);
```

### 认证流程

```
用户提交认证 → 审核系统 → 审核通过 → 更新租户verified=1 → 允许下单
```

---

## 企业vs个人对比

| 维度 | 企业用户 | 个人用户 |
|------|---------|---------|
| **租户** | 1企业=1租户 | 1人=1租户 |
| **组织** | 多层级组织树 | 自动创建"个人工作区" |
| **项目** | 多个成本中心 | 默认1个,可创建多个 |
| **用户** | 多个员工账号 | 主账号1个,可邀请协作者 |
| **认证** | 企业营业执照 | 个人身份证 |
| **结算** | 支持月结/授信 | 仅余额/在线支付 |
| **发票** | 增值税专用发票 | 增值税普通发票 |

---

## 业务ID生成器

### ID生成规则

```go
type IDGenerator struct {
    redis *redis.Client
}

// 租户ID: T + YYYYMMDD + 序号(3位)
func (g *IDGenerator) GenerateTenantID() string {
    date := time.Now().Format("20060102")
    seq := g.getNextSequence("tenant:" + date)
    return fmt.Sprintf("T%s%03d", date, seq)
    // 示例: T20240117001
}

// 组织ID: ORG + YYYYMMDD + 序号(4位)
func (g *IDGenerator) GenerateOrganizationID() string {
    date := time.Now().Format("20060102")
    seq := g.getNextSequence("org:" + date)
    return fmt.Sprintf("ORG%s%04d", date, seq)
    // 示例: ORG202401170001
}

// 项目ID
func (g *IDGenerator) GenerateProjectID() string {
    date := time.Now().Format("20060102")
    seq := g.getNextSequence("project:" + date)
    return fmt.Sprintf("PRJ%s%04d", date, seq)
}

// 用户ID
func (g *IDGenerator) GenerateUserID() string {
    date := time.Now().Format("20060102")
    seq := g.getNextSequence("user:" + date)
    return fmt.Sprintf("USR%s%06d", date, seq)
}

// Redis自增序号
func (g *IDGenerator) getNextSequence(key string) int64 {
    seq, _ := g.redis.Incr(context.Background(), key).Result()
    g.redis.Expire(context.Background(), key, 24*time.Hour)
    return seq
}
```

---

## 关键设计要点

### 1. 统一租户模型

企业和个人都走租户体系,保持代码逻辑一致:
- 个人用户注册时自动创建默认组织和项目
- 个人用户可以创建多项目
- 个人用户可以升级为企业用户

### 2. 双主键设计

- **物理主键** (id): 数据库内部使用,自增
- **业务主键** (xxx_id): 业务层使用,便于沟通

优势:
- ✅ 业务ID可读性强 (`T20240117001` 比 `id=1001` 清晰)
- ✅ 跨系统集成方便 (业务ID不变)
- ✅ 便于沟通反馈
- ✅ 避免暴露内部自增ID

### 3. 实名认证前置

- 下单前必须完成实名认证
- 个人认证: OCR识别身份证
- 企业认证: 验证营业执照
- 认证通过后才能购买资源

---

## 相关文档

- [系统架构](./01-architecture.md)
- [产品模型](./03-product-models.md)
- [订单模型](./05-order-models.md)
