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
		{Group: "系统配置", Name: "查看系统配置", Path: "/api/v1/system-config/get", Method: "POST"},
		{Group: "系统配置", Name: "修改系统配置", Path: "/api/v1/system-config/update", Method: "POST"},
		{Group: "系统配置", Name: "发送 SMTP 测试邮件", Path: "/api/v1/system-config/email/test", Method: "POST"},
		{Group: "资产管理", Name: "查看资产分组", Path: "/api/v1/assets/groups/list", Method: "POST"},
		{Group: "资产管理", Name: "新增资产分组", Path: "/api/v1/assets/groups/create", Method: "POST"},
		{Group: "资产管理", Name: "修改资产分组", Path: "/api/v1/assets/groups/update", Method: "POST"},
		{Group: "资产管理", Name: "删除资产分组", Path: "/api/v1/assets/groups/delete", Method: "POST"},
		{Group: "资产管理", Name: "查看资产凭证", Path: "/api/v1/assets/credentials/list", Method: "POST"},
		{Group: "资产管理", Name: "新增资产凭证", Path: "/api/v1/assets/credentials/create", Method: "POST"},
		{Group: "资产管理", Name: "修改资产凭证", Path: "/api/v1/assets/credentials/update", Method: "POST"},
		{Group: "资产管理", Name: "删除资产凭证", Path: "/api/v1/assets/credentials/delete", Method: "POST"},
		{Group: "资产管理", Name: "查看云账号", Path: "/api/v1/assets/cloud-accounts/list", Method: "POST"},
		{Group: "资产管理", Name: "新增云账号", Path: "/api/v1/assets/cloud-accounts/create", Method: "POST"},
		{Group: "资产管理", Name: "修改云账号", Path: "/api/v1/assets/cloud-accounts/update", Method: "POST"},
		{Group: "资产管理", Name: "删除云账号", Path: "/api/v1/assets/cloud-accounts/delete", Method: "POST"},
		{Group: "资产管理", Name: "查看主机", Path: "/api/v1/assets/hosts/list", Method: "POST"},
		{Group: "资产管理", Name: "新增主机", Path: "/api/v1/assets/hosts/create", Method: "POST"},
		{Group: "资产管理", Name: "修改主机", Path: "/api/v1/assets/hosts/update", Method: "POST"},
		{Group: "资产管理", Name: "删除主机", Path: "/api/v1/assets/hosts/delete", Method: "POST"},
		{Group: "资产管理", Name: "查看数据库实例", Path: "/api/v1/assets/db-instances/list", Method: "POST"},
		{Group: "资产管理", Name: "新增数据库实例", Path: "/api/v1/assets/db-instances/create", Method: "POST"},
		{Group: "资产管理", Name: "修改数据库实例", Path: "/api/v1/assets/db-instances/update", Method: "POST"},
		{Group: "资产管理", Name: "删除数据库实例", Path: "/api/v1/assets/db-instances/delete", Method: "POST"},
		{Group: "资产管理", Name: "连接主机终端", Path: "/api/v1/assets/terminal/connect/:id", Method: "GET"},
		{Group: "资产管理", Name: "查看终端审计", Path: "/api/v1/assets/terminal-sessions/list", Method: "POST"},
		{Group: "资产管理", Name: "查看终端录制", Path: "/api/v1/assets/terminal-sessions/recording", Method: "POST"},
		{Group: "资产管理", Name: "删除终端审计", Path: "/api/v1/assets/terminal-sessions/delete", Method: "POST"},
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
