package casbinx

type Permission struct {
	Group  string `json:"group"`
	Name   string `json:"name"`
	Path   string `json:"path"`
	Method string `json:"method"`
	Key    string `json:"key"`
}

func Catalog() []Permission {
	items := []Permission{
		{Group: "仪表盘", Name: "查看仪表盘", Path: "/api/v1/dashboard/stats", Method: "GET"},
		{Group: "用户管理", Name: "查看用户", Path: "/api/v1/users", Method: "GET"},
		{Group: "用户管理", Name: "新增用户", Path: "/api/v1/users", Method: "POST"},
		{Group: "用户管理", Name: "修改用户", Path: "/api/v1/users/:id", Method: "PUT"},
		{Group: "用户管理", Name: "删除用户", Path: "/api/v1/users/:id", Method: "DELETE"},
		{Group: "角色管理", Name: "查看角色", Path: "/api/v1/roles", Method: "GET"},
		{Group: "角色管理", Name: "新增角色", Path: "/api/v1/roles", Method: "POST"},
		{Group: "角色管理", Name: "修改角色", Path: "/api/v1/roles/:id", Method: "PUT"},
		{Group: "角色管理", Name: "删除角色", Path: "/api/v1/roles/:id", Method: "DELETE"},
		{Group: "角色管理", Name: "查看权限目录", Path: "/api/v1/permissions/catalog", Method: "GET"},
		{Group: "角色管理", Name: "查看角色权限", Path: "/api/v1/roles/:id/permissions", Method: "GET"},
		{Group: "角色管理", Name: "设置角色权限", Path: "/api/v1/roles/:id/permissions", Method: "PUT"},
	}
	for i := range items {
		items[i].Key = items[i].Method + " " + items[i].Path
	}
	return items
}
