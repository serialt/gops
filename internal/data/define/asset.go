package define

type PageCond struct {
	Page  int
	Limit int
}

type AssetGroupListCond struct {
	Page     int
	Limit    int
	Keyword  string
	ParentID int64
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
	Status  *int64
	IDs     []int64
}

type DBInstanceListCond struct {
	PageCond
	Keyword string
	GroupID int64
	DbType  string
	IDs     []int64
}
