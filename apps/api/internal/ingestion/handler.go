package ingestion

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mjengofiti/api/internal/auth"
)

// Handler exposes HTTP API routes for the ingestion subsystem.
type Handler struct {
	svc    *Service
	logger *slog.Logger
}

// NewHandler creates a new Ingestion Handler.
func NewHandler(svc *Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// Routes registers the ingestion endpoints on a chi router.
func (h *Handler) Routes(r chi.Router) {
	r.With(auth.RequirePermission("datasets:write")).Post("/api/v1/ingestion/sources", h.RegisterSource)
		r.Get("/api/v1/ingestion/sources", h.ListSources)
	r.With(auth.RequirePermission("datasets:write")).Post("/api/v1/ingestion/acquire", h.Acquire)
	r.With(auth.RequirePermission("datasets:write")).Get("/api/v1/ingestion/artifacts/{versionId}", h.GetArtifact)
	r.With(auth.RequirePermission("datasets:write")).Get("/api/v1/ingestion/jobs", h.ListJobs)
	r.With(auth.RequirePermission("datasets:write")).Post("/api/v1/ingestion/jobs", h.CreateJob)
	r.With(auth.RequirePermission("datasets:write")).Get("/api/v1/ingestion/jobs/{id}", h.GetJob)
	r.With(auth.RequirePermission("datasets:write")).Post("/api/v1/ingestion/validate", h.Validate)
	r.With(auth.RequirePermission("datasets:write")).Get("/api/v1/ingestion/reports/{versionId}", h.GetReport)
	r.With(auth.RequirePermission("datasets:promote")).Post("/api/v1/ingestion/promote", h.Promote)
}


func (h *Handler) ListSources(w http.ResponseWriter, r *http.Request) {
	sources := h.svc.ListSources()
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"sources": sources,
		"total":   len(sources),
	})
}

func (h *Handler) RegisterSource(w http.ResponseWriter, r *http.Request) {
	var src Source
	if !decodeRequest(w, r, 1<<20, &src) {
		return
	}

	created, err := h.svc.RegisterSource(src)
	if err != nil {
		respondError(w, http.StatusBadRequest, "registration_failed", "Source registration was rejected")
		return
	}

	respondJSON(w, http.StatusCreated, created)
}

func (h *Handler) ListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.svc.ListJobs(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "list_jobs_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"jobs":  jobs,
		"total": len(jobs),
	})
}

func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DatasetID string `json:"dataset_id"`
		Version   string `json:"version"`
	}

		if !decodeRequest(w, r, 1<<16, &req) {
		return
	}

		user := auth.UserFromContext(r.Context())
		if user == nil {
			respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required")
			return
		}
		job, err := h.svc.CreateJob(r.Context(), req.DatasetID, req.Version, user.ID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "create_job_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, job)
}

func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	job, err := h.svc.GetJob(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "job_not_found", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, job)
}

func (h *Handler) Validate(w http.ResponseWriter, r *http.Request) {
	versionID := r.URL.Query().Get("version_id")

		const maxPayloadSize = 25 * 1024 * 1024
		rawData, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxPayloadSize))
	if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				respondError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "GeoJSON payload exceeds the 25 MB limit")
				return
			}
		respondError(w, http.StatusBadRequest, "read_error", "Failed to read payload")
		return
	}

	report, err := h.svc.ValidateDataset(r.Context(), versionID, rawData)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "validation_failed", err.Error())
		return
	}

	status := http.StatusOK
	if report.OverallStatus == "failed" {
		status = http.StatusUnprocessableEntity
	}

	respondJSON(w, status, report)
}

func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	versionID := chi.URLParam(r, "versionId")
	report, err := h.svc.GetValidationReport(r.Context(), versionID)
	if err != nil {
		respondError(w, http.StatusNotFound, "report_not_found", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, report)
}

func (h *Handler) Promote(w http.ResponseWriter, r *http.Request) {
	var req PromotionRequest
		if !decodeRequest(w, r, 1<<16, &req) {
		return
	}

		user := auth.UserFromContext(r.Context())
		if user == nil {
			respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required")
			return
		}
		req.PromotedBy = user.ID
	res, err := h.svc.PromoteVersion(r.Context(), req)
	if err != nil {
			respondError(w, http.StatusConflict, "promotion_rejected", "Dataset promotion was rejected")
		return
	}

	respondJSON(w, http.StatusOK, res)
}

func (h *Handler) Acquire(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SourceID  string `json:"source_id"`
		VersionID string `json:"version_id"`
	}
		if !decodeRequest(w, r, 1<<16, &req) {
		return
	}

	artifact, err := h.svc.AcquireSource(r.Context(), req.SourceID, req.VersionID)
	if err != nil {
			respondError(w, http.StatusBadGateway, "acquisition_failed", "External source acquisition failed")
		return
	}

	respondJSON(w, http.StatusOK, artifact)
}

func (h *Handler) GetArtifact(w http.ResponseWriter, r *http.Request) {
	versionID := chi.URLParam(r, "versionId")
	artifact, err := h.svc.GetArtifact(r.Context(), versionID)
	if err != nil {
		respondError(w, http.StatusNotFound, "artifact_not_found", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, artifact)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func decodeRequest(w http.ResponseWriter, r *http.Request, maxBytes int64, destination any) bool {
	err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBytes)).Decode(destination)
	if err == nil {
		return true
	}
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		respondError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "Request body exceeds the allowed size")
	} else {
		respondError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON request body")
	}
	return false
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, map[string]interface{}{
		"error":   code,
		"message": message,
	})
}
