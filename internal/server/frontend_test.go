package server

import (
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
}

func TestRegisterFrontendKeepsAPI404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerFrontend(r)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/missing", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("expected JSON response, got %q", recorder.Header().Get("Content-Type"))
	}
}
