package inspections

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mjengofiti/api/internal/auth"
)

// Handler provides HTTP handlers for inspection operations.
type Handler struct {
	repo   *Repository
	logger *slog.Logger
}

// NewHandler creates a new inspections HTTP handler.
func NewHandler(repo *Repository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

// Routes registers inspection routes on the given router.
func (h *Handler) Routes(r chi.Router) {
	r.With(auth.RequirePermission("inspections:read")).Get("/api/v1/inspection-cases", h.List)
	r.With(auth.RequirePermission("inspections:triage")).Post("/api/v1/inspection-cases", h.Create)
	r.With(auth.RequirePermission("inspections:assign")).Put("/api/v1/inspection-cases/{id}/assignment", h.Assign)
	r.With(auth.RequirePermission("inspections:read")).Get("/api/v1/inspection-cases/{id}", h.GetByID)
	r.With(auth.RequirePermission("inspections:triage")).Patch("/api/v1/inspection-cases/{id}", h.UpdateStatus)
	r.With(auth.RequirePermission("inspections:read")).Get("/api/v1/inspection-cases/{id}/history", h.GetHistory)
}

// List returns inspection cases with optional status filter.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit := intParam(r, "limit", 50)
	offset := intParam(r, "offset", 0)
	statusFilter := r.URL.Query().Get("status")

	var statusPtr *string
	if statusFilter != "" {
		statusPtr = &statusFilter
	}

			user := auth.UserFromContext(r.Context())
			if user == nil {
				respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required")
				return
			}
			var cases []Case
			var total int
			var err error
			if user.Role == auth.RoleInspector {
				cases, total, err = h.repo.ListAssigned(r.Context(), statusPtr, user.ID, limit, offset)
			} else {
				cases, total, err = h.repo.List(r.Context(), statusPtr, limit, offset)
			}
	if err != nil {
		h.logger.Error("listing cases failed", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "list_failed", "Failed to list inspection cases")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"cases":  cases,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// CreateCaseRequest is the request body for creating a new case.
type CreateCaseRequest struct {
	BuildingID     string   `json:"building_id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Priority       string   `json:"priority"`
	ReferralReason string   `json:"referral_reason"`
	EvidenceRefs   []string `json:"evidence_refs"`
}

// Create creates a new inspection case.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required")
		return
	}
	var req CreateCaseRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if req.BuildingID == "" || req.Title == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "building_id and title are required")
		return
	}

	// Validate priority
	switch req.Priority {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical:
		// valid
	default:
		req.Priority = PriorityMedium
	}

	c := &Case{
		BuildingID:     req.BuildingID,
		Title:          req.Title,
		Description:    req.Description,
		Priority:       req.Priority,
		ReferralReason: req.ReferralReason,
		EvidenceRefs:   req.EvidenceRefs,
		CreatedBy:      user.ID,
	}

	if err := h.repo.Create(r.Context(), c); err != nil {
		h.logger.Error("creating case failed", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "create_failed", "Failed to create inspection case")
		return
	}

	respondJSON(w, http.StatusCreated, c)
}

type AssignCaseRequest struct {
	AssignedTo string `json:"assigned_to"`
}

func (h *Handler) Assign(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required")
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "Case ID is required")
		return
	}
	var request AssignCaseRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&request); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}
	if err := h.repo.Assign(r.Context(), id, request.AssignedTo, user.ID); err != nil {
		respondError(w, http.StatusConflict, "assignment_rejected", "Case assignment was rejected")
		return
	}
	caseRecord, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "fetch_failed", "Assignment saved but case could not be retrieved")
		return
	}
	respondJSON(w, http.StatusOK, caseRecord)
}

// GetByID returns a single inspection case.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "Case ID is required")
		return
	}

	c, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Inspection case not found")
		return
	}
		if user := auth.UserFromContext(r.Context()); user != nil && user.Role == auth.RoleInspector && (c.AssignedTo == nil || *c.AssignedTo != user.ID) {
			respondError(w, http.StatusNotFound, "not_found", "Inspection case not found")
			return
		}

	respondJSON(w, http.StatusOK, c)
}

// UpdateStatusRequest is the request body for changing case status.
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// UpdateStatus changes the status of an inspection case.
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "Case ID is required")
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if req.Status == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "status is required")
		return
	}

			user := auth.UserFromContext(r.Context())
			if user == nil {
				respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required")
				return
			}
			err := h.repo.UpdateStatus(r.Context(), id, req.Status, user.ID)
	if err != nil {
		h.logger.Error("updating case status failed", slog.String("case_id", id), slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "update_failed", err.Error())
		return
	}

	// Return updated case
	c, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "fetch_failed", "Status updated but failed to fetch case")
		return
	}

	respondJSON(w, http.StatusOK, c)
}

// GetHistory returns the audit history for a case.
func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "Case ID is required")
		return
	}
		if user := auth.UserFromContext(r.Context()); user != nil && user.Role == auth.RoleInspector {
			caseRecord, err := h.repo.FindByID(r.Context(), id)
			if err != nil || caseRecord.AssignedTo == nil || *caseRecord.AssignedTo != user.ID {
				respondError(w, http.StatusNotFound, "not_found", "Inspection case not found")
				return
			}
		}

	history, err := h.repo.GetHistory(r.Context(), id)
	if err != nil {
		h.logger.Error("getting case history failed", slog.String("case_id", id), slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "history_failed", "Failed to retrieve case history")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"case_id": id,
		"history": history,
	})
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
