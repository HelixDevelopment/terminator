//go:build integration

package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/helixdevelopment/health-service/internal/checker"
	"github.com/helixdevelopment/health-service/internal/handler"
)

func TestHealthCheck_Endpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	chk := checker.New(map[string]string{}, time.Second)
	h := handler.New(chk)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/healthz", nil)
	h.HealthCheck(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestReadinessCheck_Endpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	chk := checker.New(map[string]string{}, time.Second)
	h := handler.New(chk)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/healthz/ready", nil)
	h.ReadinessCheck(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
