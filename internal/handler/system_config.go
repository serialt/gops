package handler

import (
	"log/slog"

	"gops/internal/biz"
	"gops/internal/data/api"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type SystemConfigHandler struct{ biz *biz.SystemConfigBiz }

func NewSystemConfigHandler(systemConfigBiz *biz.SystemConfigBiz) *SystemConfigHandler {
	return &SystemConfigHandler{biz: systemConfigBiz}
}

// Get POST /api/v1/system-config/get 查询系统配置。
func (h *SystemConfigHandler) Get(c *gin.Context) {
	result, err := h.biz.Get()
	if err != nil {
		response.Error(c, 500, "query system config failed")
		return
	}
	response.OK(c, result)
}

// Update POST /api/v1/system-config/update 更新系统配置。
func (h *SystemConfigHandler) Update(c *gin.Context) {
	req := new(api.SystemConfigRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "invalid system config")
		return
	}
	result, err := h.biz.Update(req)
	if err != nil {
		response.Error(c, 500, "update system config failed")
		return
	}
	response.OK(c, result)
}

// TestEmail POST /api/v1/system-config/email/test 使用当前 SMTP 配置发送测试邮件。
func (h *SystemConfigHandler) TestEmail(c *gin.Context) {
	req := new(api.SystemConfigEmailTestRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "invalid test email request")
		return
	}
	if err := h.biz.TestEmail(req); err != nil {
		slog.Error("SMTP test email failed", "recipient", req.Recipient, "err", err)
		response.Error(c, 500, "send test email failed")
		return
	}
	response.OK(c, nil)
}
