package handler

import (
	"gops/internal/biz"
	"gops/internal/data/api"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type SSLCertHandler struct{ biz *biz.SSLCertBiz }

func NewSSLCertHandler(b *biz.SSLCertBiz) *SSLCertHandler { return &SSLCertHandler{biz: b} }

func (h *SSLCertHandler) RegisterRoutes(g *gin.RouterGroup) {
	g.POST("/ssl-cert/certificates/list", h.ListCertificates)
	g.POST("/ssl-cert/certificates/get", h.GetCertificate)
	g.POST("/ssl-cert/certificates/save", h.SaveCertificate)
	g.POST("/ssl-cert/certificates/delete", h.DeleteCertificate)
	g.POST("/ssl-cert/certificates/renew", h.RenewCertificate)
	g.POST("/ssl-cert/applications/create", h.ApplyCertificate)
	g.POST("/ssl-cert/dns-providers/list", h.ListDNSProviders)
	g.POST("/ssl-cert/dns-providers/get", h.GetDNSProvider)
	g.POST("/ssl-cert/dns-providers/save", h.SaveDNSProvider)
	g.POST("/ssl-cert/dns-providers/delete", h.DeleteDNSProvider)
	g.POST("/ssl-cert/dns-providers/test", h.TestDNSProvider)
	g.POST("/ssl-cert/deploy-configs/list", h.ListDeployConfigs)
	g.POST("/ssl-cert/deploy-configs/get", h.GetDeployConfig)
	g.POST("/ssl-cert/deploy-configs/save", h.SaveDeployConfig)
	g.POST("/ssl-cert/deploy-configs/delete", h.DeleteDeployConfig)
	g.POST("/ssl-cert/deploy-configs/test", h.TestDeployConfig)
	g.POST("/ssl-cert/deploy-configs/execute", h.ExecuteDeploy)
	g.POST("/ssl-cert/tasks/list", h.ListTasks)
	g.POST("/ssl-cert/tasks/get", h.GetTask)
}

// ApplyCertificate POST /api/v1/ssl-cert/applications/create 创建 ACME 证书申请任务。
func (h *SSLCertHandler) ApplyCertificate(c *gin.Context) {
	req := new(api.SSLCertificateApplyRequest)
	if bindSSL(c, req) {
		data, err := h.biz.ApplyCertificate(c.Request.Context(), req)
		sslReply(c, data, err)
	}
}

func bindSSL(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "请求参数错误")
		return false
	}
	return true
}
func sslReply(c *gin.Context, data any, err error) {
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// ListCertificates POST /api/v1/ssl-cert/certificates/list 查询 SSL 证书。
func (h *SSLCertHandler) ListCertificates(c *gin.Context) {
	req := new(api.SSLCertificateListRequest)
	if bindSSL(c, req) {
		data, err := h.biz.ListCertificates(c.Request.Context(), req)
		sslReply(c, data, err)
	}
}

// GetCertificate POST /api/v1/ssl-cert/certificates/get 查询 SSL 证书详情。
func (h *SSLCertHandler) GetCertificate(c *gin.Context) {
	req := new(api.SSLCertificateIDRequest)
	if bindSSL(c, req) {
		data, err := h.biz.GetCertificate(c.Request.Context(), req.ID)
		sslReply(c, data, err)
	}
}

// SaveCertificate POST /api/v1/ssl-cert/certificates/save 保存 SSL 证书。
func (h *SSLCertHandler) SaveCertificate(c *gin.Context) {
	req := new(api.SSLCertificateRequest)
	if bindSSL(c, req) {
		data, err := h.biz.SaveCertificate(c.Request.Context(), req)
		sslReply(c, data, err)
	}
}

// DeleteCertificate POST /api/v1/ssl-cert/certificates/delete 删除 SSL 证书。
func (h *SSLCertHandler) DeleteCertificate(c *gin.Context) {
	req := new(api.SSLCertificateIDRequest)
	if bindSSL(c, req) {
		sslReply(c, nil, h.biz.DeleteCertificate(c.Request.Context(), req.ID))
	}
}

// RenewCertificate POST /api/v1/ssl-cert/certificates/renew 创建 SSL 证书续期任务。
func (h *SSLCertHandler) RenewCertificate(c *gin.Context) {
	req := new(api.SSLCertificateIDRequest)
	if bindSSL(c, req) {
		data, err := h.biz.RenewCertificate(c.Request.Context(), req.ID)
		sslReply(c, data, err)
	}
}

// ListDNSProviders POST /api/v1/ssl-cert/dns-providers/list 查询 DNS 配置。
func (h *SSLCertHandler) ListDNSProviders(c *gin.Context) {
	req := new(api.SSLDNSProviderListRequest)
	if bindSSL(c, req) {
		data, err := h.biz.ListDNSProviders(c.Request.Context(), req)
		sslReply(c, data, err)
	}
}

// GetDNSProvider POST /api/v1/ssl-cert/dns-providers/get 查询 DNS 配置详情。
func (h *SSLCertHandler) GetDNSProvider(c *gin.Context) {
	req := new(api.SSLCertificateIDRequest)
	if bindSSL(c, req) {
		data, err := h.biz.GetDNSProvider(c.Request.Context(), req.ID)
		sslReply(c, data, err)
	}
}

// SaveDNSProvider POST /api/v1/ssl-cert/dns-providers/save 保存 DNS 配置。
func (h *SSLCertHandler) SaveDNSProvider(c *gin.Context) {
	req := new(api.SSLDNSProviderRequest)
	if bindSSL(c, req) {
		data, err := h.biz.SaveDNSProvider(c.Request.Context(), req)
		sslReply(c, data, err)
	}
}

// DeleteDNSProvider POST /api/v1/ssl-cert/dns-providers/delete 删除 DNS 配置。
func (h *SSLCertHandler) DeleteDNSProvider(c *gin.Context) {
	req := new(api.SSLCertificateIDRequest)
	if bindSSL(c, req) {
		sslReply(c, nil, h.biz.DeleteDNSProvider(c.Request.Context(), req.ID))
	}
}

// TestDNSProvider POST /api/v1/ssl-cert/dns-providers/test 测试 DNS 配置。
func (h *SSLCertHandler) TestDNSProvider(c *gin.Context) {
	req := new(api.SSLCertificateIDRequest)
	if bindSSL(c, req) {
		sslReply(c, nil, h.biz.TestDNSProvider(c.Request.Context(), req.ID))
	}
}

// ListDeployConfigs POST /api/v1/ssl-cert/deploy-configs/list 查询证书部署配置。
func (h *SSLCertHandler) ListDeployConfigs(c *gin.Context) {
	req := new(api.SSLDeployConfigListRequest)
	if bindSSL(c, req) {
		data, err := h.biz.ListDeployConfigs(c.Request.Context(), req)
		sslReply(c, data, err)
	}
}

// GetDeployConfig POST /api/v1/ssl-cert/deploy-configs/get 查询证书部署配置详情。
func (h *SSLCertHandler) GetDeployConfig(c *gin.Context) {
	req := new(api.SSLCertificateIDRequest)
	if bindSSL(c, req) {
		data, err := h.biz.GetDeployConfig(c.Request.Context(), req.ID)
		sslReply(c, data, err)
	}
}

// SaveDeployConfig POST /api/v1/ssl-cert/deploy-configs/save 保存证书部署配置。
func (h *SSLCertHandler) SaveDeployConfig(c *gin.Context) {
	req := new(api.SSLDeployConfigRequest)
	if bindSSL(c, req) {
		data, err := h.biz.SaveDeployConfig(c.Request.Context(), req)
		sslReply(c, data, err)
	}
}

// DeleteDeployConfig POST /api/v1/ssl-cert/deploy-configs/delete 删除证书部署配置。
func (h *SSLCertHandler) DeleteDeployConfig(c *gin.Context) {
	req := new(api.SSLCertificateIDRequest)
	if bindSSL(c, req) {
		sslReply(c, nil, h.biz.DeleteDeployConfig(c.Request.Context(), req.ID))
	}
}

// TestDeployConfig POST /api/v1/ssl-cert/deploy-configs/test 测试证书部署配置。
func (h *SSLCertHandler) TestDeployConfig(c *gin.Context) {
	req := new(api.SSLCertificateIDRequest)
	if bindSSL(c, req) {
		sslReply(c, nil, h.biz.TestDeployConfig(c.Request.Context(), req.ID))
	}
}

// ExecuteDeploy POST /api/v1/ssl-cert/deploy-configs/execute 创建证书部署任务。
func (h *SSLCertHandler) ExecuteDeploy(c *gin.Context) {
	req := new(api.SSLCertificateIDRequest)
	if bindSSL(c, req) {
		data, err := h.biz.ExecuteDeploy(c.Request.Context(), req.ID)
		sslReply(c, data, err)
	}
}

// ListTasks POST /api/v1/ssl-cert/tasks/list 查询 SSL 任务记录。
func (h *SSLCertHandler) ListTasks(c *gin.Context) {
	req := new(api.SSLTaskListRequest)
	if bindSSL(c, req) {
		data, err := h.biz.ListTasks(c.Request.Context(), req)
		sslReply(c, data, err)
	}
}

// GetTask POST /api/v1/ssl-cert/tasks/get 查询 SSL 任务详情。
func (h *SSLCertHandler) GetTask(c *gin.Context) {
	req := new(api.SSLCertificateIDRequest)
	if bindSSL(c, req) {
		data, err := h.biz.GetTask(c.Request.Context(), req.ID)
		sslReply(c, data, err)
	}
}
