-- +goose Up



-- 表：系统角色表
CREATE TABLE IF NOT EXISTS "role" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT, -- 角色主键
    "code" VARCHAR(64) NOT NULL, -- 角色编码，系统内唯一
    "name" VARCHAR(64) NOT NULL, -- 角色名称
    "description" VARCHAR(255), -- 角色说明
    "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    "updated_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 更新时间
    "deleted_at" DATETIME, -- 软删除时间
    CONSTRAINT uq_role_code UNIQUE ("code")
);

-- 表：系统用户表
CREATE TABLE IF NOT EXISTS "user" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT, -- 用户主键
    "username" VARCHAR(64) NOT NULL, -- 登录用户名，系统内唯一
    "password_hash" VARCHAR(255) NOT NULL, -- 密码哈希
    "nickname" VARCHAR(64) NOT NULL, -- 用户昵称
    "status" BOOLEAN NOT NULL DEFAULT TRUE, -- 启用状态
    "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    "updated_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 更新时间
    "deleted_at" DATETIME, -- 软删除时间
    CONSTRAINT uq_user_username UNIQUE ("username")
);

-- 表：用户与角色关联表
CREATE TABLE IF NOT EXISTS "user_role" (
    "user_id" INTEGER NOT NULL, -- 用户主键
    "role_id" INTEGER NOT NULL, -- 角色主键
    PRIMARY KEY ("user_id", "role_id")
 
);

-- 优化5: 针对软删除的局部索引（Partial Index）
-- 业务中通常只查询未删除的数据，局部索引可以大幅减小索引体积并提升查询速度
CREATE UNIQUE INDEX IF NOT EXISTS idx_role_code_active ON "role" ("code") WHERE "deleted_at" IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_username_active ON "user" ("username") WHERE "deleted_at" IS NULL;

-- 优化6: 为关联表的外键字段创建索引（加速 JOIN 查询）
-- 虽然 PK 包含了 user_id，但按 role_id 查询时仍需索引
CREATE INDEX IF NOT EXISTS idx_user_role_role_id ON "user_role" ("role_id");

-- +goose Down
