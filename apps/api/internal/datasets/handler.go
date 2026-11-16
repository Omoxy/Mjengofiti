package datasets

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mjengofiti/api/internal/auth"
)

// Handler provides HTTP handlers for dataset operations.
type Handler struct {
	repo   *Repository
	logger *slog.Logger
}

// NewHandler creates a new datasets HTTP handler.
func NewHandler(repo *Repository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

// Routes registers dataset routes on the given router.
func (h *Handler) Routes(r chi.Router) {
	r.Get("/api/v1/datasets", h.List)
	r.Get("/api/v1/datasets/{id}", h.GetByID)
	r.With(auth.RequirePermission("datasets:read")).Get("/api/v1/datasets/{id}/versions", h.ListVersions)
	r.With(auth.RequirePermission("datasets:write")).Get("/api/v1/jobs/{id}", h.GetJob)
}

// List returns all registered datasets.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit := intParam(r, "limit", 50)
	offset := intParam(r, "offset", 0)

	datasets, total, err := h.repo.List(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("listing datasets failed", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "list_failed", "Failed to list datasets")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"datasets": datasets,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetByID returns a single dataset by ID.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "Dataset ID is required")
		return
	}

	dataset, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Dataset not found")
		return
	}

	respondJSON(w, http.StatusOK, dataset)
}

// ListVersions returns all versions of a dataset.
func (h *Handler) ListVersions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "Dataset ID is required")
		return
	}

	versions, err := h.repo.ListVersions(r.Context(), id)
	if err != nil {
		h.logger.Error("listing versions failed", slog.String("dataset_id", id), slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "list_failed", "Failed to list dataset versions")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"dataset_id": id,
		"versions":   versions,
	})
}

// GetJob returns the status of an ingestion job.
func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "Job ID is required")
		return
	}

	job, err := h.repo.FindJobByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Ingestion job not found")
		return
	}

	respondJSON(w, http.StatusOK, job)
}

func intParam(r *http.Request, key string, defaultVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return defaultVal
	}
	return v
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": code, "message": message})
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
