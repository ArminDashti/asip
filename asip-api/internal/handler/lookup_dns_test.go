package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ArminDashti/as-ip/server/internal/handler"
	"github.com/ArminDashti/as-ip/server/internal/service"
	"github.com/gin-gonic/gin"
)

func TestLookupDnsBadRequest(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	lookupHandler := handler.NewLookupHandler(service.NewLookupService(nil))
	engine := gin.New()
	engine.GET("/api/v1/dns/lookup/*domain", lookupHandler.LookupDns)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dns/lookup/%20", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body["error"] != "bad_request" {
		t.Fatalf("error = %v, want bad_request", body["error"])
	}
}

func TestDnsLookupRouteRegistered(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	lookupHandler := handler.NewLookupHandler(service.NewLookupService(nil))
	docsHandler := handler.NewDocsHandler()
	engine := gin.New()
	engine.GET("/api/v1/docs", docsHandler.GetDocs)
	engine.GET("/api/v1/dns/lookup/*domain", lookupHandler.LookupDns)

	found := false
	for _, route := range engine.Routes() {
		if route.Method == http.MethodGet && route.Path == "/api/v1/dns/lookup/*domain" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected /api/v1/dns/lookup/*domain route to be registered")
	}
}
