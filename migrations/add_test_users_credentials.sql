-- 为测试用户添加登录凭证
-- 说明: 添加username和password_hash字段，密码为"123456"
-- bcrypt hash for "123456": $2a$10$YourHashHere (需要在运行时生成)

-- 更新个人开发者的主账号用户
-- 用户名: testuser
-- 密码: 123456
UPDATE `users`
SET
    `username` = 'testuser',
    `password_hash` = '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'
WHERE `user_id` = 'user_d6e3b5f8c2';

-- 更新企业用户的主账号用户
-- 用户名: admin
-- 密码: 123456
UPDATE `users`
SET
    `username` = 'admin',
    `password_hash` = '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'
WHERE `user_id` = 'user_b1f4a7d9e3';

-- 验证更新结果
SELECT user_id, tenant_id, username, real_name, email, phone, is_primary, status
FROM `users`
WHERE `username` IS NOT NULL;
