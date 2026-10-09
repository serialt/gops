package biz

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gops/internal/data/model"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/providers/dns/huaweicloud"
	"github.com/go-acme/lego/v4/providers/dns/tencentcloud"
	"github.com/go-acme/lego/v4/registration"
)

type legoUser struct {
	email        string
	registration *registration.Resource
	privateKey   crypto.PrivateKey
}

func (u *legoUser) GetEmail() string                        { return u.email }
func (u *legoUser) GetRegistration() *registration.Resource { return u.registration }
func (u *legoUser) GetPrivateKey() crypto.PrivateKey        { return u.privateKey }

type aliDNSConfig struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	SecurityToken   string `json:"security_token"`
	Region          string `json:"region"`
}

type tencentCloudDNSConfig struct {
	SecretID     string `json:"secret_id"`
	SecretKey    string `json:"secret_key"`
	SessionToken string `json:"session_token"`
	Region       string `json:"region"`
}

type cloudflareDNSConfig struct {
	APIToken  string `json:"api_token"`
	ZoneToken string `json:"zone_token"`
	Email     string `json:"email"`
	APIKey    string `json:"api_key"`
}

type huaweiCloudDNSConfig struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Region          string `json:"region"`
}

func newLegoDNSProvider(row *model.SslDNSProvider) (challenge.Provider, error) {
	if row == nil || strings.TrimSpace(row.Config) == "" || strings.TrimSpace(row.Config) == "{}" {
		return nil, errors.New("DNS 配置不完整")
	}
	switch strings.ToLower(strings.TrimSpace(row.Provider)) {
	case "alidns", "aliyun":
		var value aliDNSConfig
		if err := json.Unmarshal([]byte(row.Config), &value); err != nil {
			return nil, fmt.Errorf("解析 AliDNS 配置: %w", err)
		}
		config := alidns.NewDefaultConfig()
		config.APIKey, config.SecretKey, config.SecurityToken = value.AccessKeyID, value.AccessKeySecret, value.SecurityToken
		if value.Region != "" {
			config.RegionID = value.Region
		}
		provider, err := alidns.NewDNSProviderConfig(config)
		return provider, wrapDNSProviderError("AliDNS", err)
	case "tencentcloud":
		var value tencentCloudDNSConfig
		if err := json.Unmarshal([]byte(row.Config), &value); err != nil {
			return nil, fmt.Errorf("解析 Tencent Cloud DNS 配置: %w", err)
		}
		config := tencentcloud.NewDefaultConfig()
		config.SecretID, config.SecretKey, config.SessionToken, config.Region = value.SecretID, value.SecretKey, value.SessionToken, value.Region
		provider, err := tencentcloud.NewDNSProviderConfig(config)
		return provider, wrapDNSProviderError("Tencent Cloud DNS", err)
	case "cloudflare":
		var value cloudflareDNSConfig
		if err := json.Unmarshal([]byte(row.Config), &value); err != nil {
			return nil, fmt.Errorf("解析 Cloudflare 配置: %w", err)
		}
		config := cloudflare.NewDefaultConfig()
		config.AuthToken, config.ZoneToken, config.AuthEmail, config.AuthKey = value.APIToken, value.ZoneToken, value.Email, value.APIKey
		provider, err := cloudflare.NewDNSProviderConfig(config)
		return provider, wrapDNSProviderError("Cloudflare", err)
	case "huaweicloud", "huawei":
		var value huaweiCloudDNSConfig
		if err := json.Unmarshal([]byte(row.Config), &value); err != nil {
			return nil, fmt.Errorf("解析 Huawei Cloud DNS 配置: %w", err)
		}
		config := huaweicloud.NewDefaultConfig()
		config.AccessKeyID, config.SecretAccessKey, config.Region = value.AccessKeyID, value.SecretAccessKey, value.Region
		provider, err := huaweicloud.NewDNSProviderConfig(config)
		return provider, wrapDNSProviderError("Huawei Cloud DNS", err)
	default:
		return nil, fmt.Errorf("不支持的 DNS 服务商 %q", row.Provider)
	}
}

func wrapDNSProviderError(provider string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s 配置无效: %w", provider, err)
}

func (b *SSLCertBiz) runACMETask(taskID, certificateID int64) {
	ctx := context.Background()
	task, err := b.repo.GetTask(ctx, b.db, taskID)
	if err != nil {
		return
	}
	cert, err := b.repo.GetCertificate(ctx, b.db, certificateID)
	if err != nil {
		b.finishACMETask(ctx, task, nil, nil, err)
		return
	}
	startedAt := time.Now()
	task.Status, task.StartedAt, task.UpdatedAt = "running", startedAt, startedAt
	_ = b.repo.SaveTask(ctx, b.db, task)

	resource, issueErr := b.obtainCertificate(ctx, cert)
	b.finishACMETask(ctx, task, cert, resource, issueErr)
}

