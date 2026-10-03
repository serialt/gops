package middleware

import (
	"gops/internal/config"

	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/slog"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
)

func TimeoutResponse(c *gin.Context) {
	c.String(http.StatusRequestTimeout, "timeout")
}

func TimeoutMiddleware() gin.HandlerFunc {
	defaultTimeout := timeout.New(
		timeout.WithTimeout(50*time.Second),
		timeout.WithResponse(TimeoutResponse),
	)
	return func(c *gin.Context) {
		switch c.Request.URL.Path {
		case "/api/v1/assets/terminal/sftp/upload", "/api/v1/assets/terminal/sftp/download":
			// File transfers can legitimately exceed the ordinary API timeout and
			// must stream directly instead of using timeout middleware's buffer.
			c.Next()
		default:
			defaultTimeout(c)
		}
	}
}

func RegisterMiddleware(app *gin.Engine, cfg *config.Config, whitelistMap map[string]string) {
	// app.Use(
	// 	ginzap.Ginzap(pkg.Log, time.RFC3339, true),
	// 	ginzap.RecoveryWithZap(pkg.Log, true),
	// 	cors.Default())
	app.Use(
		TimeoutMiddleware(),
		// authM.AuthRequired(whitelistMap),
		// authM.Casbin(whitelistMap),
		cors.Default(),
		slog.SetLogger(),
		gin.Recovery(),
	)

}
