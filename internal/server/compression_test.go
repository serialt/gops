package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGzipMiddlewareCompressesLargeAPIResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gzipMiddleware())
	r.POST("/api/v1/example/list", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": strings.Repeat("value", gzipMinLength)})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/example/list", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if recorder.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip response, got %q", recorder.Header().Get("Content-Encoding"))
	}
}

func TestGzipMiddlewareSkipsSmallAndRangeResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gzipMiddleware())
	r.POST("/small", func(c *gin.Context) { c.String(http.StatusOK, "small") })
	r.POST("/range", func(c *gin.Context) { c.String(http.StatusOK, strings.Repeat("x", gzipMinLength+1)) })

	for _, test := range []struct {
		path        string
		rangeHeader string
	}{
		{path: "/small"},
		{path: "/range", rangeHeader: "bytes=0-99"},
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, test.path, nil)
		request.Header.Set("Accept-Encoding", "gzip")
		if test.rangeHeader != "" {
			request.Header.Set("Range", test.rangeHeader)
		}
		r.ServeHTTP(recorder, request)
		if recorder.Header().Get("Content-Encoding") != "" {
			t.Fatalf("expected %s response to remain uncompressed", test.path)
		}
	}
}
