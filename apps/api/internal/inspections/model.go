package inspections

import (
	"time"
)

// Case represents an inspection case.
type Case struct {
	ID              string     `json:"id"`
	BuildingID      string     `json:"building_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Priority        string     `json:"priority"`
	Status          string     `json:"status"`
	ReferralReason  string     `json:"referral_reason"`
	EvidenceRefs    []string   `json:"evidence_refs"`
	AssignedTo      *string    `json:"assigned_to,omitempty"`
	AssignedAt      *time.Time `json:"assigned_at,omitempty"`
	DueDate         *time.Time `json:"due_date,omitempty"`
	CreatedBy       string     `json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	ClosedAt        *time.Time `json:"closed_at,omitempty"`
}

// Finding represents a professional inspection finding.
type Finding struct {
	ID          string    `json:"id"`
	CaseID      string    `json:"case_id"`
	FindingType string    `json:"finding_type"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	RecordedBy  string    `json:"recorded_by"`
	RecordedAt  time.Time `json:"recorded_at"`
	Attachments []string  `json:"attachments,omitempty"`
}

// CaseHistory records a change to an inspection case.
type CaseHistory struct {
	ID         string    `json:"id"`
	CaseID     string    `json:"case_id"`
	Action     string    `json:"action"`
	FromStatus *string   `json:"from_status,omitempty"`
	ToStatus   *string   `json:"to_status,omitempty"`
	ChangedBy  string    `json:"changed_by"`
	ChangedAt  time.Time `json:"changed_at"`
	Details    *string   `json:"details,omitempty"`
}

// Assignment represents a case assignment to a person.
type Assignment struct {
	CaseID     string    `json:"case_id"`
	AssignedTo string    `json:"assigned_to"`
	AssignedBy string    `json:"assigned_by"`
	AssignedAt time.Time `json:"assigned_at"`
	Notes      *string   `json:"notes,omitempty"`
}

// Case statuses and allowed transitions.
const (
	StatusOpen        = "open"
	StatusTriaged     = "triaged"
	StatusAssigned    = "assigned"
	StatusInProgress  = "in_progress"
	StatusPendingReview = "pending_review"
	StatusClosed      = "closed"
	StatusReopened    = "reopened"
)

// AllowedTransitions defines which status transitions are valid.
var AllowedTransitions = map[string][]string{
	StatusOpen:          {StatusTriaged, StatusClosed},
	StatusTriaged:       {StatusAssigned, StatusClosed},
	StatusAssigned:      {StatusInProgress, StatusClosed},
	StatusInProgress:    {StatusPendingReview, StatusClosed},
	StatusPendingReview: {StatusClosed, StatusReopened},
	StatusClosed:        {StatusReopened},
	StatusReopened:      {StatusTriaged, StatusAssigned},
}

// IsValidTransition checks whether a status transition is allowed.
func IsValidTransition(from, to string) bool {
	allowed, ok := AllowedTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// Priority levels.
const (
	PriorityLow      = "low"
	PriorityMedium   = "medium"
	PriorityHigh     = "high"
	PriorityCritical = "critical"
)
