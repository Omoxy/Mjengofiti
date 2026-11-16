package auth

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handler provides the current authenticated profile and public role metadata.
type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/api/v1/auth/me", h.HandleMe)
	r.Get("/api/v1/auth/roles", h.HandleRoles)
}

func (h *Handler) HandleMe(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		writeAuthError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required to access this resource")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(user)
}

func (h *Handler) HandleRoles(w http.ResponseWriter, r *http.Request) {
	roles := []RoleDefinition{
		{
			Role:        RoleViewer,
			Title:       "Read-only Viewer",
			Description: "Read access to public building, dataset, indicator, and standards information.",
			Permissions: []string{"buildings:read", "datasets:read", "standards:read", "indicators:read"},
		},
		{
			Role:        RoleDataResearcher,
			Title:       "Data Researcher / GIS Analyst",
			Description: "Reviews dataset provenance and indicator methods and runs approved analysis jobs.",
			Permissions: []string{"buildings:read", "datasets:read", "indicators:read", "indicators:compute", "standards:read"},
		},
		{
			Role:        RoleDataSteward,
			Title:       "Geospatial Data Steward",
			Description: "Registers, validates, and promotes dataset versions.",
			Permissions: []string{"buildings:read", "datasets:read", "datasets:write", "datasets:promote", "standards:read", "audit:read"},
		},
		{
			Role:        RoleInspector,
			Title:       "Building Inspector",
			Description: "Reads assigned inspection cases and submits professional findings.",
			Permissions: []string{"buildings:read", "inspections:read", "inspections:findings", "standards:read"},
		},
		{
			Role:        RoleInspectionSupervisor,
			Title:       "Inspection Supervisor",
			Description: "Reviews, triages, assigns, and closes inspection cases.",
			Permissions: []string{"buildings:read", "inspections:read", "inspections:triage", "inspections:assign", "inspections:findings", "inspections:close", "standards:read", "audit:read"},
		},
		{
			Role:        RoleAdmin,
			Title:       "System Administrator",
			Description: "System administration and security operations.",
			Permissions: []string{"*"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"roles": roles,
		"total": len(roles),
	})
}
