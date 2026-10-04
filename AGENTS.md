# 项目协作说明

## 技术栈

- 后端：Go、Gin、Casbin。
- 前端：Vue 3、TypeScript、Vite、Element Plus。
- 前端代码位于 `web/`；完成前端改动后运行 `cd web && pnpm typecheck`。

## 界面规范

- 保持界面简洁、紧凑，右侧业务内容的基础字体使用 11px；数据数值与页面主标题可按层级适度放大。
- 侧栏宽度：展开 180px，收起 36px；分栏线应保持可拖拽，右侧内容随剩余空间自适应。
- 使用已有设计令牌（如 `--primary`、`--line`、`--text`），不要在无必要时引入新的颜色体系。
- 表格操作保持轻量化。编辑操作采用描边按钮，而非实心主按钮：
  - 白色背景；
  - 1px 蓝紫色描边；
  - 蓝紫色文字；
  - 约 8px 圆角；
  - 轻微阴影；
  - 悬停时使用很浅的蓝紫色背景，不使用大面积渐变或位移效果。
- “删除”使用危险色文字操作；“清空选择”等次要操作使用无背景的文字按钮。
- Element Plus 按钮必须按语义区分实心、描边和文字样式，禁止让全局实心主按钮规则覆盖 `plain`、`text` 或 `link` 按钮：
  - 全局 `.el-button--primary` 实心样式必须使用 `:not(.is-plain):not(.is-text):not(.is-link)` 限定；
  - 蓝紫渐变背景的实心主按钮必须使用白色文字；需要实心按钮时使用普通 `type="primary"`，不要叠加 `text` 或 `plain`；
  - 主色描边按钮使用 `type="primary" plain`，应保持白色或极浅主色背景、主色文字和主色描边，不得出现渐变背景；
  - 主色文字操作使用 `type="primary" text`，应保持透明背景和主色文字，不得出现实心背景、渐变、阴影或位移；
  - 危险文字操作使用 `type="danger" text`；新增局部按钮后必须检查默认、悬停、聚焦和禁用状态下的文字对比度。
- 权限配置按业务分类展示。每个分类必须支持全选、取消全选和半选状态，且分类操作不能影响其他分类。
- 所有前端数据列表统一使用自适应布局，列表容器填充主内容区的可用空间，窄屏时允许合理换行或滚动。
- 资产管理列表的卡片与网格必须采用全高纵向 flex 布局，网格伸展填满工具栏以下的剩余空间，分页贴近卡片底边；卡片底部不得保留导致分页下方出现明显空白的额外内边距，布局参照审计管理“登录日志”页面。
- 所有业务数据列表统一使用 `AppDataGrid`（AG Grid），不得新增 Element Plus `el-table`；列必须可拖拽调整宽度。
- AG Grid 操作列必须固定在表格右侧（`pinned: 'right'`），横向滚动时始终可见；操作列使用固定宽度，按钮在操作列内左对齐。
- 普通数据列按表头和单元格内容收缩；操作列左侧最后一个数据列接收剩余宽度。用户拖拽其他列后，必须重新将余量分配给该列，表格中不得出现无字段对应的空白区域。
- 列表操作按钮统一使用圆角矩形样式：编辑为蓝紫描边浅底色，删除为危险色描边浅底色，配置权限为成功色浅底色；状态标签沿用同样的圆角矩形视觉，不使用胶囊形状。
- 列表、组织树等搜索框采用输入即搜索/过滤；不得依赖“查询”按钮触发。需要请求接口时，输入和清空均应立即刷新结果。
- 所有列表筛选文本输入框宽度按 placeholder（提示文字）的实际渲染宽度自适应，并计入输入框左右内边距、边框、前缀/清除图标等实际占用空间；不设置输入长度上限，输入超出框宽时在输入框内横向滚动，不得随输入内容扩宽或填满剩余空间。该规则只作用于列表筛选输入，不影响表单、弹窗及多行文本输入。
- 列表不得直接展示后端资源 ID，首列统一显示连续序号；ID 仅用于编辑、删除、关联等内部操作。
- 列表底部右侧统一使用中文分页，必须显示总条数；分页文字和控件字号统一为 10px。
- 所有列表操作列默认使用 `150px` 宽度；角色权限页因“配置权限、编辑、删除”三个操作使用 `220px` 宽度。编辑使用描边按钮，删除使用危险色文字操作。
- 顶部导航统一包含菜单搜索、固定“首页”Tab 和动态页面 Tab；菜单搜索只展示当前用户有权限访问的页面。
- 内置 `admin`（超级管理员）默认拥有全部菜单和全部接口权限；新增菜单、路由或 Casbin 权限目录项后，必须确保 `admin` 无需额外配置即可访问，并在登录态刷新后展示对应菜单。
- 页面 Tab 根据标题内容自适应宽度，不使用固定宽度；Tab 过多时允许横向滚动，当前路由必须高亮，非首页 Tab 支持关闭。
- 顶部导航右侧保留用户信息和退出入口；搜索区域默认紧凑显示图标，聚焦后展开输入框，不得挤压用户信息区域。
- 所有页面主内容区必须填充可用高度和宽度；侧栏、分隔线、内容区和页面卡片使用 `min-width: 0`，避免横向溢出。
- 桌面端页面内容使用统一紧凑间距；窄屏时侧栏和业务分栏允许收缩或改单列，工具栏、Tab 和分页允许换行或横向滚动。

