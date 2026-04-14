package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthHandler_Returns200(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHealthHandler_ReturnsValidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	healthHandler(w, req)

	var resp HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
	if resp.Service != serviceName {
		t.Errorf("expected service %q, got %q", serviceName, resp.Service)
	}
	if resp.Version != serviceVersion {
		t.Errorf("expected version %q, got %q", serviceVersion, resp.Version)
	}
	if resp.Timestamp == "" {
		t.Error("timestamp must not be empty")
	}
	if _, err := time.Parse(time.RFC3339, resp.Timestamp); err != nil {
		t.Errorf("timestamp is not RFC3339: %v", err)
	}
}

func TestHealthHandler_EngineCheckAlwaysOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	healthHandler(w, req)

	var resp HealthResponse
	json.NewDecoder(w.Body).Decode(&resp) //nolint:errcheck

	if resp.Checks["engine"] != "ok" {
		t.Errorf("engine check should always be 'ok', got %q", resp.Checks["engine"])
	}
}

func TestHealthHandler_OrchestratorDownGraceful(t *testing.T) {
	// No orchestrator running on default port — must still return 200.
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("health should return 200 even when orchestrator is down, got %d", w.Code)
	}

	var resp HealthResponse
	json.NewDecoder(w.Body).Decode(&resp) //nolint:errcheck

	if resp.Status != "ok" {
		t.Errorf("top-level status must remain 'ok' when orchestrator is unavailable")
	}
	orchStatus := resp.Checks["orchestrator"]
	if orchStatus != "unavailable" && orchStatus != "degraded" {
		t.Errorf("orchestrator check should be 'unavailable' or 'degraded', got %q", orchStatus)
	}
}

func TestHealthHandler_CORSHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	healthHandler(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("expected CORS header '*', got %q", origin)
	}
}

func TestPingOrchestrator_LiveServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	result := pingOrchestrator(t.Context(), ts.URL)
	if result != "ok" {
		t.Errorf("expected 'ok' from healthy mock server, got %q", result)
	}
}

func TestPingOrchestrator_Degraded(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	result := pingOrchestrator(t.Context(), ts.URL)
	if result != "degraded" {
		t.Errorf("expected 'degraded' from 503 server, got %q", result)
	}
}

func TestPingOrchestrator_Unreachable(t *testing.T) {
	// Port 19999 is unlikely to be occupied.
	result := pingOrchestrator(t.Context(), "http://localhost:19999")
	if result != "unavailable" {
		t.Errorf("expected 'unavailable' for unreachable host, got %q", result)
	}
}
