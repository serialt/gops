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
		{Group: "仪表盘", Name: "查看仪表盘", Path: "/api/v1/dashboard/stats", Method: "POST"},
		{Group: "用户管理", Name: "查看用户", Path: "/api/v1/users/list", Method: "POST"},
		{Group: "用户管理", Name: "新增用户", Path: "/api/v1/users/create", Method: "POST"},
		{Group: "用户管理", Name: "修改用户", Path: "/api/v1/users/update", Method: "POST"},
		{Group: "用户管理", Name: "删除用户", Path: "/api/v1/users/delete", Method: "POST"},
		{Group: "组织管理", Name: "查看组织", Path: "/api/v1/organizations/list", Method: "POST"},
		{Group: "组织管理", Name: "新增组织", Path: "/api/v1/organizations/create", Method: "POST"},
		{Group: "组织管理", Name: "修改组织", Path: "/api/v1/organizations/update", Method: "POST"},
		{Group: "组织管理", Name: "移动组织", Path: "/api/v1/organizations/move", Method: "POST"},
		{Group: "组织管理", Name: "删除组织", Path: "/api/v1/organizations/delete", Method: "POST"},
		{Group: "角色管理", Name: "查看角色", Path: "/api/v1/roles/list", Method: "POST"},
		{Group: "角色管理", Name: "新增角色", Path: "/api/v1/roles/create", Method: "POST"},
		{Group: "角色管理", Name: "修改角色", Path: "/api/v1/roles/update", Method: "POST"},
		{Group: "角色管理", Name: "删除角色", Path: "/api/v1/roles/delete", Method: "POST"},
		{Group: "角色管理", Name: "查看权限目录", Path: "/api/v1/permissions/catalog", Method: "POST"},
		{Group: "角色管理", Name: "查看角色权限", Path: "/api/v1/roles/permissions/get", Method: "POST"},
		{Group: "角色管理", Name: "设置角色权限", Path: "/api/v1/roles/permissions/set", Method: "POST"},
	}
	for i := range items {
		items[i].Key = items[i].Method + " " + items[i].Path
	}
	return items
}
