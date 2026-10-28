package clientip_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ArminDashti/as-ip/server/internal/clientip"
	"github.com/gin-gonic/gin"
)

func TestFromContextPrefersCFConnectingIP(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("CF-Connecting-IP", "203.0.113.55")
	req.Header.Set("X-Forwarded-For", "203.0.113.55, 172.70.85.239")
	req.RemoteAddr = "172.16.0.2:8080"
	c.Request = req

	if got, want := clientip.FromContext(c), "203.0.113.55"; got != want {
		t.Fatalf("FromContext() = %q, want %q", got, want)
	}
}

func TestFromContextFallsBackToGinClientIP(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.10:12345"
	c.Request = req

	if got, want := clientip.FromContext(c), "203.0.113.10"; got != want {
		t.Fatalf("FromContext() = %q, want %q", got, want)
	}
}
