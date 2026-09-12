-- +goose Up
-- 字段：软删除时间
ALTER TABLE "system_config" ADD COLUMN "deleted_at" DATETIME;

-- +goose Down
ALTER TABLE "system_config" DROP COLUMN "deleted_at";
