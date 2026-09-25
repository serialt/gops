-- +goose Up
ALTER TABLE db_instance ADD COLUMN group_id BIGINT NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_db_instance_group_active ON db_instance (group_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_db_instance_group_active;
ALTER TABLE db_instance DROP COLUMN group_id;
