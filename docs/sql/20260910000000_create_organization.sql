-- +goose Up
CREATE TABLE IF NOT EXISTS "organization" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT,
  "parent_id" INTEGER REFERENCES "organization"("id") ON DELETE RESTRICT,
  "name" VARCHAR(64) NOT NULL,
  "code" VARCHAR(64) NOT NULL,
  "manager" VARCHAR(64) NOT NULL DEFAULT '',
  "status" BOOLEAN NOT NULL DEFAULT TRUE,
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" DATETIME,
  "sort_order" INTEGER NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_organization_code_active ON "organization" ("code") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_organization_parent_id ON "organization" ("parent_id");

CREATE TABLE IF NOT EXISTS "organization_user" (
  "organization_id" INTEGER NOT NULL REFERENCES "organization"("id") ON DELETE CASCADE,
  "user_id" INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
  PRIMARY KEY ("organization_id", "user_id")
);
CREATE INDEX IF NOT EXISTS idx_organization_user_user_id ON "organization_user" ("user_id");

-- 默认根组织及历史用户的默认组织关联。
INSERT INTO "organization" ("name", "code", "manager", "status")
SELECT '默认组织', 'root', '', TRUE
WHERE NOT EXISTS (
  SELECT 1 FROM "organization" WHERE "code" = 'root' AND "deleted_at" IS NULL
);

INSERT INTO "organization_user" ("organization_id", "user_id")
SELECT root."id", users."id"
FROM "organization" AS root
CROSS JOIN "user" AS users
WHERE root."code" = 'root'
  AND root."deleted_at" IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM "organization_user" AS relation
    WHERE relation."organization_id" = root."id" AND relation."user_id" = users."id"
  );


-- +goose Down
DROP TABLE IF EXISTS "organization_user";
DROP TABLE IF EXISTS "organization";
