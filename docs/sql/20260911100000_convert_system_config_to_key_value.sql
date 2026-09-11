-- +goose Up
CREATE TABLE "system_config_kv" (
  "config_key" VARCHAR(128) PRIMARY KEY,
  "config_value" TEXT NOT NULL DEFAULT '',
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO "system_config_kv" ("config_key", "config_value")
SELECT 'system_name', "system_name" FROM "system_config" WHERE "id" = 1
ON CONFLICT ("config_key") DO UPDATE SET "config_value" = EXCLUDED."config_value", "updated_at" = CURRENT_TIMESTAMP;

INSERT INTO "system_config_kv" ("config_key", "config_value")
SELECT 'system_description', "system_description" FROM "system_config" WHERE "id" = 1
ON CONFLICT ("config_key") DO UPDATE SET "config_value" = EXCLUDED."config_value", "updated_at" = CURRENT_TIMESTAMP;

DROP TABLE "system_config";
ALTER TABLE "system_config_kv" RENAME TO "system_config";

COMMENT ON TABLE "system_config" IS '系统配置键值表';
COMMENT ON COLUMN "system_config"."config_key" IS '配置键';
COMMENT ON COLUMN "system_config"."config_value" IS '配置值';
COMMENT ON COLUMN "system_config"."created_at" IS '创建时间';
COMMENT ON COLUMN "system_config"."updated_at" IS '更新时间';

-- +goose Down
CREATE TABLE "system_config_legacy" (
  "id" BIGINT PRIMARY KEY,
  "system_name" VARCHAR(128) NOT NULL DEFAULT 'V3 Admin',
  "system_description" VARCHAR(500) NOT NULL DEFAULT '',
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO "system_config_legacy" ("id", "system_name", "system_description")
SELECT 1,
  COALESCE((SELECT "config_value" FROM "system_config" WHERE "config_key" = 'system_name'), 'V3 Admin'),
  COALESCE((SELECT "config_value" FROM "system_config" WHERE "config_key" = 'system_description'), '');
DROP TABLE "system_config";
ALTER TABLE "system_config_legacy" RENAME TO "system_config";
