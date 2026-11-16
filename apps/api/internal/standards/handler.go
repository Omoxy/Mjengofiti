package standards

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for Kenyan standards.
type Handler struct {
	repo   *Repository
	logger *slog.Logger
}

func NewHandler(repo *Repository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/api/v1/standards", h.HandleList)
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	standards, err := h.repo.GetAuthoritativeStandards(r.Context())
	if err != nil {
		h.logger.Error("failed to retrieve standards", slog.String("error", err.Error()))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed_retrieving_standards"})
		return
	}

	resp := ListResponse{
		Standards: standards,
		Total:     len(standards),
				Message:   "Standards records are shown only when reviewed entries are available.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
