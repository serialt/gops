-- +goose Up
-- 表：系统配置键值表（迁移过程中的临时表）
CREATE TABLE "system_config_kv" (
  "config_key" VARCHAR(128) PRIMARY KEY, -- 配置键
  "config_value" TEXT NOT NULL DEFAULT '', -- 配置值
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  "updated_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP -- 更新时间
);

INSERT INTO "system_config_kv" ("config_key", "config_value")
SELECT 'system_name', "system_name" FROM "system_config" WHERE "id" = 1
ON CONFLICT ("config_key") DO UPDATE SET "config_value" = EXCLUDED."config_value", "updated_at" = CURRENT_TIMESTAMP;

INSERT INTO "system_config_kv" ("config_key", "config_value")
SELECT 'system_description', "system_description" FROM "system_config" WHERE "id" = 1
ON CONFLICT ("config_key") DO UPDATE SET "config_value" = EXCLUDED."config_value", "updated_at" = CURRENT_TIMESTAMP;

DROP TABLE "system_config";
ALTER TABLE "system_config_kv" RENAME TO "system_config";

-- +goose Down
CREATE TABLE "system_config_legacy" (
  "id" INTEGER PRIMARY KEY, -- 配置主键
  "system_name" VARCHAR(128) NOT NULL DEFAULT 'V3 Admin', -- 系统名称
  "system_description" VARCHAR(500) NOT NULL DEFAULT '', -- 系统描述
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  "updated_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP -- 更新时间
);
INSERT INTO "system_config_legacy" ("id", "system_name", "system_description")
SELECT 1,
  COALESCE((SELECT "config_value" FROM "system_config" WHERE "config_key" = 'system_name'), 'V3 Admin'),
  COALESCE((SELECT "config_value" FROM "system_config" WHERE "config_key" = 'system_description'), '');
DROP TABLE "system_config";
ALTER TABLE "system_config_legacy" RENAME TO "system_config";
