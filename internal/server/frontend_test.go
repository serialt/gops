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

func TestRegisterFrontendSPAFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerFrontend(r)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/assets/hosts", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("expected HTML response, got %q", recorder.Header().Get("Content-Type"))
	}
	if !strings.Contains(recorder.Body.String(), "Service running") {
		t.Fatal("expected the styled frontend fallback page")
	}
}

func TestRegisterFrontendServesFallbackIllustration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerFrontend(r)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/__fallback/illustration.svg", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Header().Get("Content-Type"), "image/svg+xml") {
		t.Fatalf("expected SVG response, got %q", recorder.Header().Get("Content-Type"))
	}
}

func TestRegisterFrontendServesFallbackLogo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerFrontend(r)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/__fallback/gops.svg", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Header().Get("Content-Type"), "image/svg+xml") {
		t.Fatalf("expected SVG response, got %q", recorder.Header().Get("Content-Type"))
	}
}

func TestRegisterFrontendKeepsAPI404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerFrontend(r)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/missing", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("expected JSON response, got %q", recorder.Header().Get("Content-Type"))
	}
	if recorder.Header().Get("Content-Encoding") != "" {
		t.Fatalf("expected API response to remain uncompressed, got %q", recorder.Header().Get("Content-Encoding"))
	}
}

func TestRegisterFrontendCompressesWithGzip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gzipMiddleware())
	registerFrontend(r)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/assets/hosts", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if recorder.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip response, got %q", recorder.Header().Get("Content-Encoding"))
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
	if !strings.Contains(string(decompressed), "Service running") {
		t.Fatal("expected compressed frontend content")
	}
}

func TestRegisterFrontendRespectsGzipQualityZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gzipMiddleware())
	registerFrontend(r)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/assets/hosts", nil)
	request.Header.Set("Accept-Encoding", "gzip;q=0")
	r.ServeHTTP(recorder, request)

	if recorder.Header().Get("Content-Encoding") != "" {
		t.Fatalf("expected uncompressed response, got %q", recorder.Header().Get("Content-Encoding"))
	}
	if !strings.Contains(recorder.Header().Get("Vary"), "Accept-Encoding") {
		t.Fatalf("expected Vary header, got %q", recorder.Header().Get("Vary"))
	}
}
