-- +goose Up
ALTER TABLE "system_config" ADD COLUMN "deleted_at" DATETIME;

-- +goose Down
ALTER TABLE "system_config" DROP COLUMN "deleted_at";