## 接口规范

- 所有业务接口统一使用 HTTP `POST`，不使用 `GET`、`PUT`、`PATCH` 或 `DELETE`。
- 接口路径采用动作式命名，避免同一路径的查询与写入冲突，例如：`/users/list`、`/users/create`、`/users/update`、`/users/delete`。
- 查询条件与写入参数均通过 JSON 请求体传递；删除和更新操作的资源 ID 也放在请求体中。
- 文件上传和下载是 JSON 请求体约定的例外：上传使用 `multipart/form-data`，Gin handler 通过 `c.FormFile`/`c.Request.FormFile` 获取文件并传给 Biz 处理，不得将文件名和 base64/二进制内容包装进 JSON；下载由 Gin 直接返回文件流或附件（例如 `c.File`、`c.FileAttachment` 或 `c.DataFromReader`），设置正确的 `Content-Type` 和 `Content-Disposition`，不得把文件名和 base64/二进制内容放入 JSON 响应。前端上传使用 `FormData`，下载以 `Blob` 接收并从响应头解析文件名。
- 新增或调整接口时，必须同步修改 Gin 路由、Casbin 权限目录、前端 API 封装、页面调用及 `v-permission` 权限键。
- 每个对外 Gin handler 方法必须紧邻方法声明前添加 Go 文档注释，格式为：`// 方法名 POST /api/v1/完整路由 功能说明。`；方法与路由变更时必须同步更新该注释。
- 所有数据列表接口必须接收 `page`、`limit`，由后端查询层执行 `Count` 和 `Offset/Limit`；统一响应包络中的 `data` 必须包含当前页数据 `list` 和匹配总数 `total`，结构为 `{ "code": 0, "message": "", "data": { "list": [...], "total": n } }`。`total` 必须是筛选条件命中的记录总数，不是当前页数组长度，供前端分页器显示总条数和计算页数；前端不得获取全量数据后自行分页。树形选项、关联下拉等非列表展示场景也必须显式传分页参数，并根据业务需要设置受控的 `limit`。
- 前端列表请求中的 `page` 必须取列表右下角分页器的当前页码，`limit` 必须取分页器当前选择的每页条数，分页总数读取响应 `data.total`；切换每页条数时将 `page` 重置为 `1` 后重新请求，切换页码时保持当前 `limit`。
- 列表导出必须复用当前列表的全部筛选条件，不受当前分页页码或每页条数限制；用户有勾选记录时，只导出“当前筛选结果”与“已勾选记录”的交集，未勾选时导出当前筛选条件命中的全部记录。筛选、选择状态及导出请求参数必须保持一致，后端负责按筛选条件和所选资源标识执行导出。

## 后端分层规范

