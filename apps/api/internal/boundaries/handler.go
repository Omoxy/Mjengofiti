package boundaries

import (
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mjengofiti/api/migrations"
)

// Handler serves administrative boundary geometries.
type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/api/v1/boundaries", h.HandleNairobiBoundary)
}

// HandleNairobiBoundary returns the verified geoBoundaries Nairobi ADM1 feature.
func (h *Handler) HandleNairobiBoundary(w http.ResponseWriter, r *http.Request) {
	boundaryGeoJSON, err := fs.ReadFile(migrations.FS, "nairobi_county.geojson")
	if err != nil {
		http.Error(w, "Nairobi boundary data unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/geo+json")
	_, _ = w.Write(boundaryGeoJSON)
}
