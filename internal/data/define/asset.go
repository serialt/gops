package define

type PageCond struct {
	Page  int
	Limit int
}

type AssetGroupListCond struct {
	Page    int
	Limit   int
	Keyword string
}

type CredentialListCond struct {
	PageCond
	Keyword string
}

type CloudAccountListCond struct {
	PageCond
	Keyword string
}

type HostListCond struct {
	PageCond
	Keyword string
	GroupID int64
}

type DBInstanceListCond struct {
	PageCond
	Keyword string
	GroupID int64
}
