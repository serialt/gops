# Gops

Gops 是一个面向运维场景的管理平台。开发阶段采用 Go + Vue 3 前后端分离架构，生产构建可将前端资源嵌入单个 Go 可执行文件，覆盖权限管理、资产管理、域名监控、告警、任务、SSL 证书、IoT、审计和 Kubernetes 管理等功能。

> 当前项目使用 SQLite 作为唯一数据库。所有业务接口采用 `POST` 请求和动作式路径，内置 `admin` 超级管理员默认拥有全部菜单与接口权限。

## 功能概览

- **系统管理**：用户、角色权限、组织架构、系统基础配置、安全配置和验证码配置。
- **资产管理**：资产分组、凭证、云账号、主机、数据库、Web 终端及终端审计。
- **监控告警**：域名可用性、HTTPS/证书检查、最近检查历史、邮件及常见机器人告警通道、告警接收人和告警日志。
- **Kubernetes**：集群、节点、命名空间、工作负载、Pod、网络、配置、存储、访问控制、终端审计、应用诊断和集群巡检。
- **运维工具**：任务作业、SSL 证书管理。
- **IoT 管理**：内置 MQTT Broker、外部 Broker 桥接、Topic、ACL、客户端连接和消息记录。
- **安全审计**：登录日志、操作日志和数据变更日志。
- **账号安全**：JWT 双 Token、角色权限、验证码、MFA/WebAuthn 等能力。

## 技术栈

| 层级 | 主要技术 |
| --- | --- |
| 后端 | Go、Gin、GORM、GORM Gen、Goose、SQLite |
| 认证与授权 | JWT、Casbin RBAC、bcrypt |
| 前端 | Vue 3、TypeScript、Vite、Pinia、Vue Router |
| UI 与表格 | Element Plus、TDesign Vue Next、AG Grid |
| 终端与集群 | xterm.js、Kubernetes client-go、WebSocket |

## 快速开始

### 环境要求

- Go 1.27 或更高版本
- Node.js 22
- pnpm 10（建议通过 Corepack 启用）
- Git

SQLite 驱动采用纯 Go 实现，本地开发不需要安装 PostgreSQL，也不需要额外启动数据库服务。

### 1. 获取代码

```bash
git clone <repository-url>
cd gin-vue
```

### 2. 配置后端

复制配置模板：

```bash
cp config.yaml.example config.yaml
```

最小配置如下：

```yaml
dsn: "gops.db"

log:
  file: ""
  level: "info"

server:
  http_port: 9876
  jwt_secret: "请替换为足够长的随机字符串"
  aes_key: "请替换为 16、24 或 32 字节的 AES 密钥"
  access_token_timeout: 30m
  refresh_token_timeout: 168h

iot:
  enabled: false
```

配置项说明：

| 配置项 | 说明 |
| --- | --- |
| `dsn` | SQLite 文件路径或 SQLite URI |
| `log.level` | 日志级别；使用 `debug` 可输出更详细的调试信息 |
| `server.http_port` | 后端监听端口，前端开发代理默认使用 `9876` |
| `server.jwt_secret` | JWT 签名密钥，生产环境必须替换 |
| `server.aes_key` | 用于加密数据库中的凭证、密码等敏感数据 |
| `server.access_token_timeout` | Access Token 有效期，默认 `15m` |
| `server.refresh_token_timeout` | Refresh Token 有效期，默认 `168h`（7 天） |
| `iot.enabled` | 是否启动内置 MQTT 服务；监听器详细配置见 `config.yaml.example` |

运行时缓存使用进程内 `go-cache`。缓存随进程重启清空，不在多个服务实例之间共享。

不要将包含真实密钥的 `config.yaml` 提交到版本库。

### 3. 初始化数据库并启动后端

安装依赖并启动服务：

```bash
go mod download
go run ./cmd/server
```

数据库文件不存在时，程序会自动执行全部 Goose 迁移，并创建默认组织、权限数据和管理员。已有数据库启动前建议显式执行迁移：

```bash
go run ./cmd/server -t up
go run ./cmd/server
```

