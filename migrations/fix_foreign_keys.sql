-- 修复产品模块测试数据的外键引用
-- 问题：使用了数字 ID 而不是业务 ID (category_id, spu_id)

USE happy_billing;

-- 1. 修复 product_categories 表的 parent_id
UPDATE product_categories SET parent_id = 'CAT20240117001' WHERE parent_id = '1';
UPDATE product_categories SET parent_id = 'CAT20240117004' WHERE parent_id = '4';
UPDATE product_categories SET parent_id = 'CAT20240117007' WHERE parent_id = '7';

-- 2. 修复 product_spu 表的 category_id
UPDATE product_spu SET category_id = 'CAT20240117002' WHERE category_id = '2';

-- 3. 修复 product_sku 表的 spu_id
UPDATE product_sku SET spu_id = 'SPU20240117001' WHERE spu_id = '1';
UPDATE product_sku SET spu_id = 'SPU20240117002' WHERE spu_id = '2';

-- 验证修复结果
SELECT '=== Product Categories ===' AS info;
SELECT category_id, category_code, category_name, parent_id, level
FROM product_categories
ORDER BY level, sort_order;

SELECT '=== Product SPU ===' AS info;
SELECT spu_id, spu_code, spu_name, category_id, product_type
FROM product_spu;

SELECT '=== Product SKU ===' AS info;
SELECT sku_id, sku_code, spu_id, spu_code, sku_name
FROM product_sku;
