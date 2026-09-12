-- +goose Up
CREATE TABLE "system_config_kv" (
  "config_key" VARCHAR(128) PRIMARY KEY,
  "config_value" TEXT NOT NULL DEFAULT '',
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
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
  "id" INTEGER PRIMARY KEY,
  "system_name" VARCHAR(128) NOT NULL DEFAULT 'V3 Admin',
  "system_description" VARCHAR(500) NOT NULL DEFAULT '',
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO "system_config_legacy" ("id", "system_name", "system_description")
SELECT 1,
  COALESCE((SELECT "config_value" FROM "system_config" WHERE "config_key" = 'system_name'), 'V3 Admin'),
  COALESCE((SELECT "config_value" FROM "system_config" WHERE "config_key" = 'system_description'), '');
DROP TABLE "system_config";
ALTER TABLE "system_config_legacy" RENAME TO "system_config";
