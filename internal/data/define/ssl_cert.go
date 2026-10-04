package define

type SSLCertificateListCond struct {
	PageCond
	Keyword    string
	Status     string
	SourceType string
}

type SSLDNSProviderListCond struct {
	PageCond
	Keyword  string
	Provider string
}

type SSLDeployConfigListCond struct {
	PageCond
	Keyword       string
	CertificateID int64
	DeployType    string
}

type SSLTaskListCond struct {
	PageCond
	CertificateID int64
	TaskType      string
	Status        string
	TriggerType   string
}
