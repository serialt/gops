package middleware

import (
	"strconv"

	"gops/internal/pkg/casbinx"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func Casbin(service *casbinx.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := CurrentUserID(c)
		if !ok {
			response.Error(c, 401, "unauthorized")
			return
		}

		allowed, err := service.Enforcer.Enforce(
			strconv.FormatUint(uint64(userID), 10),
			c.Request.URL.Path,
			c.Request.Method,
		)
		if err != nil {
			response.Error(c, 500, "permission check failed")
			return
		}
		if !allowed {
			response.Error(c, 403, "permission denied")
			return
		}
		c.Next()
	}
}
