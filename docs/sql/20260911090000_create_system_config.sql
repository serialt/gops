-- +goose Up
CREATE TABLE IF NOT EXISTS "system_config" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT,
  "system_name" VARCHAR(128) NOT NULL DEFAULT 'V3 Admin',
  "system_description" VARCHAR(500) NOT NULL DEFAULT '',
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO "system_config" ("id", "system_name", "system_description")
SELECT 1, 'V3 Admin', ''
WHERE NOT EXISTS (SELECT 1 FROM "system_config" WHERE "id" = 1);

-- +goose Down
DROP TABLE IF EXISTS "system_config";
