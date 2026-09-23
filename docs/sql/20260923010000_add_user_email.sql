-- +goose Up
ALTER TABLE "user" ADD COLUMN "email" VARCHAR(255) NOT NULL DEFAULT ''; -- 用户邮箱，用于登录邮件验证
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_email_active
ON "user" ("email")
WHERE "deleted_at" IS NULL AND "email" <> '';

-- +goose Down
DROP INDEX IF EXISTS idx_user_email_active;
ALTER TABLE "user" DROP COLUMN "email";