func (b *SSLCertBiz) obtainCertificate(ctx context.Context, cert *model.SslCertificate) (*certificate.Resource, error) {
	if strings.TrimSpace(cert.AcmeEmail) == "" {
		return nil, errors.New("ACME 邮箱不能为空")
	}
	providerRow, err := b.repo.GetDNSProvider(ctx, b.db, cert.DNSProviderID)
	if err != nil {
		return nil, errors.New("DNS 配置不存在")
	}
	if !providerRow.Enabled {
		return nil, errors.New("DNS 配置未启用")
	}
	dnsProvider, err := newLegoDNSProvider(providerRow)
	if err != nil {
		return nil, err
	}
	accountKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("生成 ACME 账户密钥: %w", err)
	}
	user := &legoUser{email: cert.AcmeEmail, privateKey: accountKey}
	config := lego.NewConfig(user)
	config.CADirURL = lego.LEDirectoryProduction
	if strings.EqualFold(cert.CaProvider, "letsencrypt_staging") {
		config.CADirURL = lego.LEDirectoryStaging
	} else if !strings.EqualFold(cert.CaProvider, "letsencrypt") {
		return nil, fmt.Errorf("暂不支持的 ACME CA %q", cert.CaProvider)
	}
	config.Certificate.KeyType = legoKeyType(cert.KeyAlgorithm)
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("创建 ACME 客户端: %w", err)
	}
	if err = client.Challenge.SetDNS01Provider(dnsProvider); err != nil {
		return nil, fmt.Errorf("配置 DNS-01: %w", err)
	}
	registrationResource, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("注册 ACME 账户: %w", err)
	}
	user.registration = registrationResource

	domains := certificateDomains(cert.Domain, cert.SanDomains)
	if len(domains) == 0 {
		return nil, errors.New("至少需要一个有效域名")
	}
	var privateKey crypto.PrivateKey
	if strings.TrimSpace(cert.PrivateKey) != "" {
		privateKey, err = certcrypto.ParsePEMPrivateKey([]byte(cert.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("解析现有证书私钥: %w", err)
		}
	}
	resource, err := client.Certificate.Obtain(certificate.ObtainRequest{Domains: domains, PrivateKey: privateKey, Bundle: false})
	if err != nil {
		return nil, fmt.Errorf("申请 ACME 证书: %w", err)
	}
	return resource, nil
}

func (b *SSLCertBiz) finishACMETask(ctx context.Context, task *model.SslTask, cert *model.SslCertificate, resource *certificate.Resource, taskErr error) {
	finishedAt := time.Now()
	task.FinishedAt, task.UpdatedAt = finishedAt, finishedAt
	if taskErr != nil {
		task.Status, task.ErrorMessage = "failed", taskErr.Error()
		if cert != nil {
			cert.Status, cert.LastError, cert.UpdatedAt = "error", taskErr.Error(), finishedAt
			_ = b.repo.SaveCertificate(ctx, b.db, cert)
		}
	} else {
		cert.Certificate = string(resource.Certificate)
		cert.PrivateKey = string(resource.PrivateKey)
		cert.CertChain = string(resource.IssuerCertificate)
		cert.LastRenewAt, cert.LastError, cert.UpdatedAt = finishedAt, "", finishedAt
		if err := populateCertificate(cert); err != nil {
			task.Status, task.ErrorMessage = "failed", err.Error()
			cert.Status, cert.LastError = "error", err.Error()
		} else {
			task.Status, task.ErrorMessage = "success", ""
			result, _ := json.Marshal(map[string]any{"domain": resource.Domain, "cert_url": resource.CertURL, "not_after": cert.NotAfter})
			task.Result = string(result)
		}
		if err := b.repo.SaveCertificate(ctx, b.db, cert); err != nil {
			task.Status, task.ErrorMessage = "failed", err.Error()
		}
	}
	_ = b.repo.SaveTask(ctx, b.db, task)
}

func legoKeyType(value string) certcrypto.KeyType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "rsa4096":
		return certcrypto.RSA4096
	case "ec256":
		return certcrypto.EC256
	case "ec384":
		return certcrypto.EC384
	default:
		return certcrypto.RSA2048
	}
}

func certificateDomains(primary, sans string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)
	for _, value := range append([]string{primary}, strings.FieldsFunc(sans, func(r rune) bool { return r == ',' || r == '\n' || r == ';' })...) {
		domain := strings.TrimSpace(value)
		if domain == "" {
			continue
		}
		if _, ok := seen[domain]; ok {
			continue
		}
		seen[domain] = struct{}{}
		result = append(result, domain)
	}
	return result
}
