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
	return timeout.New(
		timeout.WithTimeout(50*time.Second),
		timeout.WithResponse(TimeoutResponse),
	)
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
