package biz

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"gops/internal/data/api"
	"gops/internal/data/define"
	"gops/internal/data/model"
	"gops/internal/data/repo"

	"gorm.io/gorm"
)

type SSLCertBiz struct {
	db         *gorm.DB
	repo       *repo.SSLCertRepo
	assets     *AssetBiz
	kubernetes *KubernetesBiz
}

func NewSSLCertBiz(db *gorm.DB, sslRepo *repo.SSLCertRepo, assets *AssetBiz, kubernetes *KubernetesBiz) *SSLCertBiz {
	return &SSLCertBiz{db: db, repo: sslRepo, assets: assets, kubernetes: kubernetes}
}

func (b *SSLCertBiz) ListCertificates(ctx context.Context, req *api.SSLCertificateListRequest) (*api.PageResponse[*api.SSLCertificateResponse], error) {
	rows, total, err := b.repo.ListCertificates(ctx, b.db, &define.SSLCertificateListCond{PageCond: define.PageCond{Page: req.Page, Limit: req.Limit}, Keyword: req.Keyword, Status: req.Status, SourceType: req.SourceType})
	if err != nil {
		return nil, err
	}
	result := make([]*api.SSLCertificateResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, certificateResponse(row))
	}
	return &api.PageResponse[*api.SSLCertificateResponse]{List: result, Total: total}, nil
}

func (b *SSLCertBiz) GetCertificate(ctx context.Context, id int64) (*api.SSLCertificateDetailResponse, error) {
	row, err := b.repo.GetCertificate(ctx, b.db, id)
	if err != nil {
		return nil, err
	}
	return certificateDetailResponse(row), nil
}

