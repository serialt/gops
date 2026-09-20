-- +goose Up
CREATE TABLE IF NOT EXISTS kubernetes_user_kube_config (id INTEGER PRIMARY KEY AUTOINCREMENT, cluster_id INTEGER NOT NULL, user_id INTEGER NOT NULL, service_account TEXT NOT NULL, namespace TEXT NOT NULL DEFAULT 'gops-auth', is_active INTEGER NOT NULL DEFAULT 1, created_by INTEGER NOT NULL, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, revoked_at DATETIME, UNIQUE(cluster_id, user_id, service_account)); -- 用户 Kubernetes 凭据
CREATE INDEX IF NOT EXISTS idx_kubernetes_user_kube_config_cluster ON kubernetes_user_kube_config(cluster_id);
CREATE INDEX IF NOT EXISTS idx_kubernetes_user_kube_config_user ON kubernetes_user_kube_config(user_id);

CREATE TABLE IF NOT EXISTS kubernetes_user_role_binding (id INTEGER PRIMARY KEY AUTOINCREMENT, cluster_id INTEGER NOT NULL, user_id INTEGER NOT NULL, role_name TEXT NOT NULL, role_namespace TEXT NOT NULL DEFAULT '', role_type TEXT NOT NULL, bound_by INTEGER NOT NULL, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, UNIQUE(cluster_id, user_id, role_name, role_namespace)); -- 用户 Kubernetes 角色绑定
CREATE INDEX IF NOT EXISTS idx_kubernetes_user_role_binding_cluster ON kubernetes_user_role_binding(cluster_id);
CREATE INDEX IF NOT EXISTS idx_kubernetes_user_role_binding_user ON kubernetes_user_role_binding(user_id);

-- +goose Down
DROP TABLE IF EXISTS kubernetes_user_role_binding;
DROP TABLE IF EXISTS kubernetes_user_kube_config;
