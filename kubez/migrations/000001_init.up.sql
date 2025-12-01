-- 创建迁移版本表（由 migrate 工具自动管理）
-- 此文件用于测试迁移功能

CREATE TABLE IF NOT EXISTS test_migration (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