- `internal/handler/` 只负责 Gin 请求绑定、读取 middleware 上下文、调用 `internal/biz/` 和将结果转换为 HTTP 响应；禁止在 handler 中直接访问数据库、Casbin、密码或 JWT 业务逻辑。
- 每个业务域由 `internal/biz/` 提供对应的 Biz 类型和方法；Biz 负责编排事务边界、业务校验、model 转换及外部服务调用，数据库读写通过 `internal/data/repo/` 提供的 Repo 完成，禁止 Biz 直接访问数据库或生成的 Query。
- 请求和响应 DTO 统一定义在 `internal/data/api/`，按业务域拆分文件；handler 和 biz 共同使用这些 DTO，禁止在接口响应中直接返回 `internal/data/model/` 生成模型。
- API 请求 DTO 在 handler 中使用 `new(api.XxxRequest)` 创建并以指针传递，Biz 方法接收请求 DTO 指针；API 响应 DTO 和列表元素使用指针返回，嵌套响应对象同样使用指针。
- API 响应结构必须显式定义需要公开的字段，敏感字段（例如密码哈希）不得通过 model 序列化暴露；关联对象也必须使用 API DTO。
- 新增 handler 时，构造函数只注入对应 Biz；路由注册处负责创建并注入 Biz，禁止让 handler 持有 `*gorm.DB`。

## 修改原则

- 优先复用 Element Plus 组件和现有 CSS 变量。
- 样式优先放在对应 Vue 组件的 scoped 样式中；跨页面通用规范放在 `web/src/styles/global.css`。
- 避免修改与当前需求无关的文件或覆盖已有用户改动。

## 数据模型与查询代码

- 所有业务数据库读写必须封装在 `internal/data/repo/` 中；`internal/data/query/` 生成的类型安全 Query API 只能由 Repo 层直接使用，handler、biz 及其他业务代码禁止直接引用生成的 Query。
- Repo 必须使用 GORM Gen Query API 完成查询和写入，禁止使用 `db.Where`、`db.First`、`db.Find`、`db.Raw` 等方式拼写业务查询。新增字段或 Query 能力不足时，应先完成 SQL 迁移并重新生成代码。
- Repo 方法必须显式接收 `*gorm.DB` 参数，并通过 `query.Use(db)` 创建 GORM Gen Query；Repo 结构体不得长期持有数据库连接。Biz 在普通调用时传入自身数据库连接，在事务中传入当前 `tx`，以保证同一 Repo 方法同时支持普通查询和事务查询。
- Repo 中所有 `List` 类型方法必须接收定义在 `internal/data/define/` 中的 `XxxListCond` 指针，不得通过多个独立参数传递分页、关键词或业务过滤条件。API Request 仅用于接口层，Biz 负责将 Request 显式转换为对应 Cond，Repo 禁止依赖 `internal/data/api`。
- Biz 只向 Repo 透传请求中的 `page`、`limit`，不得在 Biz 中计算 offset。Repo 负责规范化分页参数、计算 `offset = (page - 1) * limit`，并在完成过滤和排序条件构造后调用 GORM Gen 的 `FindByPage(offset, limit)`，直接取得当前页数据和总数；禁止分别调用 `Count()` 与 `Offset/Limit/Find()` 拼装分页结果。
- 分页参数规范化和 offset 计算统一复用 `internal/pkg/utils.PageBounds`，禁止在各 Repo 中重复定义分页计算工具。
- 需要事务时，由 Biz 发起业务事务编排，并将事务 `*gorm.DB` 传递给 Repo；Repo 不得自行开启或提交业务事务。仅数据库初始化、Goose 迁移、GORM Gen 代码生成等基础设施代码可以直接操作底层 `*gorm.DB`。
- 数据模型位于 `internal/data/model/`，使用 SQLite model 格式：包含 `TableName<模型名>` 常量、`// <模型> mapped from table <表名>` 注释，以及指针接收器的 `TableName()` 方法。
- model 字段统一使用 `gorm:"column:<name>;..." json:"<name>"` 格式；字段业务说明通过同一行末尾注释维护。
- 字段业务说明必须写在字段定义的同一行末尾 `//` 注释中，不单独占行，并与 PostgreSQL 表/字段注释保持一致。例如：`Name string \`gorm:"column:name;not null;comment:组织名称" json:"name"\` // 组织名称`。
- 除非字段类型映射存在歧义，不强制在 model Tag 中写 `type:`；以表结构迁移和 GORM 默认映射为准。
- 数据结构采用 SQL 优先：新增或变更表时，先新增 Goose SQL 到 `docs/sql/`，执行迁移后使用 GORM Gen 从 SQLite 表结构生成 model 与 query。
- 禁止手工修改生成的 `.gen.go` model/query 文件；字段说明、类型、默认值和约束应定义在 Goose SQL 中，再重新生成代码。
- 在 `internal/data/table_model.go` 的 `gopsTable` 中使用 `GenerateModel` 从已迁移的 SQLite 数据库表定义生成 model，并注册 query 生成目标。所有 model 间的 `many2many` 关系必须在此处通过 `field.RelateConfig` 与中间表名声明，禁止在生成后的 model 文件中补写。
- `func GenModel(cfg *config.Config)` 生成 `internal/data/model/` 与 `internal/data/query/` 下的代码；生成前必须先完成 Goose 升级。
- 生成命令为 `go run ./cmd/server -g`。完成后确认生成对应的 `internal/data/query/<table>.gen.go`，并运行 `go test ./...`。