func (b *SSLCertBiz) DownloadCertificates(ctx context.Context, req *api.SSLCertificateDownloadRequest) ([]byte, error) {
	if len(req.IDs) == 0 {
		return nil, errors.New("请选择要下载的证书")
	}
	var output bytes.Buffer
	archive := zip.NewWriter(&output)
	usedNames := make(map[string]struct{}, len(req.IDs))
	for _, id := range req.IDs {
		cert, err := b.repo.GetCertificate(ctx, b.db, id)
		if err != nil {
			_ = archive.Close()
			return nil, errors.New("所选证书不存在")
		}
		if strings.TrimSpace(cert.Certificate) == "" {
			_ = archive.Close()
			return nil, fmt.Errorf("证书“%s”缺少证书内容", cert.Name)
		}
		if strings.TrimSpace(cert.PrivateKey) == "" {
			_ = archive.Close()
			return nil, fmt.Errorf("证书“%s”缺少私钥", cert.Name)
		}
		directory := nginxCertificateArchiveName(cert.Name)
		if _, exists := usedNames[directory]; exists {
			directory = fmt.Sprintf("%s-%d", directory, cert.ID)
		}
		usedNames[directory] = struct{}{}
		certificate := strings.TrimSpace(cert.Certificate) + "\n"
		chain := strings.TrimSpace(cert.CertChain)
		fullChain := certificate
		if chain != "" {
			fullChain += chain + "\n"
		}
		files := []struct {
			name string
			body string
			mode uint32
		}{{"cert.pem", certificate, 0o644}, {"chain.pem", chain + "\n", 0o644}, {"fullchain.pem", fullChain, 0o644}, {"privkey.pem", strings.TrimSpace(cert.PrivateKey) + "\n", 0o600}, {"nginx.conf.example", fmt.Sprintf("ssl_certificate     /etc/nginx/ssl/%s/fullchain.pem;\nssl_certificate_key /etc/nginx/ssl/%s/privkey.pem;\n", directory, directory), 0o644}}
		for _, file := range files {
			if file.name == "chain.pem" && chain == "" {
				continue
			}
			header := &zip.FileHeader{Name: directory + "/" + file.name, Method: zip.Deflate}
			header.SetMode(os.FileMode(file.mode))
			writer, createErr := archive.CreateHeader(header)
			if createErr != nil {
				_ = archive.Close()
				return nil, createErr
			}
			if _, writeErr := writer.Write([]byte(file.body)); writeErr != nil {
				_ = archive.Close()
				return nil, writeErr
			}
		}
	}
	if err := archive.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func nginxCertificateArchiveName(value string) string {
	name := strings.Map(func(r rune) rune {
		if r < 32 || strings.ContainsRune(`/\\:*?"<>|`, r) {
			return '-'
		}
		return r
	}, strings.TrimSpace(value))
	name = strings.Trim(name, ".- ")
	if name == "" {
		return "certificate"
	}
	return name
}

func (b *SSLCertBiz) SaveCertificate(ctx context.Context, req *api.SSLCertificateRequest) (*api.SSLCertificateDetailResponse, error) {
	row := &model.SslCertificate{ID: req.ID, Name: strings.TrimSpace(req.Name), Domain: strings.TrimSpace(req.Domain), SanDomains: req.SanDomains, AcmeEmail: req.AcmeEmail, CaProvider: defaultString(req.CaProvider, "letsencrypt"), KeyAlgorithm: defaultString(req.KeyAlgorithm, "rsa2048"), SourceType: defaultString(req.SourceType, "manual"), CloudAccountID: req.CloudAccountID, DNSProviderID: req.DNSProviderID, Certificate: req.Certificate, PrivateKey: req.PrivateKey, CertChain: req.CertChain, AutoRenew: req.AutoRenew, RenewDaysBefore: req.RenewDaysBefore, Status: "pending"}
	if row.RenewDaysBefore <= 0 {
		row.RenewDaysBefore = 30
	}
	if req.ID > 0 {
		old, err := b.repo.GetCertificate(ctx, b.db, req.ID)
		if err != nil {
			return nil, err
		}
		row.CreatedAt, row.Status, row.Issuer, row.NotBefore, row.NotAfter, row.Fingerprint, row.LastRenewAt, row.LastError = old.CreatedAt, old.Status, old.Issuer, old.NotBefore, old.NotAfter, old.Fingerprint, old.LastRenewAt, old.LastError
		if row.Certificate == "" {
			row.Certificate = old.Certificate
		}
		if row.PrivateKey == "" {
			row.PrivateKey = old.PrivateKey
		}
		if row.CertChain == "" {
			row.CertChain = old.CertChain
		}
	}
	if row.Certificate != "" {
		if leaf, bundledChain := splitCertificateBundle(row.Certificate); bundledChain != "" {
			row.Certificate = leaf
			if strings.TrimSpace(row.CertChain) == "" {
				row.CertChain = bundledChain
			}
		}
		if err := populateCertificate(row); err != nil {
			return nil, err
		}
	}
	if req.ID > 0 {
		if err := b.repo.SaveCertificate(ctx, b.db, row); err != nil {
			return nil, err
		}
	} else {
		if err := b.repo.CreateCertificate(ctx, b.db, row); err != nil {
			return nil, err
		}
	}
	return certificateDetailResponse(row), nil
}

func (b *SSLCertBiz) DeleteCertificate(ctx context.Context, id int64) error {
	row, err := b.repo.GetCertificate(ctx, b.db, id)
	if err != nil {
		return err
	}
	return b.repo.DeleteCertificate(ctx, b.db, row)
}
func (b *SSLCertBiz) RenewCertificate(ctx context.Context, id int64) (*api.SSLTaskResponse, error) {
	cert, err := b.repo.GetCertificate(ctx, b.db, id)
	if err != nil {
		return nil, err
	}
	if cert.SourceType != "acme" {
		return nil, errors.New("仅 ACME 证书支持自动续期")
	}
	now := time.Now()
	task := &model.SslTask{CertificateID: id, TaskType: "renew", Status: "pending", TriggerType: "manual", CreatedAt: now, UpdatedAt: now}
	if err := b.repo.CreateTask(ctx, b.db, task); err != nil {
		return nil, err
	}
	go b.runACMETask(task.ID, cert.ID)
	return taskResponse(task, cert.Name), nil
}

func (b *SSLCertBiz) ApplyCertificate(ctx context.Context, req *api.SSLCertificateApplyRequest) (*api.SSLTaskResponse, error) {
	provider, err := b.repo.GetDNSProvider(ctx, b.db, req.DNSProviderID)
	if err != nil {
		return nil, errors.New("DNS 配置不存在")
	}
	if !provider.Enabled {
		return nil, errors.New("DNS 配置未启用")
	}
	if _, err = newLegoDNSProvider(provider); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Name)
	domain := strings.TrimSpace(req.Domain)
	if name == "" || domain == "" {
		return nil, errors.New("证书名称和主域名不能为空")
	}
	if req.RenewDaysBefore <= 0 {
		req.RenewDaysBefore = 30
	}
	now := time.Now()
	cert := &model.SslCertificate{Name: name, Domain: domain, SanDomains: strings.TrimSpace(req.SanDomains), AcmeEmail: strings.TrimSpace(req.AcmeEmail), CaProvider: defaultString(req.CAProvider, "letsencrypt"), KeyAlgorithm: defaultString(req.KeyAlgorithm, "rsa2048"), SourceType: "acme", DNSProviderID: req.DNSProviderID, AutoRenew: req.AutoRenew, RenewDaysBefore: req.RenewDaysBefore, Status: "pending", CreatedAt: now, UpdatedAt: now}
	task := &model.SslTask{TaskType: "issue", Status: "pending", TriggerType: "manual", CreatedAt: now, UpdatedAt: now}
	err = b.db.Transaction(func(tx *gorm.DB) error {
		if err := b.repo.CreateCertificate(ctx, tx, cert); err != nil {
			return err
		}
		task.CertificateID = cert.ID
		return b.repo.CreateTask(ctx, tx, task)
	})
	if err != nil {
		return nil, err
	}
	go b.runACMETask(task.ID, cert.ID)
	return taskResponse(task, cert.Name), nil
}

