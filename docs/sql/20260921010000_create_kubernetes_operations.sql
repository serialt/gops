-- +goose Up
CREATE TABLE IF NOT EXISTS kubernetes_terminal_session (id INTEGER PRIMARY KEY AUTOINCREMENT, cluster_id INTEGER NOT NULL, namespace TEXT NOT NULL DEFAULT '', pod_name TEXT NOT NULL DEFAULT '', container_name TEXT NOT NULL DEFAULT '', user_id INTEGER NOT NULL, recording TEXT NOT NULL DEFAULT '', duration INTEGER NOT NULL DEFAULT 0, command_count INTEGER NOT NULL DEFAULT 0, status TEXT NOT NULL DEFAULT 'completed', started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, ended_at DATETIME, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP); -- Kubernetes 终端审计
CREATE INDEX IF NOT EXISTS idx_kubernetes_terminal_session_cluster ON kubernetes_terminal_session(cluster_id);
CREATE INDEX IF NOT EXISTS idx_kubernetes_terminal_session_user ON kubernetes_terminal_session(user_id);

CREATE TABLE IF NOT EXISTS kubernetes_cluster_inspection (id INTEGER PRIMARY KEY AUTOINCREMENT, cluster_id INTEGER NOT NULL, cluster_name TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'running', score INTEGER NOT NULL DEFAULT 0, check_count INTEGER NOT NULL DEFAULT 0, pass_count INTEGER NOT NULL DEFAULT 0, warning_count INTEGER NOT NULL DEFAULT 0, fail_count INTEGER NOT NULL DEFAULT 0, duration INTEGER NOT NULL DEFAULT 0, report_data TEXT NOT NULL DEFAULT '', user_id INTEGER NOT NULL, started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, ended_at DATETIME, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP); -- Kubernetes 集群巡检
CREATE INDEX IF NOT EXISTS idx_kubernetes_cluster_inspection_cluster ON kubernetes_cluster_inspection(cluster_id);

-- +goose Down
DROP TABLE IF EXISTS kubernetes_cluster_inspection;
DROP TABLE IF EXISTS kubernetes_terminal_session;
