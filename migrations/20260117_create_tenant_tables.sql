-- Happy Billing 租户模块数据库表
-- 创建时间: 2026-01-17
-- 说明: 租户、组织、项目、用户、认证表

-- ============================================================================
-- 1. 租户表 (tenants)
-- 说明: 租户是系统的顶层隔离单位，支持企业和个人用户
-- ============================================================================
CREATE TABLE IF NOT EXISTS `tenants` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `tenant_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: tenant_a3f9b2c4d5',
  `tenant_code` VARCHAR(64) UNIQUE NOT NULL COMMENT '租户编码',
  `name` VARCHAR(255) NOT NULL COMMENT '租户名称',
  `tenant_type` VARCHAR(32) NOT NULL COMMENT '租户类型: ENTERPRISE/INDIVIDUAL',
  `preferred_currency` VARCHAR(8) DEFAULT 'CNY' COMMENT '首选币种',

  `verified` TINYINT DEFAULT 0 COMMENT '是否已认证: 0-未认证, 1-已认证',
  `verified_type` VARCHAR(32) COMMENT '认证类型: INDIVIDUAL/ENTERPRISE',
  `verified_at` TIMESTAMP NULL COMMENT '认证时间',
  `verified_info` JSON COMMENT '认证信息',

  `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_tenant_id` (`tenant_id`),
  INDEX `idx_type` (`tenant_type`),
  INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户表';

-- ============================================================================
-- 2. 组织表 (organizations)
-- 说明: 组织是企业主体，支持树形结构（部门层级）
-- ============================================================================
CREATE TABLE IF NOT EXISTS `organizations` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `organization_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: org_b4e1c3f2a6',

  `tenant_id` VARCHAR(64) NOT NULL COMMENT '所属租户ID',
  `parent_organization_id` VARCHAR(64) COMMENT '父组织ID（树形结构）',

  `org_code` VARCHAR(64) UNIQUE NOT NULL COMMENT '组织编码',
  `name` VARCHAR(255) NOT NULL COMMENT '组织名称',
  `org_type` VARCHAR(32) COMMENT '组织类型: 总部/分公司/部门等',
  `level` INT DEFAULT 1 COMMENT '组织层级',

  `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_organization_id` (`organization_id`),
  INDEX `idx_tenant` (`tenant_id`),
  INDEX `idx_parent` (`parent_organization_id`),
  FOREIGN KEY (`tenant_id`) REFERENCES `tenants`(`tenant_id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='组织表';

-- ============================================================================
-- 3. 项目表 (projects)
-- 说明: 项目是成本中心/工作区，用于资源隔离和成本核算
-- ============================================================================
CREATE TABLE IF NOT EXISTS `projects` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `project_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: proj_c5d2a4e7b1',

  `tenant_id` VARCHAR(64) NOT NULL COMMENT '所属租户ID',
  `organization_id` VARCHAR(64) NOT NULL COMMENT '所属组织ID',

  `project_code` VARCHAR(64) UNIQUE NOT NULL COMMENT '项目编码',
  `name` VARCHAR(255) NOT NULL COMMENT '项目名称',
  `description` TEXT COMMENT '项目描述',

  `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_project_id` (`project_id`),
  INDEX `idx_tenant` (`tenant_id`),
  INDEX `idx_org` (`organization_id`),
  INDEX `idx_status` (`status`),
  FOREIGN KEY (`tenant_id`) REFERENCES `tenants`(`tenant_id`) ON DELETE RESTRICT,
  FOREIGN KEY (`organization_id`) REFERENCES `organizations`(`organization_id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目表';

-- ============================================================================
-- 4. 用户表 (users)
-- 说明: 用户是实际使用者，个人用户有主账号，企业用户可以有多个员工账号
-- ============================================================================
CREATE TABLE IF NOT EXISTS `users` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `user_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: user_d6e3b5f8c2',

  `tenant_id` VARCHAR(64) NOT NULL COMMENT '所属租户ID',

  `is_primary` TINYINT DEFAULT 0 COMMENT '是否主账号: 0-否, 1-是',
  `username` VARCHAR(128) COMMENT '用户名',
  `real_name` VARCHAR(128) COMMENT '真实姓名',
  `id_card` VARCHAR(64) COMMENT '身份证号',

  `email` VARCHAR(255) COMMENT '邮箱',
  `phone` VARCHAR(32) COMMENT '手机号',
  `password_hash` VARCHAR(255) COMMENT '密码哈希',

  `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_tenant` (`tenant_id`),
  INDEX `idx_email` (`email`),
  INDEX `idx_phone` (`phone`),
  FOREIGN KEY (`tenant_id`) REFERENCES `tenants`(`tenant_id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ============================================================================
-- 5. 认证表 (verifications)
-- 说明: 实名认证记录，支持个人认证和企业认证
-- ============================================================================
CREATE TABLE IF NOT EXISTS `verifications` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `verification_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: verify_e7f4c6a9d3',

  `tenant_id` VARCHAR(64) NOT NULL COMMENT '所属租户ID',
  `user_id` VARCHAR(64) NOT NULL COMMENT '申请用户ID',

  `verify_type` VARCHAR(32) NOT NULL COMMENT '认证类型: INDIVIDUAL/ENTERPRISE',

  -- 个人认证字段
  `real_name` VARCHAR(128) COMMENT '真实姓名',
  `id_card` VARCHAR(64) COMMENT '身份证号',
  `id_card_front` VARCHAR(512) COMMENT '身份证正面照URL',
  `id_card_back` VARCHAR(512) COMMENT '身份证背面照URL',

  -- 企业认证字段
  `company_name` VARCHAR(255) COMMENT '公司名称',
  `credit_code` VARCHAR(64) COMMENT '统一社会信用代码',
  `license_url` VARCHAR(512) COMMENT '营业执照URL',
  `legal_person` VARCHAR(128) COMMENT '法人代表',

  `status` VARCHAR(32) NOT NULL DEFAULT 'PENDING' COMMENT '状态: PENDING/APPROVED/REJECTED',
  `reject_reason` TEXT COMMENT '拒绝原因',
  `verified_at` TIMESTAMP NULL COMMENT '审核通过时间',
  `verified_by` VARCHAR(64) COMMENT '审核人',

  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_verification_id` (`verification_id`),
  INDEX `idx_tenant` (`tenant_id`),
  INDEX `idx_user` (`user_id`),
  INDEX `idx_status` (`status`),
  FOREIGN KEY (`tenant_id`) REFERENCES `tenants`(`tenant_id`) ON DELETE RESTRICT,
  FOREIGN KEY (`user_id`) REFERENCES `users`(`user_id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='认证表';

-- ============================================================================
-- 初始化数据
-- ============================================================================

-- 插入测试租户（个人开发者）
INSERT INTO `tenants` (`tenant_id`, `tenant_code`, `name`, `tenant_type`, `verified`)
VALUES ('tenant_a3f9b2c4d5', 'test_individual', '测试个人开发者', 'INDIVIDUAL', 1);

-- 插入测试租户（企业用户）
INSERT INTO `tenants` (`tenant_id`, `tenant_code`, `name`, `tenant_type`, `verified`)
VALUES ('tenant_e8d7c2a1f4', 'test_enterprise', '测试企业用户', 'ENTERPRISE', 1);

-- 插入测试组织（个人工作区）
INSERT INTO `organizations` (`organization_id`, `tenant_id`, `org_code`, `name`, `org_type`)
VALUES ('org_b4e1c3f2a6', 'tenant_a3f9b2c4d5', 'personal_workspace', '个人工作区', 'PERSONAL');

-- 插入测试组织（企业总部）
INSERT INTO `organizations` (`organization_id`, `tenant_id`, `org_code`, `name`, `org_type`)
VALUES ('org_f7a2d5e8c1', 'tenant_e8d7c2a1f4', 'headquarters', '总部', 'HEADQUARTERS');

-- 插入测试项目
INSERT INTO `projects` (`project_id`, `tenant_id`, `organization_id`, `project_code`, `name`)
VALUES ('proj_c5d2a4e7b1', 'tenant_a3f9b2c4d5', 'org_b4e1c3f2a6', 'default_project', '默认项目');

INSERT INTO `projects` (`project_id`, `tenant_id`, `organization_id`, `project_code`, `name`)
VALUES ('proj_a9f3e6b2c8', 'tenant_e8d7c2a1f4', 'org_f7a2d5e8c1', 'ai_platform', 'AI平台项目');

-- 插入测试用户
INSERT INTO `users` (`user_id`, `tenant_id`, `is_primary`, `real_name`, `email`, `phone`)
VALUES ('user_d6e3b5f8c2', 'tenant_a3f9b2c4d5', 1, '张三', 'zhangsan@example.com', '13800138000');

INSERT INTO `users` (`user_id`, `tenant_id`, `is_primary`, `real_name`, `email`, `phone`)
VALUES ('user_b1f4a7d9e3', 'tenant_e8d7c2a1f4', 1, '李四', 'lisi@example.com', '13900139000');
