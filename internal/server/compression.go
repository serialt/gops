package server

import (
	"net/http"
	"strconv"
	"strings"

	gingzip "github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

const gzipMinLength = 1024

var gzipExcludedExtensions = []string{
	".png", ".gif", ".jpeg", ".jpg", ".webp", ".avif", ".ico",
	".woff", ".woff2", ".ttf", ".otf",
	".zip", ".gz", ".br", ".7z", ".rar", ".tar", ".tgz", ".bz2", ".xz",
	".mp3", ".mp4", ".webm", ".ogg", ".wav", ".pdf",
}

// gzipMiddleware applies response compression in the same place as a reverse
// proxy: globally, before application routes are registered.
func gzipMiddleware() gin.HandlerFunc {
	compress := gingzip.Gzip(
		gingzip.DefaultCompression,
		gingzip.WithMinLength(gzipMinLength),
		gingzip.WithExcludedExtensions(gzipExcludedExtensions),
		gingzip.WithCustomShouldCompressFn(func(c *gin.Context) bool {
			return c.Request.Method != http.MethodHead &&
				c.Request.Header.Get("Range") == "" &&
				acceptsGzip(c.Request.Header.Get("Accept-Encoding"))
		}),
	)
	return func(c *gin.Context) {
		c.Writer.Header().Add("Vary", "Accept-Encoding")
		compress(c)
	}
}

func acceptsGzip(header string) bool {
	wildcardQuality := -1.0
	for _, item := range strings.Split(header, ",") {
		parts := strings.Split(item, ";")
		encoding := strings.TrimSpace(strings.ToLower(parts[0]))
		if encoding != "gzip" && encoding != "*" {
			continue
		}

		quality := 1.0
		for _, parameter := range parts[1:] {
			keyValue := strings.SplitN(strings.TrimSpace(parameter), "=", 2)
			if len(keyValue) != 2 || !strings.EqualFold(keyValue[0], "q") {
				continue
			}
			parsed, err := strconv.ParseFloat(keyValue[1], 64)
			if err != nil || parsed < 0 || parsed > 1 {
				quality = 0
			} else {
				quality = parsed
			}
		}
		if encoding == "gzip" {
			return quality > 0
		}
		wildcardQuality = quality
	}
	return wildcardQuality > 0
}