func (b *SSLCertBiz) ListDNSProviders(ctx context.Context, req *api.SSLDNSProviderListRequest) (*api.PageResponse[*api.SSLDNSProviderResponse], error) {
	rows, total, err := b.repo.ListDNSProviders(ctx, b.db, &define.SSLDNSProviderListCond{PageCond: define.PageCond{Page: req.Page, Limit: req.Limit}, Keyword: req.Keyword, Provider: req.Provider})
	if err != nil {
		return nil, err
	}
	result := make([]*api.SSLDNSProviderResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, dnsResponse(row, false))
	}
	return &api.PageResponse[*api.SSLDNSProviderResponse]{List: result, Total: total}, nil
}
func (b *SSLCertBiz) GetDNSProvider(ctx context.Context, id int64) (*api.SSLDNSProviderResponse, error) {
	row, err := b.repo.GetDNSProvider(ctx, b.db, id)
	if err != nil {
		return nil, err
	}
	return dnsResponse(row, true), nil
}
func (b *SSLCertBiz) SaveDNSProvider(ctx context.Context, req *api.SSLDNSProviderRequest) (*api.SSLDNSProviderResponse, error) {
	if !json.Valid([]byte(req.Config)) {
		return nil, errors.New("DNS 配置必须是有效 JSON")
	}
	row := &model.SslDNSProvider{ID: req.ID, Name: req.Name, Provider: req.Provider, Config: req.Config, Email: req.Email, Phone: req.Phone, Enabled: req.Enabled}
	if req.ID > 0 {
		old, err := b.repo.GetDNSProvider(ctx, b.db, req.ID)
		if err != nil {
			return nil, err
		}
		row.CreatedAt, row.LastTestAt, row.LastTestOk = old.CreatedAt, old.LastTestAt, old.LastTestOk
		if err = b.repo.SaveDNSProvider(ctx, b.db, row); err != nil {
			return nil, err
		}
	} else if err := b.repo.CreateDNSProvider(ctx, b.db, row); err != nil {
		return nil, err
	}
	return dnsResponse(row, true), nil
}
func (b *SSLCertBiz) DeleteDNSProvider(ctx context.Context, id int64) error {
	row, err := b.repo.GetDNSProvider(ctx, b.db, id)
	if err != nil {
		return err
	}
	return b.repo.DeleteDNSProvider(ctx, b.db, row)
}
func (b *SSLCertBiz) TestDNSProvider(ctx context.Context, id int64) error {
	row, err := b.repo.GetDNSProvider(ctx, b.db, id)
	if err != nil {
		return err
	}
	if _, err = newLegoDNSProvider(row); err != nil {
		return err
	}
	row.LastTestAt, row.LastTestOk = time.Now(), true
	return b.repo.SaveDNSProvider(ctx, b.db, row)
}

