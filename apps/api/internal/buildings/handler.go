package buildings

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// Handler provides HTTP handlers for building operations.
type Handler struct {
	repo   *Repository
	logger *slog.Logger
}

// NewHandler creates a new buildings HTTP handler.
func NewHandler(repo *Repository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

// Routes registers building routes on the given router.
func (h *Handler) Routes(r chi.Router) {
	r.Get("/api/v1/buildings", h.List)
	r.Get("/api/v1/buildings/{id}", h.GetByID)
	r.Get("/api/v1/buildings/{id}/indicators", h.GetIndicators)
	r.Get("/api/v1/buildings/{id}/evidence", h.GetEvidence)
}

// List returns buildings filtered by bounding box or search text.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	// Check for bbox query
	bboxStr := r.URL.Query().Get("bbox")
	if bboxStr != "" {
		h.listByBBox(w, r, bboxStr)
		return
	}

	// Text search
	q := SearchQuery{
		Text:   r.URL.Query().Get("q"),
		Limit:  parseIntParam(r, "limit", DefaultLimit),
		Offset: parseIntParam(r, "offset", 0),
	}

	if src := r.URL.Query().Get("source"); src != "" {
		q.SourceName = &src
	}

	result, err := h.repo.Search(r.Context(), q)
	if err != nil {
		h.logger.Error("building search failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "search_failed", "Failed to search buildings")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) listByBBox(w http.ResponseWriter, r *http.Request, bboxStr string) {
	coords, err := parseBBox(bboxStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_bbox", "Expected bbox=minLng,minLat,maxLng,maxLat")
		return
	}

	q := BBoxQuery{
		MinLng: coords[0],
		MinLat: coords[1],
		MaxLng: coords[2],
		MaxLat: coords[3],
		Limit:  parseIntParam(r, "limit", DefaultLimit),
		Offset: parseIntParam(r, "offset", 0),
	}

	// Validate bbox size to prevent excessive queries
	lngSpan := q.MaxLng - q.MinLng
	latSpan := q.MaxLat - q.MinLat
	const maxSpan = 1.0 // roughly 111km at equator
	if lngSpan > maxSpan || latSpan > maxSpan {
		writeError(w, http.StatusBadRequest, "bbox_too_large", "Bounding box exceeds maximum allowed area. Zoom in further.")
		return
	}

	result, err := h.repo.FindByBBox(r.Context(), q)
	if err != nil {
		h.logger.Error("bbox query failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "query_failed", "Failed to query buildings in area")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetByID returns a single building with full detail.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "Building ID is required")
		return
	}

	building, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		h.logger.Error("building not found", slog.String("id", id), slog.String("error", err.Error()))
		writeError(w, http.StatusNotFound, "not_found", "Building not found")
		return
	}

	writeJSON(w, http.StatusOK, building)
}

// GetIndicators returns calculated indicators for a building.
func (h *Handler) GetIndicators(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "Building ID is required")
		return
	}

	building, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Building not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"building_id": id,
		"indicators":  building.Indicators,
	})
}

// GetEvidence returns sourced observations for a building.
func (h *Handler) GetEvidence(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "Building ID is required")
		return
	}

	building, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Building not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"building_id":  id,
		"observations": building.Observations,
	})
}

// parseBBox parses "minLng,minLat,maxLng,maxLat" into four float64 values.
func parseBBox(s string) ([4]float64, error) {
	var result [4]float64
	parts := splitComma(s)
	if len(parts) != 4 {
		return result, fmt.Errorf("expected 4 bbox values, got %d", len(parts))
	}
	for i, p := range parts {
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return result, fmt.Errorf("invalid coordinate at position %d: %w", i, err)
		}
		result[i] = v
	}
	return result, nil
}

func splitComma(s string) []string {
	var parts []string
	current := ""
	for _, c := range s {
		if c == ',' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	parts = append(parts, current)
	return parts
}

func parseIntParam(r *http.Request, key string, defaultVal int) int {
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

type apiError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(apiError{Error: code, Message: message})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