后端默认地址为 `http://localhost:9876`，健康检查：

```bash
curl -X POST http://localhost:9876/health
```

### 4. 启动前端

另开一个终端：

```bash
cd web
pnpm install
pnpm dev
```

打开 `http://localhost:5173`。Vite 会将 `/api` 请求代理到 `http://localhost:9876`。

### 5. 登录

首次启动会创建超级管理员：

- 用户名：`admin`
- 初始密码：`password`

请在首次登录后立即修改密码。已有数据库不会重复创建或覆盖管理员账号。

## 常用命令

### 后端

```bash
# 运行全部测试
go test ./...

# 查看版本信息
go run ./cmd/server -v

# 使用指定配置文件启动
go run ./cmd/server -c /path/to/config.yaml

# 执行或回滚数据库迁移
go run ./cmd/server -t up
go run ./cmd/server -t down
go run ./cmd/server -t status

# 基于当前 SQLite 表结构重新生成 model 和 query
go run ./cmd/server -g

# 列出用户
go run ./cmd/server -u list

# 创建管理员（自动关联 admin 角色和默认根组织；昵称可省略）
go run ./cmd/server -u create operator 'StrongPassword' '运维用户'

# 重置指定用户密码
go run ./cmd/server -u reset admin 'NewPassword'

# 兼容旧版重置密码写法
go run ./cmd/server -u admin 'NewPassword'
```

也可以使用 Makefile：

```bash
make help             # 查看全部构建目标
make deps             # 安装前后端依赖
make build            # 构建内置前端资源的单文件程序
make check            # 后端测试与前端类型检查
make migrate-up       # 执行 Goose 迁移
make migration-create NAME=add_example  # 创建迁移文件
make gmodel           # 生成 GORM model/query
make wire             # 重新生成 Wire 依赖注入代码
make release          # 构建多平台单文件程序
```

### 前端

```bash
cd web
pnpm dev        # 开发服务器
pnpm typecheck  # TypeScript 类型检查
pnpm build      # 类型检查并构建生产资源
pnpm preview    # 本地预览生产构建
```

## 目录结构

```text
.
├── cmd/server/                 # 后端入口、Wire 注入代码
├── docs/
│   └── sql/                    # Goose SQLite 迁移
├── internal/
│   ├── biz/                    # 业务逻辑、事务、数据转换和外部服务调用
│   ├── config/                 # YAML 配置结构与加载
│   ├── data/
│   │   ├── api/                # 请求与响应 DTO
│   │   ├── define/             # Repo 查询条件等内部定义
│   │   ├── model/              # GORM Gen 生成的数据模型
│   │   ├── query/              # GORM Gen 生成的查询代码
│   │   └── repo/               # 数据访问与分页查询
│   ├── handler/                # Gin 参数绑定与响应转换
│   ├── iot/                    # MQTT Broker 与外部 Broker 连接
│   ├── middleware/             # JWT、Casbin、审计等中间件
│   ├── pkg/                    # 认证、加密、权限、响应等基础能力
│   └── server/                 # 路由注册、初始化和默认数据
├── web/
│   ├── src/
│   │   ├── api/                # 按业务域拆分的 API 与 DTO
│   │   ├── components/         # 通用组件（含 AppDataGrid）
│   │   ├── layouts/            # 管理后台布局
│   │   ├── plugins/            # 业务插件页面
│   │   ├── router/             # 路由与访问控制
│   │   ├── stores/             # Pinia 状态
│   │   ├── utils/              # Axios 等通用封装
│   │   └── views/              # 页面视图
│   ├── Dockerfile
│   └── nginx.conf
├── container/
│   ├── Dockerfile              # 后端镜像
│   ├── docker-compose.yml      # 容器编排配置
│   ├── docker/                 # 镜像基础文件与入口脚本
│   └── helm/                   # Kubernetes Helm Chart
├── config.yaml.example         # 配置模板
├── config.yaml                 # 本地运行配置（不要提交真实密钥）
├── Makefile
└── go.mod
```

## 架构约定