func (b *SSLCertBiz) ListDeployConfigs(ctx context.Context, req *api.SSLDeployConfigListRequest) (*api.PageResponse[*api.SSLDeployConfigResponse], error) {
	rows, total, err := b.repo.ListDeployConfigs(ctx, b.db, &define.SSLDeployConfigListCond{PageCond: define.PageCond{Page: req.Page, Limit: req.Limit}, Keyword: req.Keyword, CertificateID: req.CertificateID, DeployType: req.DeployType})
	if err != nil {
		return nil, err
	}
	result := make([]*api.SSLDeployConfigResponse, 0, len(rows))
	for _, row := range rows {
		cert, _ := b.repo.GetCertificate(ctx, b.db, row.CertificateID)
		name := ""
		if cert != nil {
			name = cert.Name
		}
		result = append(result, deployResponse(row, name))
	}
	return &api.PageResponse[*api.SSLDeployConfigResponse]{List: result, Total: total}, nil
}
func (b *SSLCertBiz) GetDeployConfig(ctx context.Context, id int64) (*api.SSLDeployConfigResponse, error) {
	row, err := b.repo.GetDeployConfig(ctx, b.db, id)
	if err != nil {
		return nil, err
	}
	cert, _ := b.repo.GetCertificate(ctx, b.db, row.CertificateID)
	name := ""
	if cert != nil {
		name = cert.Name
	}
	return deployResponse(row, name), nil
}
func (b *SSLCertBiz) SaveDeployConfig(ctx context.Context, req *api.SSLDeployConfigRequest) (*api.SSLDeployConfigResponse, error) {
	if !json.Valid([]byte(req.TargetConfig)) {
		return nil, errors.New("目标配置必须是有效 JSON")
	}
	cert, err := b.repo.GetCertificate(ctx, b.db, req.CertificateID)
	if err != nil {
		return nil, errors.New("关联证书不存在")
	}
	row := &model.SslDeployConfig{ID: req.ID, CertificateID: req.CertificateID, Name: req.Name, DeployType: req.DeployType, TargetConfig: req.TargetConfig, AutoDeploy: req.AutoDeploy, Enabled: req.Enabled}
	if req.ID > 0 {
		old, e := b.repo.GetDeployConfig(ctx, b.db, req.ID)
		if e != nil {
			return nil, e
		}
		row.CreatedAt, row.LastDeployAt, row.LastDeployOk, row.LastError = old.CreatedAt, old.LastDeployAt, old.LastDeployOk, old.LastError
		err = b.repo.SaveDeployConfig(ctx, b.db, row)
	} else {
		err = b.repo.CreateDeployConfig(ctx, b.db, row)
	}
	if err != nil {
		return nil, err
	}
	return deployResponse(row, cert.Name), nil
}
func (b *SSLCertBiz) DeleteDeployConfig(ctx context.Context, id int64) error {
	row, err := b.repo.GetDeployConfig(ctx, b.db, id)
	if err != nil {
		return err
	}
	return b.repo.DeleteDeployConfig(ctx, b.db, row)
}
func (b *SSLCertBiz) TestDeployConfig(ctx context.Context, id int64) error {
	row, err := b.repo.GetDeployConfig(ctx, b.db, id)
	if err != nil {
		return err
	}
	return b.testDeployTarget(ctx, row)
}
func (b *SSLCertBiz) ExecuteDeploy(ctx context.Context, id int64) (*api.SSLTaskResponse, error) {
	row, err := b.repo.GetDeployConfig(ctx, b.db, id)
	if err != nil {
		return nil, err
	}
	cert, err := b.repo.GetCertificate(ctx, b.db, row.CertificateID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	task := &model.SslTask{CertificateID: cert.ID, TaskType: "deploy", Status: "pending", TriggerType: "manual", Result: `{"deploy_config_id":` + jsonNumber(id) + `}`, CreatedAt: now, UpdatedAt: now}
	if err = b.repo.CreateTask(ctx, b.db, task); err != nil {
		return nil, err
	}
	go b.runDeployTask(task.ID, row.ID)
	return taskResponse(task, cert.Name), nil
}

func (b *SSLCertBiz) ListTasks(ctx context.Context, req *api.SSLTaskListRequest) (*api.PageResponse[*api.SSLTaskResponse], error) {
	rows, total, err := b.repo.ListTasks(ctx, b.db, &define.SSLTaskListCond{PageCond: define.PageCond{Page: req.Page, Limit: req.Limit}, CertificateID: req.CertificateID, TaskType: req.TaskType, Status: req.Status, TriggerType: req.TriggerType})
	if err != nil {
		return nil, err
	}
	result := make([]*api.SSLTaskResponse, 0, len(rows))
	for _, row := range rows {
		cert, _ := b.repo.GetCertificate(ctx, b.db, row.CertificateID)
		name := ""
		if cert != nil {
			name = cert.Name
		}
		result = append(result, taskResponse(row, name))
	}
	return &api.PageResponse[*api.SSLTaskResponse]{List: result, Total: total}, nil
}
func (b *SSLCertBiz) GetTask(ctx context.Context, id int64) (*api.SSLTaskResponse, error) {
	row, err := b.repo.GetTask(ctx, b.db, id)
	if err != nil {
		return nil, err
	}
	cert, _ := b.repo.GetCertificate(ctx, b.db, row.CertificateID)
	name := ""
	if cert != nil {
		name = cert.Name
	}
	return taskResponse(row, name), nil
}

func populateCertificate(row *model.SslCertificate) error {
	block, _ := pem.Decode([]byte(row.Certificate))
	if block == nil {
		return errors.New("无法解析 PEM 证书")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return errors.New("证书内容无效")
	}
	sum := sha256.Sum256(cert.Raw)
	row.Issuer, row.NotBefore, row.NotAfter, row.Fingerprint = cert.Issuer.String(), cert.NotBefore, cert.NotAfter, strings.ToUpper(hex.EncodeToString(sum[:]))
	row.Status = "active"
	if time.Now().After(cert.NotAfter) {
		row.Status = "expired"
	} else if time.Until(cert.NotAfter) <= time.Duration(row.RenewDaysBefore)*24*time.Hour {
		row.Status = "expiring"
	}
	return nil
}
func certificateResponse(x *model.SslCertificate) *api.SSLCertificateResponse {
	return &api.SSLCertificateResponse{ID: x.ID, Name: x.Name, Domain: x.Domain, SanDomains: x.SanDomains, AcmeEmail: x.AcmeEmail, CaProvider: x.CaProvider, KeyAlgorithm: x.KeyAlgorithm, SourceType: x.SourceType, CloudAccountID: x.CloudAccountID, DNSProviderID: x.DNSProviderID, Issuer: x.Issuer, Fingerprint: x.Fingerprint, Status: x.Status, AutoRenew: x.AutoRenew, RenewDaysBefore: x.RenewDaysBefore, NotBefore: nonZeroTime(x.NotBefore), NotAfter: nonZeroTime(x.NotAfter), LastRenewAt: nonZeroTime(x.LastRenewAt), LastError: x.LastError, CreatedAt: x.CreatedAt, UpdatedAt: x.UpdatedAt}
}
func certificateDetailResponse(x *model.SslCertificate) *api.SSLCertificateDetailResponse {
	_, bundledChain := splitCertificateBundle(x.Certificate)
	displayChain := strings.TrimSpace(x.CertChain)
	if displayChain == "" {
		displayChain = bundledChain
	}
	chainPEM := strings.TrimSpace(x.Certificate + "\n" + displayChain)
	chain, err := parseCertificateChain(chainPEM)
	response := &api.SSLCertificateDetailResponse{
		SSLCertificateResponse: certificateResponse(x),
		Certificate:            x.Certificate,
		PrivateKey:             x.PrivateKey,
		CertChain:              displayChain,
		ChainCertificates:      chain}
	if err != nil {
		response.ChainParseError = err.Error()
	}
	return response
}

func splitCertificateBundle(value string) (string, string) {
	remaining := []byte(strings.TrimSpace(value))
	certificates := make([]string, 0)
	for len(remaining) > 0 {
		block, rest := pem.Decode(remaining)
		if block == nil {
			break
		}
		remaining = rest
		if block.Type == "CERTIFICATE" {
			certificates = append(certificates, strings.TrimSpace(string(pem.EncodeToMemory(block))))
		}
	}
	if len(certificates) == 0 {
		return strings.TrimSpace(value), ""
	}
	if len(certificates) == 1 {
		return certificates[0], ""
	}
	return certificates[0], strings.Join(certificates[1:], "\n")
}

func parseCertificateChain(value string) ([]*api.SSLCertificateChainResponse, error) {
	remaining := []byte(strings.TrimSpace(value))
	result := make([]*api.SSLCertificateChainResponse, 0)
	seen := make(map[string]struct{})
	for len(remaining) > 0 {
		block, rest := pem.Decode(remaining)
		if block == nil {
			return result, errors.New("证书链包含无法解析的 PEM 内容")
		}
		remaining = rest
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return result, errors.New("证书链包含无效的 X.509 证书")
		}
		sum := sha256.Sum256(cert.Raw)
		fingerprint := strings.ToUpper(hex.EncodeToString(sum[:]))
		if _, ok := seen[fingerprint]; ok {
			continue
		}
		seen[fingerprint] = struct{}{}
		result = append(result, &api.SSLCertificateChainResponse{
			Position:     len(result) + 1,
			Version:      cert.Version,
			CommonName:   cert.Subject.CommonName,
			Subject:      cert.Subject.String(),
			Issuer:       cert.Issuer.String(),
			SerialNumber: strings.ToUpper(cert.SerialNumber.Text(16)),
			DNSNames:     cert.DNSNames, EmailAddresses: cert.EmailAddresses, IPAddresses: certificateIPAddresses(cert), URIs: certificateURIs(cert), NotBefore: nonZeroTime(cert.NotBefore), NotAfter: nonZeroTime(cert.NotAfter), Fingerprint: fingerprint, IsCA: cert.IsCA, BasicConstraints: cert.BasicConstraintsValid, MaxPathLen: cert.MaxPathLen, MaxPathLenZero: cert.MaxPathLenZero, SignatureAlgorithm: cert.SignatureAlgorithm.String(), PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(), PublicKeyBits: certificatePublicKeyBits(cert), KeyUsage: certificateKeyUsages(cert.KeyUsage), ExtKeyUsage: certificateExtKeyUsages(cert.ExtKeyUsage), SubjectKeyID: strings.ToUpper(hex.EncodeToString(cert.SubjectKeyId)), AuthorityKeyID: strings.ToUpper(hex.EncodeToString(cert.AuthorityKeyId)), OCSPServers: cert.OCSPServer, IssuerURLs: cert.IssuingCertificateURL, CRLPoints: cert.CRLDistributionPoints, Policies: certificatePolicies(cert)})
	}
	if strings.TrimSpace(value) != "" && len(result) == 0 {
		return result, errors.New("证书链中未找到 X.509 证书")
	}
	return result, nil
}

