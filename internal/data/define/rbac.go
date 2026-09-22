package define

type RbacCond struct {
	UserIds      []int64
	UsernameList []string
	RoleIds      []int64
}

type UserListCond struct {
	UserIds        []int64
	Keyword        string `json:"keyword"`
	OrganizationID int64  `json:"organization_id"`
}