### 后端分层

请求按以下方向流转：

```text
Gin Router -> Middleware -> Handler -> Biz -> Repo -> GORM Gen Query
                                           └------> External Service
```

- `handler` 仅负责请求绑定、读取中间件上下文、调用 Biz 和返回响应。
- `biz` 负责编排事务、业务校验、DTO/model 转换和外部服务调用，不直接使用生成的 Query。
- `repo` 封装全部业务数据库读写，通过 GORM Gen Query 访问 SQLite；事务由 Biz 创建后传入 Repo。
- 对外 DTO 放在 `internal/data/api`，接口不得直接序列化数据库 model。
- 通用响应结构为 `{ code, message, data }`。

### API 约定

- API 前缀：`/api/v1`。
- 除 WebSocket 连接外，业务接口统一使用 `POST`。
- 路径采用动作式命名，例如 `/users/list`、`/users/create`、`/users/update`、`/users/delete`。
- 查询、更新和删除参数统一通过 JSON 请求体传递。
- 文件上传使用 `multipart/form-data`，文件下载直接返回文件流；二进制内容不包装进 JSON。
- 列表接口接收 `page`、`limit`，返回 `data.list` 和筛选结果总数 `data.total`；分页由后端执行。
- 登录后通过 `Authorization: Bearer <token>` 访问受保护接口。

登录示例：

```bash
curl -X POST http://localhost:9876/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"password"}'
```

### 权限模型

Casbin 策略由角色、API 路径和 HTTP 方法正则组成：

```text
p, role_code, api_path, http_method_regex
g, user_id, role_code
```

内置超级管理员策略：

```text
p, admin, /api/v1/*, .*
```

新增路由、菜单或权限目录项时，需要同步更新 Gin 路由、Casbin 权限目录、前端 API、页面权限键和菜单配置。

### 数据库迁移与代码生成

项目采用 SQL 优先流程，不使用 GORM `AutoMigrate`：

1. 在 `docs/sql/` 新增 SQLite 兼容的 Goose SQL。
2. 执行 `go run ./cmd/server -t up` 更新数据库。
3. 执行 `go run ./cmd/server -g` 生成 `internal/data/model` 和 `internal/data/query`。
4. 执行 `go test ./...` 验证后端。

创建迁移文件可使用：

```bash
make migration-create NAME=add_example
```

迁移必须使用 SQLite 兼容语法。含 `deleted_at` 的表需要业务唯一性时，应使用带 `WHERE deleted_at IS NULL` 的部分唯一索引。

生成的 `*.gen.go` 文件不要手工修改；字段、索引、默认值和注释应在 SQL 迁移中维护。

## 构建与部署

### 构建单文件程序

```bash
make build
./dist/gops -t up
./dist/gops -c ./config.yaml
```

`make build` 会先执行前端类型检查和生产构建，再将 `web/dist` 嵌入 Go 二进制。最终的 `dist/gops` 同时提供 API、前端静态资源和 Vue Router 的 SPA 回退，不需要单独部署前端文件或 Nginx。

### 构建前端

```bash
cd web
pnpm install --frozen-lockfile
pnpm build
```

单独执行前端构建时，结果位于 `web/dist/`，可用于前端资源调试或独立部署。常规生产发布推荐直接使用 `make build` 生成的单文件程序。

### Docker

`container/Dockerfile` 会以仓库根目录作为构建上下文，在多阶段构建中编译前端并将其内置到最终后端镜像；不再需要额外的前端容器。运行容器时使用 `container/docker-compose.yml`，并挂载 SQLite 数据目录。`web/Dockerfile` 仅在需要独立部署前端时使用。

```bash
# 在仓库根目录构建镜像
docker build -f container/Dockerfile -t gin-vue:local .

# 生成生产密钥；也可写入仅本机可读的 .env 文件
export GOPS_JWT_SECRET="$(openssl rand -hex 32)"
export GOPS_AES_KEY="$(openssl rand -hex 16)"

# Compose 默认将数据持久化到宿主机 /opt/app/gops/data
docker compose -f container/docker-compose.yml up -d
```