## 数据库

- 数据库仅支持 SQLite，连接 DSN 为 SQLite 数据库文件路径或 SQLite URI。
- `NewDB` 固定使用 `github.com/glebarez/sqlite`（底层驱动 `github.com/glebarez/go-sqlite`）；`GenTable` 使用 Goose 的 `sqlite3` 方言，并由 `github.com/glebarez/go-sqlite` 打开连接。
- 不使用 GORM `AutoMigrate`。表、字段、索引、约束及数据初始化均由 Goose SQL 迁移管理；`GenTable` 只执行 Goose。
- 默认根组织必须由可重复安全的 Goose 数据迁移确保存在：`code` 为 `root`，初始 `name` 为“默认组织”。根组织名称允许修改，但 `code` 不可修改、根组织不可删除。
- 所有用户必须关联默认根组织。Goose 数据迁移为历史用户补齐根组织关联；用户新增或编辑时，始终保留根组织关联，并可额外关联其他组织。
- SQL 迁移统一放在 `docs/sql/`，须使用 SQLite 兼容语法，例如 `INTEGER PRIMARY KEY AUTOINCREMENT`、`DATETIME` 和 `ON CONFLICT`；不得使用 PostgreSQL 专有语法或 `COMMENT ON`。
- 含 `deleted_at` 软删除字段的表如需唯一约束，唯一性必须只作用于未删除记录：使用 `CREATE UNIQUE INDEX ... ON <table>(<unique_columns>) WHERE deleted_at IS NULL`。禁止在字段或表定义中使用不含软删除条件的 `UNIQUE`，否则软删除后无法重新创建相同业务值；也禁止使用 `UNIQUE(<unique_columns>, deleted_at)`，因为 SQLite 允许唯一索引中存在多个 `NULL`，无法保证未删除记录唯一。修改既有唯一约束时必须提供 Goose 迁移，移除旧的表级/字段级唯一约束后再建立部分唯一索引。
- 业务代码禁止调用 GORM 或 GORM Gen 的 `Unscoped()`；查询、更新和删除均必须遵循默认软删除作用域。`internal/data/query/*.gen.go` 中由 GORM Gen 自动生成但未被调用的 `Unscoped()` 方法定义不做手工修改，业务代码不得引用。新增业务数据必须始终插入新记录，不得清空历史记录的 `deleted_at`，不得复用软删除记录的 ID，也不得通过 `Save`、`FirstOrCreate`、`Assign` 或 upsert 将“新增”隐式实现为恢复。软删除历史必须保持不变；是否恢复记录应由独立、明确授权的恢复接口处理。创建前的重复校验只查询未删除记录，最终仍由带 `WHERE deleted_at IS NULL` 的部分唯一索引保证并发唯一性。
- Goose 保留用于 SQLite 的数据修复与数据补丁；数据修复迁移必须可重复安全执行，并提供对应的 Down 操作或明确的不可逆说明。
- 修改模型或迁移后，执行 SQLite 迁移并运行 `go test ./...`。