func certificatePublicKeyBits(cert *x509.Certificate) int {
	switch key := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		return key.N.BitLen()
	case *ecdsa.PublicKey:
		return key.Curve.Params().BitSize
	case ed25519.PublicKey:
		return len(key) * 8
	default:
		return 0
	}
}

func certificateIPAddresses(cert *x509.Certificate) []string {
	values := make([]string, 0, len(cert.IPAddresses))
	for _, value := range cert.IPAddresses {
		values = append(values, value.String())
	}
	return values
}

func certificateURIs(cert *x509.Certificate) []string {
	values := make([]string, 0, len(cert.URIs))
	for _, value := range cert.URIs {
		values = append(values, value.String())
	}
	return values
}

func certificatePolicies(cert *x509.Certificate) []string {
	values := make([]string, 0, len(cert.PolicyIdentifiers))
	for _, value := range cert.PolicyIdentifiers {
		values = append(values, value.String())
	}
	return values
}

func certificateKeyUsages(usage x509.KeyUsage) []string {
	items := []struct {
		value x509.KeyUsage
		label string
	}{{x509.KeyUsageDigitalSignature, "数字签名"}, {x509.KeyUsageContentCommitment, "内容承诺"}, {x509.KeyUsageKeyEncipherment, "密钥加密"}, {x509.KeyUsageDataEncipherment, "数据加密"}, {x509.KeyUsageKeyAgreement, "密钥协商"}, {x509.KeyUsageCertSign, "证书签名"}, {x509.KeyUsageCRLSign, "CRL 签名"}, {x509.KeyUsageEncipherOnly, "仅加密"}, {x509.KeyUsageDecipherOnly, "仅解密"}}
	values := make([]string, 0)
	for _, item := range items {
		if usage&item.value != 0 {
			values = append(values, item.label)
		}
	}
	return values
}