常用容器变量：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `GOPS_HTTP_PORT` | `9876` | HTTP 监听及宿主机映射端口 |
| `GOPS_DB_DSN` | `/app/data/gops.db` | 容器内 SQLite 路径 |
| `GOPS_LOG_LEVEL` | `info` | 日志级别 |
| `GOPS_JWT_SECRET` | 示例值 | JWT 密钥，生产环境必须覆盖 |
| `GOPS_AES_KEY` | 示例值 | AES 密钥，生产环境必须覆盖 |
| `GOPS_ACCESS_TOKEN_TIMEOUT` | `15m` | Access Token 有效期 |
| `GOPS_REFRESH_TOKEN_TIMEOUT` | `168h` | Refresh Token 有效期 |
| `GOPS_IOT_ENABLED` | `false` | 是否启动内置 MQTT 服务 |

Compose 中的 Traefik 标签包含示例域名 `gops.imau.cc`，部署前请替换为自己的域名；未使用 Traefik 时可移除这些标签。

### Kubernetes / Helm

```bash
helm upgrade --install gops ./container/helm \
  --namespace gops --create-namespace \
  --set image.repository=registry.example.com/team/gops \
  --set image.tag=v1.0.0 \
  --set-string config.jwtSecret="$GOPS_JWT_SECRET" \
  --set-string config.aesKey="$GOPS_AES_KEY"
```

SQLite 默认使用 `ReadWriteOnce` PVC，Deployment 采用 `Recreate` 策略，建议保持单副本。生产环境优先通过 `config.existingSecret` 引用已有 Secret，完整参数见 `container/helm/README.md`。

## 升级与备份

升级前停止写入并备份 SQLite 数据库及配置文件。SQLite 数据库应复制主文件及可能存在的 `-wal`、`-shm` 文件，或在服务运行时使用 SQLite 在线备份能力，避免只复制主文件造成数据不一致。

```bash
# 停止服务后备份示例
cp gops.db "gops.db.$(date +%Y%m%d%H%M%S).bak"

# 拉取新版本后执行迁移并启动
go run ./cmd/server -t up
go run ./cmd/server
```

如需回滚一条迁移，先确认对应 Down 迁移可逆并已完成备份，再执行 `go run ./cmd/server -t down`。

## 开发检查

提交改动前至少执行：

```bash
go test ./...
cd web && pnpm typecheck
```

涉及数据库结构时，还应在干净的 SQLite 数据库上验证完整迁移；涉及前端生产构建时执行 `pnpm build`。

## 安全建议

- 首次登录后修改默认管理员密码。
- 生产环境使用高强度且独立的 JWT 密钥与 AES 密钥。
- 不要提交真实 SMTP 密码、云密钥、KubeConfig、资产凭证或数据库文件。
- 使用 HTTPS，并限制管理端和终端 WebSocket 的网络访问范围。
- 定期备份 SQLite 数据文件，并在升级前验证 Goose 迁移与回滚方案。
- 按最小权限原则分配角色；`admin` 仅用于平台管理。

## 常见问题

### 前端请求返回 401 或 403

- `401`：确认 Access Token 未过期，并检查请求头是否为 `Authorization: Bearer <token>`。
- `403`：确认角色已配置对应 API 权限；新增接口时需要同步更新 Gin 路由、Casbin 权限目录、菜单和前端权限键。
- 内置 `admin` 默认匹配 `/api/v1/*` 的全部方法；刷新登录态后会加载新增菜单。

### 修改数据库结构后代码不匹配

按“SQL 迁移 → 执行迁移 → `make gmodel` → `go test ./...`”的顺序操作，不要直接修改 `internal/data/model/*.gen.go` 或 `internal/data/query/*.gen.go`。

### 修改前端后单文件程序仍显示旧页面

开发时使用 `pnpm dev`；生产二进制需要重新执行 `make build`，前端资源才会重新嵌入 `dist/gops`。

## License

仓库当前未提供开源许可证文件。如需对外发布或分发，请先补充明确的许可证。
