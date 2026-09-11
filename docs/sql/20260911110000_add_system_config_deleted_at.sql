-- +goose Up
ALTER TABLE "system_config"
  ADD COLUMN IF NOT EXISTS "deleted_at" TIMESTAMPTZ;

COMMENT ON COLUMN "system_config"."deleted_at" IS '软删除时间';

-- +goose Down
ALTER TABLE "system_config"
  DROP COLUMN IF EXISTS "deleted_at";
