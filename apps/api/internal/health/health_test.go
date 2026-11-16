package health_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mjengofiti/api/internal/health"
)

func TestHealthLive(t *testing.T) {
	svc := health.NewService(nil)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	w := httptest.NewRecorder()

	svc.HandleLive(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var res map[string]string
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}

	if res["status"] != "ok" {
		t.Errorf("expected status ok, got %s", res["status"])
	}
}

func TestSystemStatusOmitsDetailedInfrastructureData(t *testing.T) {
	svc := health.NewService(nil)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/system/status", nil)
	response := httptest.NewRecorder()

	svc.HandleSystemStatus(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var result map[string]any
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatalf("decode system status: %v", err)
	}
	if result["status"] != "degraded" {
		t.Fatalf("expected degraded status without a database, got %v", result["status"])
	}
	for _, key := range []string{"environment", "version", "dependencies", "time"} {
		if _, exists := result[key]; exists {
			t.Errorf("system status unexpectedly exposed %q", key)
		}
	}
	database, ok := result["database"].(map[string]any)
	if !ok || database["status"] != "unavailable" {
		t.Fatalf("expected coarse unavailable database status, got %#v", result["database"])
	}
}

func TestHealthReadyDegraded(t *testing.T) {
	svc := health.NewService(nil)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()

	svc.HandleReady(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503 for disconnected database, got %d", w.Code)
	}
}
