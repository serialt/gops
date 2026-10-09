package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGzipMiddlewareCompressesLargeResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gzipMiddleware())
	r.POST("/large", func(c *gin.Context) {
		c.Header("Content-Length", "invalid-after-compression")
		c.String(http.StatusCreated, strings.Repeat("value", gzipMinLength))
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/large", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated || recorder.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected compressed 201 response, got status %d and encoding %q", recorder.Code, recorder.Header().Get("Content-Encoding"))
	}
	if recorder.Header().Get("Content-Length") != "" {
		t.Fatalf("compressed response retained Content-Length %q", recorder.Header().Get("Content-Length"))
	}
	reader, err := gzip.NewReader(recorder.Body)
	if err != nil {
		t.Fatal(err)
	}
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if err = reader.Close(); err != nil {
		t.Fatal(err)
	}
	if string(decompressed) != strings.Repeat("value", gzipMinLength) {
		t.Fatal("decompressed response does not match original body")
	}
}

func TestGzipMiddlewareLeavesSmallAndErrorResponsesUncompressed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gzipMiddleware())
	r.POST("/small", func(c *gin.Context) { c.String(http.StatusOK, "small") })
	r.POST("/error", func(c *gin.Context) { c.String(http.StatusBadRequest, strings.Repeat("x", gzipMinLength+1)) })

	for _, route := range []string{"/small", "/error"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, route, nil)
		request.Header.Set("Accept-Encoding", "gzip")
		r.ServeHTTP(recorder, request)
		if recorder.Header().Get("Content-Encoding") != "" {
			t.Fatalf("expected %s to remain uncompressed", route)
		}
	}
}

func TestGzipMiddlewareBypassesUnsafeRequests(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		header map[string]string
	}{
		{name: "quality zero", path: "/large", header: map[string]string{"Accept-Encoding": "gzip;q=0"}},
		{name: "range", path: "/large", header: map[string]string{"Accept-Encoding": "gzip", "Range": "bytes=0-10"}},
		{name: "websocket", path: "/large", header: map[string]string{"Accept-Encoding": "gzip", "Connection": "keep-alive, Upgrade", "Upgrade": "websocket"}},
		{name: "download", path: "/api/v1/assets/hosts/export", header: map[string]string{"Accept-Encoding": "gzip"}},
		{name: "compressed asset", path: "/image.png", header: map[string]string{"Accept-Encoding": "gzip"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(gzipMiddleware())
			r.POST(test.path, func(c *gin.Context) { c.String(http.StatusOK, strings.Repeat("x", gzipMinLength+1)) })
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, test.path, nil)
			for key, value := range test.header {
				request.Header.Set(key, value)
			}
			r.ServeHTTP(recorder, request)
			if recorder.Header().Get("Content-Encoding") != "" {
				t.Fatalf("expected response to remain uncompressed")
			}
			if !strings.Contains(recorder.Header().Get("Vary"), "Accept-Encoding") {
				t.Fatalf("expected Vary header, got %q", recorder.Header().Get("Vary"))
			}
		})
	}
}
