package indicators

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mjengofiti/api/internal/auth"
)

// Handler exposes indicator endpoints.
type Handler struct {
	repo   *Repository
	logger *slog.Logger
}

func NewHandler(repo *Repository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/api/v1/indicators", h.HandleList)
	r.Get("/api/v1/indicators/{id}", h.HandleGet)
	r.With(auth.RequirePermission("indicators:compute")).Post("/api/v1/analysis-jobs", h.HandleCreateJob)
	r.With(auth.RequirePermission("indicators:compute")).Get("/api/v1/analysis-jobs/{id}", h.HandleGetJob)
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	defs, err := h.repo.GetDefinitions(r.Context())
	if err != nil {
		h.logger.Error("failed to get indicator definitions", slog.String("error", err.Error()))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed_retrieving_indicators"})
		return
	}

	resp := ListDefinitionsResponse{
		Definitions: defs,
		Total:       len(defs),
		Disclaimer:  "All indicators are provisional screening indicators for physical inspection triage. They do not constitute structural engineering certifications or collapse determinations.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	def, err := h.repo.GetDefinitionByName(r.Context(), id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "indicator_not_found", "message": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(def)
}

func (h *Handler) HandleCreateJob(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   "analysis_unavailable",
		"message": "No analysis calculation worker is implemented.",
	})
}

func (h *Handler) HandleGetJob(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   "analysis_unavailable",
		"message": "No analysis job store is implemented.",
	})
}
