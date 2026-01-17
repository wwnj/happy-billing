-- Happy Billing MySQL 初始化脚本
-- 创建数据库已由环境变量 MYSQL_DATABASE 自动创建

-- 设置字符集
SET NAMES utf8mb4;
SET character_set_client = utf8mb4;

-- 创建测试表（用于健康检查）
USE happy_billing;

CREATE TABLE IF NOT EXISTS `health_check` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `status` VARCHAR(20) NOT NULL DEFAULT 'ok',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='健康检查表';

-- 插入初始数据
INSERT INTO `health_check` (`status`) VALUES ('ok') ON DUPLICATE KEY UPDATE `status` = 'ok';

-- 显示数据库信息
SELECT
    CONCAT('✅ Database: ', DATABASE()) as info
UNION ALL
SELECT
    CONCAT('✅ Character Set: ', @@character_set_database)
UNION ALL
SELECT
    CONCAT('✅ Collation: ', @@collation_database)
UNION ALL
SELECT
    CONCAT('✅ Time Zone: ', @@global.time_zone);
