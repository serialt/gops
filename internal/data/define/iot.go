package define

type IOTListCond struct {
	PageCond
	Keyword string
	Status  string
	RoleID  int64
}

type IOTBrokerListCond struct {
	PageCond
	Keyword string
	Enabled *bool
}
