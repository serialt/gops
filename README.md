# Gin Vue Casbin Admin

```
gops -t create  v26.1.1_add sql
```

一个开箱即用的后台管理脚手架：

- 后端：Gin、GORM、PostgreSQL、JWT、Casbin RBAC
- 前端：Vue 3、Vite、TypeScript、Pinia、Vue Router、Element Plus
- 权限模型：用户 -> 角色 -> API 路径 + HTTP 方法
- 部署：Docker Compose（PostgreSQL + Backend + Nginx Frontend）

## 默认账号

- 用户名：`admin`
- 密码：`Admin@123456`

首次启动后请立即修改默认密码。

## Docker Compose 启动

```bash
cp .env.example .env
docker compose up -d --build
```

访问：

- 管理后台：http://localhost:8088
- 后端健康检查：http://localhost:8080/health
- PostgreSQL：localhost:5432

## 本地开发

### 1. 启动 PostgreSQL

```bash
docker compose up -d postgres
```

### 2. 启动后端

```bash
cd backend
cp .env.example .env
set -a && source .env && set +a
go mod tidy
go run ./cmd/server
```

后端默认监听 `:8080`。

### 3. 启动前端

```bash
cd frontend
npm install
npm run dev
```

前端默认访问 http://localhost:5173，并将 `/api` 代理到 `http://localhost:8080`。

## 目录结构

```text
.
├── backend
│   ├── cmd/server              # 程序入口
│   ├── internal
│   │   ├── auth                # JWT 与密码
│   │   ├── casbinx             # Casbin 初始化和策略操作
│   │   ├── config              # 环境变量配置
│   │   ├── handler             # HTTP Handler
│   │   ├── middleware          # JWT/Casbin 中间件
│   │   ├── model               # GORM 模型
│   │   ├── response            # 统一响应
│   │   └── server              # 路由与依赖组装
│   └── Dockerfile
├── frontend
│   ├── src
│   │   ├── api                 # Axios 请求
│   │   ├── directives          # v-permission
│   │   ├── layouts             # 后台布局
│   │   ├── router              # 路由和权限守卫
│   │   ├── stores              # Pinia 登录状态
│   │   └── views               # 页面
│   ├── Dockerfile
│   └── nginx.conf
└── docker-compose.yml
```

## Casbin 模型

策略三元组：

```text
p, role_code, api_path, http_method_regex
```

用户和角色关系：

```text
g, user_id, role_code
```

例如：

```text
p, admin, /api/v1/*, .*
p, viewer, /api/v1/dashboard/stats, GET
g, 1, admin
```

匹配规则使用 `keyMatch2` 匹配 API 路径，使用 `regexMatch` 匹配 HTTP 方法。

## 主要接口

| 方法       | 路径                            | 说明               |
| ---------- | ------------------------------- | ------------------ |
| POST       | `/api/v1/auth/login`            | 登录               |
| GET        | `/api/v1/auth/me`               | 当前用户与前端权限 |
| GET        | `/api/v1/dashboard/stats`       | 仪表盘统计         |
| GET/POST   | `/api/v1/users`                 | 用户列表/新增用户  |
| PUT/DELETE | `/api/v1/users/:id`             | 修改/删除用户      |
| GET/POST   | `/api/v1/roles`                 | 角色列表/新增角色  |
| PUT/DELETE | `/api/v1/roles/:id`             | 修改/删除角色      |
| GET/PUT    | `/api/v1/roles/:id/permissions` | 查询/设置角色权限  |
| GET        | `/api/v1/permissions/catalog`   | 权限目录           |

## 生产建议

- 修改 `JWT_SECRET`、数据库密码和管理员初始密码。
- 为登录接口增加限流、验证码或 MFA。
- 将密钥交给 Secret 管理系统，不要提交到 Git。
- 多实例部署时，可接入 Casbin Watcher 或 Redis Pub/Sub，同步策略变更。
- 根据业务需要增加菜单权限、数据权限、租户域或审计日志。
