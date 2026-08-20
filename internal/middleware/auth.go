package middleware

import (
	"strings"

	"gops/internal/pkg/auth"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"

func Auth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, 401, "missing or invalid authorization header")
			return
		}

		claims, err := jwtManager.Parse(parts[1])
		if err != nil {
			response.Error(c, 401, "invalid or expired token")
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) (uint, bool) {
	value, ok := c.Get(UserIDKey)
	if !ok {
		return 0, false
	}
	id, ok := value.(uint)
	return id, ok
}
