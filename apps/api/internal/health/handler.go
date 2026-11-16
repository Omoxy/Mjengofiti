package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mjengofiti/api/internal/database"
)

// Service provides health check operations.
type Service struct {
	db      *database.Pool
	startAt time.Time
}

// NewService creates a new health check service.
func NewService(db *database.Pool) *Service {
	return &Service{
		db:      db,
		startAt: time.Now(),
	}
}

// LiveResponse is returned by the liveness probe.
type LiveResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

// ReadyResponse is returned by the readiness probe.
type ReadyResponse struct {
	Status   string            `json:"status"`
	Time     string            `json:"time"`
	Uptime   string            `json:"uptime"`
	Checks   map[string]string `json:"checks"`
}

// SystemStatusResponse is returned by the system status endpoint.
type SystemStatusResponse struct {
	Status        string         `json:"status"`
	Database      DatabaseStatus `json:"database"`
	UptimeSeconds int64          `json:"uptime_seconds"`
}

type DatabaseStatus struct {
	Status string `json:"status"`
	Type   string `json:"type"`
}

// HandleLive returns 200 if the process is running.
func (s *Service) HandleLive(w http.ResponseWriter, r *http.Request) {
	resp := LiveResponse{
		Status: "ok",
		Time:   time.Now().UTC().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, resp)
}

// HandleReady returns 200 only if all critical dependencies are reachable.
func (s *Service) HandleReady(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)
	overallStatus := "ok"

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if s.db != nil {
		if err := s.db.HealthCheck(ctx); err != nil {
			checks["database"] = "unavailable"
			overallStatus = "degraded"
		} else {
			checks["database"] = "ok"
		}
	} else {
		checks["database"] = "not_configured"
		overallStatus = "degraded"
	}

	status := http.StatusOK
	if overallStatus != "ok" {
		status = http.StatusServiceUnavailable
	}

	resp := ReadyResponse{
		Status: overallStatus,
		Time:   time.Now().UTC().Format(time.RFC3339),
		Uptime: time.Since(s.startAt).Round(time.Second).String(),
		Checks: checks,
	}
	writeJSON(w, status, resp)
}

// HandleSystemStatus returns detailed system status for authenticated admins.
func (s *Service) HandleSystemStatus(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	overallStatus := "ok"
	database := DatabaseStatus{Status: "unavailable"}

	if s.db != nil {
		if err := s.db.HealthCheck(ctx); err != nil {
			overallStatus = "degraded"
		} else {
			database.Status = "connected"
		}
	} else {
		overallStatus = "degraded"
	}

	resp := SystemStatusResponse{
		Status: overallStatus,
		Database: DatabaseStatus{
			Status: database.Status,
			Type:   "PostgreSQL/PostGIS",
		},
		UptimeSeconds: int64(time.Since(s.startAt).Seconds()),
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error":"internal_error"}`, http.StatusInternalServerError)
	}
}
