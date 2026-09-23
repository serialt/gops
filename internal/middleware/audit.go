package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"gops/internal/data/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// OperationAudit records authenticated business requests after permission checks.
func OperationAudit(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if skipOperationAudit(path) {
			c.Next()
			return
		}

		startedAt := time.Now()
		userID, _ := CurrentUserID(c)
		username := ""
		if userID > 0 {
			_ = db.Model(&model.User{}).Where("id = ?", userID).Pluck("username", &username).Error
		}

		c.Next()

		module, action, description := operationInfo(path)
		status := c.Writer.Status()
		errorMessage := ""
		if status >= http.StatusBadRequest {
			errorMessage = http.StatusText(status)
		}
		entry := &model.AuditOperationLog{
			UserID:       int64(userID),
			Username:     username,
			Module:       module,
			Action:       action,
			Description:  description,
			Method:       c.Request.Method,
			Path:         path,
			Status:       int64(status),
			ErrorMessage: errorMessage,
			CostTime:     time.Since(startedAt).Milliseconds(),
			IP:           c.ClientIP(),
		}
		if err := db.Create(entry).Error; err != nil {
			slog.Error("保存操作日志失败", "error", err, "path", path, "username", username)
		}
	}
}

func skipOperationAudit(path string) bool {
	return path == "/api/v1/auth/me" ||
		path == "/api/v1/audit/operation-logs/list" ||
		path == "/api/v1/audit/login-logs/list" ||
		path == "/api/v1/audit/data-logs/list"
}

func operationInfo(path string) (module, action, description string) {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	domain := ""
	verb := ""
	if len(segments) >= 3 {
		domain = segments[2]
	}
	if len(segments) > 0 {
		verb = segments[len(segments)-1]
	}

	moduleNames := map[string]string{
		"users": "用户管理", "roles": "角色管理", "organizations": "组织管理",
		"assets": "资产管理", "monitor": "监控管理", "nginx": "Nginx 管理",
		"tasks": "任务中心", "kubernetes": "容器管理", "audit": "审计管理",
		"ssl-cert": "SSL 证书", "system-config": "系统配置", "dashboard": "仪表盘",
	}
	module = moduleNames[domain]
	if module == "" {
		module = "系统"
	}

	actionNames := map[string]string{
		"list": "查询", "get": "查询", "options": "查询", "stats": "查询", "history": "查询",
		"create": "创建", "save": "保存", "update": "更新", "set": "配置", "move": "移动",
		"delete": "删除", "test": "测试", "sync": "同步", "drain": "排空", "recording": "查看",
	}
	action = actionNames[verb]
	if action == "" {
		action = "操作"
	}
	description = action + module
	return module, action, description
}
