# 项目协作说明

## 技术栈

- 后端：Go、Gin、Casbin。
- 前端：Vue 3、TypeScript、Vite、Element Plus。
- 前端代码位于 `web/`；完成前端改动后运行 `cd web && pnpm typecheck`。

## 界面规范

- 保持界面简洁、紧凑，右侧业务内容的基础字体使用 11px；数据数值与页面主标题可按层级适度放大。
- 侧栏宽度：展开 180px，收起 56px；分栏线应保持可拖拽，右侧内容随剩余空间自适应。
- 使用已有设计令牌（如 `--primary`、`--line`、`--text`），不要在无必要时引入新的颜色体系。
- 表格操作保持轻量化。编辑操作采用描边按钮，而非实心主按钮：
  - 白色背景；
  - 1px 蓝紫色描边；
  - 蓝紫色文字；
  - 约 8px 圆角；
  - 轻微阴影；
  - 悬停时使用很浅的蓝紫色背景，不使用大面积渐变或位移效果。
- “删除”使用危险色文字操作；“清空选择”等次要操作使用无背景的文字按钮。
- 权限配置按业务分类展示。每个分类必须支持全选、取消全选和半选状态，且分类操作不能影响其他分类。

## 接口规范

- 所有业务接口统一使用 HTTP `POST`，不使用 `GET`、`PUT`、`PATCH` 或 `DELETE`。
- 接口路径采用动作式命名，避免同一路径的查询与写入冲突，例如：`/users/list`、`/users/create`、`/users/update`、`/users/delete`。
- 查询条件与写入参数均通过 JSON 请求体传递；删除和更新操作的资源 ID 也放在请求体中。
- 新增或调整接口时，必须同步修改 Gin 路由、Casbin 权限目录、前端 API 封装、页面调用及 `v-permission` 权限键。

## 修改原则

- 优先复用 Element Plus 组件和现有 CSS 变量。
- 样式优先放在对应 Vue 组件的 scoped 样式中；跨页面通用规范放在 `web/src/styles/global.css`。
- 避免修改与当前需求无关的文件或覆盖已有用户改动。

## 数据模型与查询代码

- 数据模型位于 `internal/data/model/`，使用 PostgreSQL model 格式：包含 `TableName<模型名>` 常量、`// <模型> mapped from table <表名>` 注释，以及指针接收器的 `TableName()` 方法。
- model 字段统一使用 `gorm:"column:<name>;...;comment:<字段说明>" json:"<name>"` 格式；`comment:` 必须存在，值与字段业务说明一致。
- 字段业务说明必须写在字段定义的同一行末尾 `//` 注释中，不单独占行，并与 PostgreSQL 表/字段注释保持一致。例如：`Name string \`gorm:"column:name;not null;comment:组织名称" json:"name"\` // 组织名称`。
- 除非字段类型映射存在歧义，不强制在 model Tag 中写 `type:`；以表结构迁移和 GORM 默认映射为准。
- 数据结构采用 SQL 优先：新增或变更表时，先新增 Goose SQL 到 `docs/sql/`，执行迁移后使用 GORM Gen 从 PostgreSQL 表结构生成 model 与 query。
- 禁止手工修改生成的 `.gen.go` model/query 文件；字段说明、类型、默认值和约束应定义在 Goose SQL 中，再重新生成代码。
- 在 `internal/data/table_model.go` 的 `gopsTable` 中使用 `GenerateModel` 从已迁移的数据库表定义生成 model，并注册 query 生成目标。所有 model 间的 `many2many` 关系必须在此处通过 `field.RelateConfig` 与中间表名声明，禁止在生成后的 model 文件中补写。
- `func GenModel(cfg *config.Config)` 生成 `internal/data/model/` 与 `internal/data/query/` 下的代码；生成前必须先完成 Goose 升级。
- 生成命令为 `go run ./cmd/server -g`。完成后确认生成对应的 `internal/data/query/<table>.gen.go`，并运行 `go test ./...`。

## 数据库

- 数据库仅支持 PostgreSQL，连接 DSN 必须使用 PostgreSQL 格式。
- `NewDB` 固定使用 `gorm.io/driver/postgres`；`GenTable` 固定使用 Goose 的 `postgres` 驱动。
- 不使用 GORM `AutoMigrate`。表、字段、索引、约束及数据初始化均由 Goose SQL 迁移管理；`GenTable` 只执行 Goose。
- 默认根组织必须由可重复安全的 Goose 数据迁移确保存在：`code` 为 `root`，初始 `name` 为“默认组织”。根组织名称允许修改，但 `code` 不可修改、根组织不可删除。
- 所有用户必须关联默认根组织。Goose 数据迁移为历史用户补齐根组织关联；用户新增或编辑时，始终保留根组织关联，并可额外关联其他组织。
- SQL 迁移统一放在 `docs/sql/`，可使用 PostgreSQL 特性，例如 `TIMESTAMPTZ`、`ILIKE`、部分索引和 `GENERATED AS IDENTITY`。
- Goose 保留用于 PostgreSQL 的数据修复与数据补丁；数据修复迁移必须可重复安全执行，并提供对应的 Down 操作或明确的不可逆说明。
- 修改模型或迁移后，执行 PostgreSQL 迁移并运行 `go test ./...`。
