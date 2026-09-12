package handler

import (
	"github.com/gin-gonic/gin"
	"gops/internal/biz"
	"gops/internal/data/api"
	"gops/internal/pkg/response"
)

type SSLCertHandler struct{ biz *biz.SSLCertBiz }

func NewSSLCertHandler(b *biz.SSLCertBiz) *SSLCertHandler { return &SSLCertHandler{b} }

// ListCertificates POST /api/v1/ssl-cert/certificates/list 查询 SSL 证书。
func (h *SSLCertHandler) ListCertificates(c *gin.Context) {
	r := new(api.SSLCertificateListRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid certificate query")
		return
	}
	d, e := h.biz.List(r)
	if e != nil {
		response.Error(c, 500, e.Error())
		return
	}
	response.OK(c, d)
}

// SaveCertificate POST /api/v1/ssl-cert/certificates/save 保存 SSL 证书。
func (h *SSLCertHandler) SaveCertificate(c *gin.Context) {
	r := new(api.SSLCertificateRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid certificate data")
		return
	}
	d, e := h.biz.Save(r)
	if e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	response.OK(c, d)
}
