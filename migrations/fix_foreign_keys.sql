-- 修复产品模块测试数据的外键引用
-- 问题：使用了数字 ID 而不是业务 ID (category_id, spu_id)
-- 注意：此脚本支持幂等执行，只修复需要修复的数据

USE happy_billing;

-- 1. 修复 product_categories 表的 parent_id（仅修复数字ID）
UPDATE product_categories
SET parent_id = 'CAT20240117001'
WHERE parent_id = '1' AND parent_id NOT LIKE 'CAT%';

UPDATE product_categories
SET parent_id = 'CAT20240117004'
WHERE parent_id = '4' AND parent_id NOT LIKE 'CAT%';

UPDATE product_categories
SET parent_id = 'CAT20240117007'
WHERE parent_id = '7' AND parent_id NOT LIKE 'CAT%';

-- 2. 修复 product_spu 表的 category_id（仅修复数字ID）
UPDATE product_spu
SET category_id = 'CAT20240117002'
WHERE category_id = '2' AND category_id NOT LIKE 'CAT%';

-- 3. 修复 product_sku 表的 spu_id（仅修复数字ID）
UPDATE product_sku
SET spu_id = 'SPU20240117001'
WHERE spu_id = '1' AND spu_id NOT LIKE 'SPU%';

UPDATE product_sku
SET spu_id = 'SPU20240117002'
WHERE spu_id = '2' AND spu_id NOT LIKE 'SPU%';

-- 显示提示信息（可选）
SELECT '外键修复脚本执行完成' AS info;
