package define

type IOTListCond struct {
	PageCond
	Keyword string
	Status  string
	RoleID  int64
}
