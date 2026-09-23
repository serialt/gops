# Gops

Gops 是一个面向运维场景的管理平台，采用 Go + Vue 3 的前后端分离架构，覆盖权限管理、资产管理、域名监控、告警、Nginx、任务、SSL 证书、审计和 Kubernetes 管理等功能。

> 当前项目使用 SQLite 作为唯一数据库。所有业务接口采用 `POST` 请求和动作式路径，内置 `admin` 超级管理员默认拥有全部菜单与接口权限。

## 功能概览

- **系统管理**：用户、角色权限、组织架构、系统基础配置、安全配置和验证码配置。
- **资产管理**：资产分组、凭证、云账号、主机、数据库、Web 终端及终端审计。
- **监控告警**：域名可用性、HTTPS/证书检查、最近检查历史、邮件及常见机器人告警通道、告警接收人和告警日志。
- **Kubernetes**：集群、节点、命名空间、工作负载、Pod、网络、配置、存储、访问控制、终端审计、应用诊断和集群巡检。
- **运维工具**：Nginx 数据源与访问日志、任务作业、SSL 证书管理。
- **安全审计**：登录日志、操作日志和数据变更日志。

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
- pnpm
- Git

SQLite 驱动采用纯 Go 实现，本地开发不需要安装 PostgreSQL，也不需要额外启动数据库服务。

### 1. 获取代码

```bash
git clone <repository-url>
cd gin-vue
```

### 2. 配置后端

在项目根目录创建 `config.yaml`：

```yaml
dsn: "gops.db"

log:
  file: ""
  level: "dev"

server:
  http_port: 9876
  jwt_secret: "请替换为足够长的随机字符串"
  aes_key: "请替换为 16、24 或 32 字节的 AES 密钥"

redis:
  host: "127.0.0.1"
  port: 6379
  password: ""
  db: 0
  pool_size: 10
  min_idle_conn: 1
```

配置项说明：

| 配置项 | 说明 |
| --- | --- |
| `dsn` | SQLite 文件路径或 SQLite URI |
| `log.level` | 日志级别；`debug` 输出 GORM SQL，`dev` 配合 `-dev` 执行开发种子数据 |
| `server.http_port` | 后端监听端口，前端开发代理默认使用 `9876` |
| `server.jwt_secret` | JWT 签名密钥，生产环境必须替换 |
| `server.aes_key` | 用于加密数据库中的凭证、密码等敏感数据 |
| `redis.*` | Redis 连接参数，供需要缓存或分布式能力的模块使用 |

不要将包含真实密钥的 `config.yaml` 提交到版本库。

### 3. 初始化数据库并启动后端

首次运行先执行 Goose 迁移，再启动服务：

```bash
go mod download
go run ./cmd/server -t up
go run ./cmd/server -dev
```

开发种子数据仅在同时指定 `-dev` 且 `log.level` 为 `dev` 时执行；普通启动不会执行 `seed`。

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

# 重置指定用户密码；命令执行后服务会继续启动
go run ./cmd/server -u admin 'NewPassword'
```

也可以使用 Makefile：

```bash
make help             # 查看全部构建目标
make deps             # 安装前后端依赖
make build            # 构建内置前端资源的单文件程序
make check            # 后端测试与前端类型检查
make migrate-up       # 执行 Goose 迁移
make generate         # 生成 GORM model/query
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
│   │   ├── model/              # GORM Gen 生成的数据模型
│   │   └── query/              # GORM Gen 生成的查询代码
│   ├── handler/                # Gin 参数绑定与响应转换
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
├── config.yaml                 # 本地运行配置（不要提交真实密钥）
├── Dockerfile                  # 后端镜像
├── Makefile
└── go.mod
```

## 架构约定

### 后端分层

请求按以下方向流转：

```text
Gin Router -> Middleware -> Handler -> Biz -> GORM Gen / External Service
```

- `handler` 仅负责请求绑定、读取中间件上下文、调用 Biz 和返回响应。
- 数据查询、事务、业务校验及外部服务调用放在 `biz`。
- 对外 DTO 放在 `internal/data/api`，接口不得直接序列化数据库 model。
- 通用响应结构为 `{ code, message, data }`。

### API 约定

- API 前缀：`/api/v1`。
- 除 WebSocket 连接外，业务接口统一使用 `POST`。
- 路径采用动作式命名，例如 `/users/list`、`/users/create`、`/users/update`、`/users/delete`。
- 查询、更新和删除参数统一通过 JSON 请求体传递。
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

仓库根目录的 `Dockerfile` 会在多阶段构建中编译前端，并将其内置到最终后端镜像；不再需要额外的前端容器。运行容器时需要挂载生产 `config.yaml` 和 SQLite 数据目录，并确保 `dsn` 指向持久化目录中的数据库文件。`web/Dockerfile` 仅在需要独立部署前端时使用。

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

## License

仓库当前未提供开源许可证文件。如需对外发布或分发，请先补充明确的许可证。
