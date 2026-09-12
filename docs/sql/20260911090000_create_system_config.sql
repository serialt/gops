-- +goose Up
-- 表：系统基础配置表
CREATE TABLE IF NOT EXISTS "system_config" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT, -- 配置主键，固定为 1
  "system_name" VARCHAR(128) NOT NULL DEFAULT 'V3 Admin', -- 系统名称
  "system_description" VARCHAR(500) NOT NULL DEFAULT '', -- 系统描述
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  "updated_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP -- 更新时间
);

INSERT INTO "system_config" ("id", "system_name", "system_description")
SELECT 1, 'V3 Admin', ''
WHERE NOT EXISTS (SELECT 1 FROM "system_config" WHERE "id" = 1);

-- +goose Down
DROP TABLE IF EXISTS "system_config";
