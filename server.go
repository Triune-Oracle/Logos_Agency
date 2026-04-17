package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	serviceVersion = "0.1.0"
	serviceName    = "logos-agency-engine"
)

// HealthResponse is the JSON body returned by /api/health and /health.
type HealthResponse struct {
	Status    string            `json:"status"`
	Service   string            `json:"service"`
	Version   string            `json:"version"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

// pingOrchestrator does a non-blocking GET /health against the orchestrator
// and returns "ok", "degraded", or "unavailable".
func pingOrchestrator(ctx context.Context, baseURL string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/health", nil)
	if err != nil {
		return "unavailable"
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "unavailable"
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return "ok"
	}
	return "degraded"
}

// healthHandler serves GET /api/health (and /health as a convenience alias).
// It always returns HTTP 200 — the health of individual subsystems is
// communicated inside the JSON body so callers can distinguish partial
// degradation from a total outage.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	orchURL := os.Getenv("ORCHESTRATOR_URL")
	if orchURL == "" {
		orchURL = "http://localhost:5000"
	}

	checks := map[string]string{
		"engine": "ok",
	}

	// Ping orchestrator with a short timeout; never blocks the response.
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	checks["orchestrator"] = pingOrchestrator(ctx, orchURL)

	body := HealthResponse{
		Status:    "ok",
		Service:   serviceName,
		Version:   serviceVersion,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("health encode error: %v", err)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/health", healthHandler) // Docker healthcheck alias

	addr := fmt.Sprintf(":%s", port)
	log.Printf("%s v%s listening on %s", serviceName, serviceVersion, addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
