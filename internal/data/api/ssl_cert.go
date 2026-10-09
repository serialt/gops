package api

import "time"

type SSLCertificateListRequest struct {
	PageRequest
	Keyword    string `json:"keyword"`
	Status     string `json:"status"`
	SourceType string `json:"source_type"`
}
type SSLCertificateIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}
type SSLCertificateDownloadRequest struct {
	IDs []int64 `json:"ids" binding:"required,min=1"`
}
type SSLCertificateApplyRequest struct {
	Name            string `json:"name" binding:"required"`
	Domain          string `json:"domain" binding:"required"`
	SanDomains      string `json:"san_domains"`
	AcmeEmail       string `json:"acme_email" binding:"required,email"`
	DNSProviderID   int64  `json:"dns_provider_id" binding:"required"`
	KeyAlgorithm    string `json:"key_algorithm"`
	CAProvider      string `json:"ca_provider"`
	AutoRenew       bool   `json:"auto_renew"`
	RenewDaysBefore int64  `json:"renew_days_before"`
}
type SSLCertificateRequest struct {
	ID              int64  `json:"id"`
	Name            string `json:"name" binding:"required"`
	Domain          string `json:"domain" binding:"required"`
	SanDomains      string `json:"san_domains"`
	AcmeEmail       string `json:"acme_email"`
	CaProvider      string `json:"ca_provider"`
	KeyAlgorithm    string `json:"key_algorithm"`
	SourceType      string `json:"source_type"`
	CloudAccountID  int64  `json:"cloud_account_id"`
	Certificate     string `json:"certificate"`
	PrivateKey      string `json:"private_key"`
	CertChain       string `json:"cert_chain"`
	AutoRenew       bool   `json:"auto_renew"`
	RenewDaysBefore int64  `json:"renew_days_before"`
	DNSProviderID   int64  `json:"dns_provider_id"`
}
type SSLCertificateResponse struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Domain          string     `json:"domain"`
	SanDomains      string     `json:"san_domains"`
	AcmeEmail       string     `json:"acme_email"`
	CaProvider      string     `json:"ca_provider"`
	KeyAlgorithm    string     `json:"key_algorithm"`
	SourceType      string     `json:"source_type"`
	CloudAccountID  int64      `json:"cloud_account_id"`
	DNSProviderID   int64      `json:"dns_provider_id"`
	Issuer          string     `json:"issuer"`
	Fingerprint     string     `json:"fingerprint"`
	Status          string     `json:"status"`
	AutoRenew       bool       `json:"auto_renew"`
	RenewDaysBefore int64      `json:"renew_days_before"`
	NotBefore       *time.Time `json:"not_before"`
	NotAfter        *time.Time `json:"not_after"`
	LastRenewAt     *time.Time `json:"last_renew_at"`
	LastError       string     `json:"last_error"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
type SSLCertificateDetailResponse struct {
	*SSLCertificateResponse
	Certificate       string                         `json:"certificate"`
	PrivateKey        string                         `json:"private_key"`
	CertChain         string                         `json:"cert_chain"`
	ChainCertificates []*SSLCertificateChainResponse `json:"chain_certificates"`
	ChainParseError   string                         `json:"chain_parse_error"`
}

type SSLCertificateChainResponse struct {
	Position           int        `json:"position"`
	Version            int        `json:"version"`
	Subject            string     `json:"subject"`
	Issuer             string     `json:"issuer"`
	SerialNumber       string     `json:"serial_number"`
	DNSNames           []string   `json:"dns_names"`
	EmailAddresses     []string   `json:"email_addresses"`
	IPAddresses        []string   `json:"ip_addresses"`
	URIs               []string   `json:"uris"`
	NotBefore          *time.Time `json:"not_before"`
	NotAfter           *time.Time `json:"not_after"`
	Fingerprint        string     `json:"fingerprint"`
	IsCA               bool       `json:"is_ca"`
	BasicConstraints   bool       `json:"basic_constraints_valid"`
	MaxPathLen         int        `json:"max_path_len"`
	MaxPathLenZero     bool       `json:"max_path_len_zero"`
	SignatureAlgorithm string     `json:"signature_algorithm"`
	PublicKeyAlgorithm string     `json:"public_key_algorithm"`
	PublicKeyBits      int        `json:"public_key_bits"`
	KeyUsage           []string   `json:"key_usage"`
	ExtKeyUsage        []string   `json:"ext_key_usage"`
	SubjectKeyID       string     `json:"subject_key_id"`
	AuthorityKeyID     string     `json:"authority_key_id"`
	OCSPServers        []string   `json:"ocsp_servers"`
	IssuerURLs         []string   `json:"issuer_urls"`
	CRLPoints          []string   `json:"crl_points"`
	Policies           []string   `json:"policies"`
}

type SSLDNSProviderListRequest struct {
	PageRequest
	Keyword  string `json:"keyword"`
	Provider string `json:"provider"`
}
type SSLDNSProviderRequest struct {
	ID       int64  `json:"id"`
	Name     string `json:"name" binding:"required"`
	Provider string `json:"provider" binding:"required"`
	Config   string `json:"config" binding:"required"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Enabled  bool   `json:"enabled"`
}
type SSLDNSProviderResponse struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Provider   string     `json:"provider"`
	Config     string     `json:"config,omitempty"`
	Email      string     `json:"email"`
	Phone      string     `json:"phone"`
	Enabled    bool       `json:"enabled"`
	LastTestAt *time.Time `json:"last_test_at"`
	LastTestOk bool       `json:"last_test_ok"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type SSLDeployConfigListRequest struct {
	PageRequest
	Keyword       string `json:"keyword"`
	CertificateID int64  `json:"certificate_id"`
	DeployType    string `json:"deploy_type"`
}
type SSLDeployConfigRequest struct {
	ID            int64  `json:"id"`
	CertificateID int64  `json:"certificate_id" binding:"required"`
	Name          string `json:"name" binding:"required"`
	DeployType    string `json:"deploy_type" binding:"required"`
	TargetConfig  string `json:"target_config" binding:"required"`
	AutoDeploy    bool   `json:"auto_deploy"`
	Enabled       bool   `json:"enabled"`
}
type SSLDeployConfigResponse struct {
	ID              int64      `json:"id"`
	CertificateID   int64      `json:"certificate_id"`
	CertificateName string     `json:"certificate_name"`
	Name            string     `json:"name"`
	DeployType      string     `json:"deploy_type"`
	TargetConfig    string     `json:"target_config"`
	AutoDeploy      bool       `json:"auto_deploy"`
	Enabled         bool       `json:"enabled"`
	LastDeployAt    *time.Time `json:"last_deploy_at"`
	LastDeployOk    bool       `json:"last_deploy_ok"`
	LastError       string     `json:"last_error"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type SSLTaskListRequest struct {
	PageRequest
	CertificateID int64  `json:"certificate_id"`
	TaskType      string `json:"task_type"`
	Status        string `json:"status"`
	TriggerType   string `json:"trigger_type"`
}
type SSLTaskResponse struct {
	ID              int64      `json:"id"`
	CertificateID   int64      `json:"certificate_id"`
	CertificateName string     `json:"certificate_name"`
	TaskType        string     `json:"task_type"`
	Status          string     `json:"status"`
	TriggerType     string     `json:"trigger_type"`
	StartedAt       *time.Time `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at"`
	ErrorMessage    string     `json:"error_message"`
	Result          string     `json:"result"`
	CreatedAt       time.Time  `json:"created_at"`
}