func certificateExtKeyUsages(usages []x509.ExtKeyUsage) []string {
	labels := map[x509.ExtKeyUsage]string{x509.ExtKeyUsageAny: "任意", x509.ExtKeyUsageServerAuth: "服务器认证", x509.ExtKeyUsageClientAuth: "客户端认证", x509.ExtKeyUsageCodeSigning: "代码签名", x509.ExtKeyUsageEmailProtection: "邮件保护", x509.ExtKeyUsageIPSECEndSystem: "IPsec 终端", x509.ExtKeyUsageIPSECTunnel: "IPsec 隧道", x509.ExtKeyUsageIPSECUser: "IPsec 用户", x509.ExtKeyUsageTimeStamping: "时间戳", x509.ExtKeyUsageOCSPSigning: "OCSP 签名", x509.ExtKeyUsageMicrosoftServerGatedCrypto: "Microsoft SGC", x509.ExtKeyUsageNetscapeServerGatedCrypto: "Netscape SGC", x509.ExtKeyUsageMicrosoftCommercialCodeSigning: "Microsoft 商业代码签名", x509.ExtKeyUsageMicrosoftKernelCodeSigning: "Microsoft 内核代码签名"}
	values := make([]string, 0, len(usages))
	for _, usage := range usages {
		if label, ok := labels[usage]; ok {
			values = append(values, label)
		} else {
			values = append(values, "其他")
		}
	}
	return values
}
func dnsResponse(x *model.SslDNSProvider, detail bool) *api.SSLDNSProviderResponse {
	config := ""
	if detail {
		config = x.Config
	}
	return &api.SSLDNSProviderResponse{ID: x.ID, Name: x.Name, Provider: x.Provider, Config: config, Email: x.Email, Phone: x.Phone, Enabled: x.Enabled, LastTestAt: nonZeroTime(x.LastTestAt), LastTestOk: x.LastTestOk, CreatedAt: x.CreatedAt, UpdatedAt: x.UpdatedAt}
}
func deployResponse(x *model.SslDeployConfig, certName string) *api.SSLDeployConfigResponse {
	return &api.SSLDeployConfigResponse{ID: x.ID, CertificateID: x.CertificateID, CertificateName: certName, Name: x.Name, DeployType: x.DeployType, TargetConfig: x.TargetConfig, AutoDeploy: x.AutoDeploy, Enabled: x.Enabled, LastDeployAt: nonZeroTime(x.LastDeployAt), LastDeployOk: x.LastDeployOk, LastError: x.LastError, CreatedAt: x.CreatedAt, UpdatedAt: x.UpdatedAt}
}
func taskResponse(x *model.SslTask, certName string) *api.SSLTaskResponse {
	return &api.SSLTaskResponse{ID: x.ID, CertificateID: x.CertificateID, CertificateName: certName, TaskType: x.TaskType, Status: x.Status, TriggerType: x.TriggerType, StartedAt: nonZeroTime(x.StartedAt), FinishedAt: nonZeroTime(x.FinishedAt), ErrorMessage: x.ErrorMessage, Result: x.Result, CreatedAt: x.CreatedAt}
}
func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
func jsonNumber(value int64) string { data, _ := json.Marshal(value); return string(data) }

func nonZeroTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}
