#!/bin/sh
set -eu

config_path="${GOPS_CONFIG_PATH:-/app/data/config.yaml}"

cat >"${config_path}" <<EOF
dsn: "${GOPS_DB_DSN:-/app/data/gops.db}"
log:
  file: ""
  level: "${GOPS_LOG_LEVEL:-info}"
server:
  http_port: ${GOPS_HTTP_PORT:-9876}
  jwt_secret: "${GOPS_JWT_SECRET:-replace-with-a-long-random-secret}"
  aes_key: "${GOPS_AES_KEY:-0123456789abcdef0123456789abcdef}"
  access_token_timeout: "${GOPS_ACCESS_TOKEN_TIMEOUT:-15m}"
  refresh_token_timeout: "${GOPS_REFRESH_TOKEN_TIMEOUT:-168h}"
iot:
  enabled: ${GOPS_IOT_ENABLED:-false}
EOF

/usr/local/bin/gops -c "${config_path}" -t up
exec /usr/local/bin/gops -c "${config_path}"
