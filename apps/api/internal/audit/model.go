package audit

import "time"

// Event represents an immutable audit log entry.
type Event struct {
	ID           string         `json:"id"`
	UserID       *string        `json:"user_id,omitempty"`
	UserEmail    string         `json:"user_email"`
	UserRole     string         `json:"user_role"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	Details      map[string]any `json:"details,omitempty"`
	IPAddress    string         `json:"ip_address"`
	CreatedAt    time.Time      `json:"created_at"`
}

// ListResponse returns a paginated list of audit events.
type ListResponse struct {
	Events []Event `json:"events"`
	Total  int     `json:"total"`
}
