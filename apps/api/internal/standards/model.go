package standards

import "time"

// Standard represents a verified Kenyan regulatory or building code clause.
type Standard struct {
	ID             string    `json:"id"`
	Authority      string    `json:"authority"`
	DocumentTitle  string    `json:"document_title"`
	Version        string    `json:"version"`
	PublicationAt  string    `json:"publication_date"`
	LegalStatus    string    `json:"legal_status"` // statutory, code, draft, guideline
	ClauseRef      string    `json:"clause_ref"`
	ClauseTitle    string    `json:"clause_title"`
	Requirement    string    `json:"requirement"`
	Interpretation string    `json:"interpretation"`
	SoftwareRule   string    `json:"software_rule"`
	ReviewedBy     string    `json:"reviewed_by"`
	ReviewedAt     time.Time `json:"reviewed_at"`
}

// ListResponse returns a list of verified statutory standards.
type ListResponse struct {
	Standards []Standard `json:"standards"`
	Total     int        `json:"total"`
	Message   string     `json:"message,omitempty"`
}
