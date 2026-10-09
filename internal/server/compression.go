package server

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const gzipMinLength = 1024

var gzipExcludedExtensions = map[string]struct{}{
	".png": {}, ".gif": {}, ".jpeg": {}, ".jpg": {}, ".webp": {}, ".avif": {}, ".ico": {},
	".woff": {}, ".woff2": {}, ".ttf": {}, ".otf": {},
	".zip": {}, ".gz": {}, ".br": {}, ".7z": {}, ".rar": {}, ".tar": {}, ".tgz": {}, ".bz2": {}, ".xz": {},
	".mp3": {}, ".mp4": {}, ".webm": {}, ".ogg": {}, ".wav": {}, ".pdf": {},
}

var gzipExcludedPaths = []string{
	"/api/v1/assets/terminal/sftp/download",
	"/api/v1/assets/hosts/export",
	"/api/v1/assets/db-instances/export",
	"/api/v1/kubernetes/pods/files/download",
}

// gzipMiddleware buffers ordinary responses and compresses them only when
// useful. Streaming, upgraded and partial responses bypass it entirely.
func gzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		addVaryHeader(c.Writer.Header(), "Accept-Encoding")
		if !shouldUseGzip(c.Request) {
			c.Next()
			return
		}

		writer := &bufferedResponseWriter{ResponseWriter: c.Writer, status: http.StatusOK}
		c.Writer = writer
		c.Next()
		writer.commit()
	}
}

func shouldUseGzip(request *http.Request) bool {
	if request.Method == http.MethodHead || request.Header.Get("Range") != "" ||
		isConnectionUpgrade(request) || !acceptsGzip(request.Header.Get("Accept-Encoding")) {
		return false
	}
	for _, excludedPath := range gzipExcludedPaths {
		if strings.HasPrefix(request.URL.Path, excludedPath) {
			return false
		}
	}
	_, excluded := gzipExcludedExtensions[strings.ToLower(path.Ext(request.URL.Path))]
	return !excluded
}

type bufferedResponseWriter struct {
	gin.ResponseWriter
	body        bytes.Buffer
	status      int
	wroteHeader bool
}

func (w *bufferedResponseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
}

func (w *bufferedResponseWriter) WriteHeaderNow() {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
}

func (w *bufferedResponseWriter) Write(data []byte) (int, error) {
	w.WriteHeaderNow()
	return w.body.Write(data)
}

func (w *bufferedResponseWriter) WriteString(data string) (int, error) {
	w.WriteHeaderNow()
	return w.body.WriteString(data)
}

func (w *bufferedResponseWriter) Status() int { return w.status }

func (w *bufferedResponseWriter) Size() int { return w.body.Len() }

func (w *bufferedResponseWriter) Written() bool { return w.wroteHeader }

func (w *bufferedResponseWriter) commit() {
	body := w.body.Bytes()
	compress := w.status >= http.StatusOK && w.status < http.StatusBadRequest &&
		len(body) >= gzipMinLength && w.Header().Get("Content-Encoding") == ""
	if !compress {
		w.ResponseWriter.WriteHeader(w.status)
		_, _ = w.ResponseWriter.Write(body)
		return
	}

	w.Header().Del("Content-Length")
	w.Header().Set("Content-Encoding", "gzip")
	w.ResponseWriter.WriteHeader(w.status)
	gzipWriter := gzip.NewWriter(w.ResponseWriter)
	_, _ = gzipWriter.Write(body)
	_ = gzipWriter.Close()
}

func addVaryHeader(header http.Header, value string) {
	for _, current := range header.Values("Vary") {
		if headerContainsToken(current, value) {
			return
		}
	}
	header.Add("Vary", value)
}

func isConnectionUpgrade(request *http.Request) bool {
	return strings.EqualFold(strings.TrimSpace(request.Header.Get("Upgrade")), "websocket") ||
		headerContainsToken(request.Header.Get("Connection"), "upgrade")
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
			if len(keyValue) != 2 || !strings.EqualFold(strings.TrimSpace(keyValue[0]), "q") {
				continue
			}
			parsed, err := strconv.ParseFloat(strings.TrimSpace(keyValue[1]), 64)
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

func headerContainsToken(header, token string) bool {
	for _, value := range strings.Split(header, ",") {
		if strings.EqualFold(strings.TrimSpace(value), token) {
			return true
		}
	}
	return false
}
