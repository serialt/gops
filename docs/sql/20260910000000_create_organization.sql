-- +goose Up
-- 表：组织架构表
CREATE TABLE IF NOT EXISTS "organization" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT, -- 组织主键
  "parent_id" INTEGER REFERENCES "organization"("id") ON DELETE RESTRICT, -- 上级组织主键，根组织为空
  "name" VARCHAR(64) NOT NULL, -- 组织名称
  "code" VARCHAR(64) NOT NULL, -- 组织编码，系统内唯一
  "manager" VARCHAR(64) NOT NULL DEFAULT '', -- 组织负责人
  "status" BOOLEAN NOT NULL DEFAULT TRUE, -- 启用状态
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  "updated_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 更新时间
  "deleted_at" DATETIME, -- 软删除时间
  "sort_order" INTEGER NOT NULL DEFAULT 0 -- 同级组织排序值
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_organization_code_active ON "organization" ("code") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_organization_parent_id ON "organization" ("parent_id");

-- 表：组织与用户关联表
CREATE TABLE IF NOT EXISTS "organization_user" (
  "organization_id" INTEGER NOT NULL REFERENCES "organization"("id") ON DELETE CASCADE, -- 组织主键
  "user_id" INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE, -- 用户主键
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
