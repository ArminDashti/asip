package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ArminDashti/as-ip/server/internal/dto"
	"github.com/ArminDashti/as-ip/server/internal/handler"
	"github.com/gin-gonic/gin"
)

func TestGetHttpHeadersEchoesAndRedacts(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	headersHandler := handler.NewHeadersHandler()
	engine := gin.New()
	engine.GET("/api/v1/http/headers", headersHandler.GetHttpHeaders)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/http/headers", nil)
	req.Header.Set("User-Agent", "asip-test/1.0")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Cookie", "secret=value")
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("Connection", "keep-alive")
	req.RemoteAddr = "203.0.113.10:12345"

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var body dto.HttpHeadersResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body.Ip != "203.0.113.10" {
		t.Fatalf("ip = %q, want %q", body.Ip, "203.0.113.10")
	}

	byName := map[string]string{}
	for _, entry := range body.Headers {
		byName[entry.Name] = entry.Value
	}

	if byName["User-Agent"] != "asip-test/1.0" {
		t.Fatalf("User-Agent = %q", byName["User-Agent"])
	}
	if byName["Accept-Language"] != "en-US" {
		t.Fatalf("Accept-Language = %q", byName["Accept-Language"])
	}
	if byName["Cookie"] != "[redacted]" {
		t.Fatalf("Cookie = %q, want [redacted]", byName["Cookie"])
	}
	if byName["Authorization"] != "[redacted]" {
		t.Fatalf("Authorization = %q, want [redacted]", byName["Authorization"])
	}
	if _, ok := byName["Connection"]; ok {
		t.Fatal("Connection header should be omitted")
	}
}
