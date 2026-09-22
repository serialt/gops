package server

import (
	"bytes"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"gops/internal/webui"
	loginassets "gops/web/src/assets/login"

	"github.com/gin-gonic/gin"
)

// registerFrontend serves the embedded frontend and falls back to index.html
// for routes handled by Vue Router. Unknown API routes keep returning JSON 404.
func registerFrontend(r *gin.Engine) {
	dist, err := fs.Sub(webui.Assets, "dist")
	if err != nil {
		panic("open embedded frontend: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(dist))
	r.StaticFS("/__fallback", http.FS(loginassets.Assets))
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "message": "route not found"})
			return
		}

		requestedPath := strings.TrimPrefix(path.Clean(c.Request.URL.Path), "/")
		if requestedPath != "." && requestedPath != "" {
			if info, statErr := fs.Stat(dist, requestedPath); statErr == nil && !info.IsDir() {
				if strings.HasPrefix(requestedPath, "assets/") {
					c.Header("Cache-Control", "public, max-age=31536000, immutable")
				}
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}

		index, readErr := fs.ReadFile(dist, "index.html")
		if readErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": "frontend unavailable"})
			return
		}
		if bytes.Contains(index, []byte("gops-fallback-placeholder")) {
			index, readErr = fs.ReadFile(webui.Assets, "fallback.html")
			if readErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": "frontend fallback unavailable"})
				return
			}
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})
}
