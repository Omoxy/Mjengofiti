package inspections

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides data access for inspection cases.
type Repository struct {
	db          *pgxpool.Pool
	mu          sync.RWMutex
	inMemCases  map[string]*Case
	inMemHist   map[string][]CaseHistory
}

// NewRepository creates a new inspections repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db:         db,
		inMemCases: make(map[string]*Case),
		inMemHist:  make(map[string][]CaseHistory),
	}
}

// Create inserts a new inspection case within a transaction or in-memory store.
func (r *Repository) Create(ctx context.Context, c *Case) error {
	c.ID = uuid.New().String()
	c.Status = StatusOpen
	c.CreatedAt = time.Now().UTC()
	c.UpdatedAt = c.CreatedAt

	if r.db != nil {
		tx, err := r.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("beginning transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, `
			INSERT INTO inspection_cases
				(id, building_id, title, description, priority, status,
				 referral_reason, evidence_refs, created_by, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`, c.ID, c.BuildingID, c.Title, c.Description, c.Priority, c.Status,
			c.ReferralReason, c.EvidenceRefs, c.CreatedBy, c.CreatedAt, c.UpdatedAt)
		if err != nil {
			return fmt.Errorf("inserting case: %w", err)
		}

		// Record history
		_, err = tx.Exec(ctx, `
			INSERT INTO case_history (id, case_id, action, to_status, changed_by, changed_at, details)
			VALUES ($1, $2, 'created', $3, $4, $5, 'Case created')
		`, uuid.New().String(), c.ID, c.Status, c.CreatedBy, c.CreatedAt)
		if err != nil {
			return fmt.Errorf("recording case history: %w", err)
		}

		return tx.Commit(ctx)
	}

	// In-memory fallback
	r.mu.Lock()
	defer r.mu.Unlock()
	caseCopy := *c
	r.inMemCases[c.ID] = &caseCopy
	details := "Case created"
	r.inMemHist[c.ID] = append(r.inMemHist[c.ID], CaseHistory{
		ID:        uuid.New().String(),
		CaseID:    c.ID,
		Action:    "created",
		ToStatus:  &c.Status,
		ChangedBy: c.CreatedBy,
		ChangedAt: c.CreatedAt,
		Details:   &details,
	})
	return nil
}

// FindByID retrieves a case by ID.
func (r *Repository) FindByID(ctx context.Context, id string) (*Case, error) {
	if r.db != nil {
		var c Case
		err := r.db.QueryRow(ctx, `
			SELECT id, building_id, title, description, priority, status,
			       referral_reason, evidence_refs, assigned_to, assigned_at,
			       due_date, created_by, created_at, updated_at, closed_at
			FROM inspection_cases
			WHERE id = $1
		`, id).Scan(
			&c.ID, &c.BuildingID, &c.Title, &c.Description, &c.Priority, &c.Status,
			&c.ReferralReason, &c.EvidenceRefs, &c.AssignedTo, &c.AssignedAt,
			&c.DueDate, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt, &c.ClosedAt,
		)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, fmt.Errorf("case not found: %s", id)
			}
			return nil, fmt.Errorf("finding case %s: %w", id, err)
		}
		return &c, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.inMemCases[id]
	if !ok {
		return nil, fmt.Errorf("case not found: %s", id)
	}
	caseCopy := *c
	return &caseCopy, nil
}

// UpdateStatus transitions a case to a new status with history.
// Assign records an inspector assignment and advances an eligible case to assigned.
func (r *Repository) Assign(ctx context.Context, id, assignee, assignedBy string) error {
	if strings.TrimSpace(assignee) == "" || strings.TrimSpace(assignedBy) == "" {
		return fmt.Errorf("assignee and assigning user are required")
	}
	if len(assignee) > 255 || len(assignedBy) > 255 {
		return fmt.Errorf("assignment identifiers exceed the maximum length")
	}

	now := time.Now().UTC()
	if r.db != nil {
		tx, err := r.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("beginning assignment: %w", err)
		}
		defer tx.Rollback(ctx)

		var currentStatus string
		if err := tx.QueryRow(ctx, "SELECT status FROM inspection_cases WHERE id = $1 FOR UPDATE", id).Scan(&currentStatus); err != nil {
			return fmt.Errorf("finding case for assignment: %w", err)
		}
		if !IsValidTransition(currentStatus, StatusAssigned) {
			return fmt.Errorf("case cannot be assigned from status %s", currentStatus)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE inspection_cases
			SET status = $1, assigned_to = $2, assigned_at = $3, updated_at = $3
			WHERE id = $4
		`, StatusAssigned, assignee, now, id); err != nil {
			return fmt.Errorf("updating case assignment: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO case_history (id, case_id, action, from_status, to_status, changed_by, changed_at, details)
			VALUES ($1, $2, 'assigned', $3, $4, $5, $6, $7)
		`, uuid.New().String(), id, currentStatus, StatusAssigned, assignedBy, now, "Assigned to "+assignee); err != nil {
			return fmt.Errorf("recording case assignment: %w", err)
		}
		return tx.Commit(ctx)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	caseRecord, ok := r.inMemCases[id]
	if !ok {
		return fmt.Errorf("case not found: %s", id)
	}
	if !IsValidTransition(caseRecord.Status, StatusAssigned) {
		return fmt.Errorf("case cannot be assigned from status %s", caseRecord.Status)
	}
	fromStatus := caseRecord.Status
	caseRecord.Status = StatusAssigned
	caseRecord.AssignedTo = &assignee
	caseRecord.AssignedAt = &now
	caseRecord.UpdatedAt = now
	details := "Assigned to " + assignee
	r.inMemHist[id] = append(r.inMemHist[id], CaseHistory{
		ID: uuid.New().String(), CaseID: id, Action: "assigned", FromStatus: &fromStatus,
		ToStatus: &caseRecord.Status, ChangedBy: assignedBy, ChangedAt: now, Details: &details,
	})
	return nil
}

// UpdateStatus transitions a case to a new status with history.
func (r *Repository) UpdateStatus(ctx context.Context, id, newStatus, changedBy string) error {
	if newStatus == StatusAssigned {
		return fmt.Errorf("use Assign to set the assignee and assigned status together")
	}
	if r.db != nil {
		tx, err := r.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("beginning transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		var currentStatus string
		err = tx.QueryRow(ctx, "SELECT status FROM inspection_cases WHERE id = $1 FOR UPDATE", id).Scan(&currentStatus)
		if err != nil {
			return fmt.Errorf("locking case: %w", err)
		}

		if !IsValidTransition(currentStatus, newStatus) {
			return fmt.Errorf("invalid transition from %s to %s", currentStatus, newStatus)
		}

		now := time.Now().UTC()

		updateQuery := `UPDATE inspection_cases SET status = $1, updated_at = $2`
		args := []any{newStatus, now}
		argIdx := 3

		if newStatus == StatusClosed {
			updateQuery += fmt.Sprintf(", closed_at = $%d", argIdx)
			args = append(args, now)
			argIdx++
		}

		updateQuery += fmt.Sprintf(" WHERE id = $%d", argIdx)
		args = append(args, id)

		_, err = tx.Exec(ctx, updateQuery, args...)
		if err != nil {
			return fmt.Errorf("updating case status: %w", err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO case_history (id, case_id, action, from_status, to_status, changed_by, changed_at)
			VALUES ($1, $2, 'status_change', $3, $4, $5, $6)
		`, uuid.New().String(), id, currentStatus, newStatus, changedBy, now)
		if err != nil {
			return fmt.Errorf("recording status change history: %w", err)
		}

		return tx.Commit(ctx)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.inMemCases[id]
	if !ok {
		return fmt.Errorf("case not found: %s", id)
	}

	if !IsValidTransition(c.Status, newStatus) {
		return fmt.Errorf("invalid transition from %s to %s", c.Status, newStatus)
	}

	fromStatus := c.Status
	now := time.Now().UTC()
	c.Status = newStatus
	c.UpdatedAt = now
	if newStatus == StatusClosed {
		c.ClosedAt = &now
	}

	r.inMemHist[id] = append(r.inMemHist[id], CaseHistory{
		ID:         uuid.New().String(),
		CaseID:     id,
		Action:     "status_change",
		FromStatus: &fromStatus,
		ToStatus:   &newStatus,
		ChangedBy:  changedBy,
		ChangedAt:  now,
	})
	return nil
}

// List returns inspection cases with optional status filter.
func (r *Repository) List(ctx context.Context, status *string, limit, offset int) ([]Case, int, error) {
	return r.list(ctx, status, "", limit, offset)
}

func (r *Repository) ListAssigned(ctx context.Context, status *string, assignee string, limit, offset int) ([]Case, int, error) {
	if assignee == "" {
		return []Case{}, 0, nil
	}
	return r.list(ctx, status, assignee, limit, offset)
}

func (r *Repository) list(ctx context.Context, status *string, assignedTo string, limit, offset int) ([]Case, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	if r.db != nil {
                filters := make([]string, 0, 2)
		var args []any
		argIdx := 1

		if status != nil && *status != "" {
                        filters = append(filters, fmt.Sprintf("status = $%d", argIdx))
			args = append(args, *status)
			argIdx++
		}
			if assignedTo != "" {
				filters = append(filters, fmt.Sprintf("assigned_to = $%d", argIdx))
				args = append(args, assignedTo)
				argIdx++
			}
			where := ""
			if len(filters) > 0 {
				where = "WHERE " + strings.Join(filters, " AND ")
			}

		var total int
		countQ := fmt.Sprintf("SELECT COUNT(*) FROM inspection_cases %s", where)
		if err := r.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("counting cases: %w", err)
		}

		selectQ := fmt.Sprintf(`
			SELECT id, building_id, title, description, priority, status,
			       referral_reason, evidence_refs, assigned_to, assigned_at,
			       due_date, created_by, created_at, updated_at, closed_at
			FROM inspection_cases %s
			ORDER BY created_at DESC
			LIMIT $%d OFFSET $%d
		`, where, argIdx, argIdx+1)
		args = append(args, limit, offset)

		rows, err := r.db.Query(ctx, selectQ, args...)
		if err != nil {
			return nil, 0, fmt.Errorf("listing cases: %w", err)
		}
		defer rows.Close()

		var cases []Case
		for rows.Next() {
			var c Case
			if err := rows.Scan(
				&c.ID, &c.BuildingID, &c.Title, &c.Description, &c.Priority, &c.Status,
				&c.ReferralReason, &c.EvidenceRefs, &c.AssignedTo, &c.AssignedAt,
				&c.DueDate, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt, &c.ClosedAt,
			); err != nil {
				return nil, 0, fmt.Errorf("scanning case: %w", err)
			}
			cases = append(cases, c)
		}

		if cases == nil {
			cases = []Case{}
		}
		return cases, total, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	var filtered []Case
	for _, c := range r.inMemCases {
		statusMatches := status == nil || *status == "" || c.Status == *status
			assignedMatches := assignedTo == "" || (c.AssignedTo != nil && *c.AssignedTo == assignedTo)
			if statusMatches && assignedMatches {
			filtered = append(filtered, *c)
		}
	}
	total := len(filtered)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return filtered[offset:end], total, nil
}

// GetHistory retrieves the audit history for a case.
func (r *Repository) GetHistory(ctx context.Context, caseID string) ([]CaseHistory, error) {
	if r.db != nil {
		rows, err := r.db.Query(ctx, `
			SELECT id, case_id, action, from_status, to_status, changed_by, changed_at, details
			FROM case_history
			WHERE case_id = $1
			ORDER BY changed_at ASC
		`, caseID)
		if err != nil {
			return nil, fmt.Errorf("listing case history: %w", err)
		}
		defer rows.Close()

		var history []CaseHistory
		for rows.Next() {
			var h CaseHistory
			if err := rows.Scan(
				&h.ID, &h.CaseID, &h.Action, &h.FromStatus, &h.ToStatus,
				&h.ChangedBy, &h.ChangedAt, &h.Details,
			); err != nil {
				return nil, fmt.Errorf("scanning history: %w", err)
			}
			history = append(history, h)
		}

		if history == nil {
			history = []CaseHistory{}
		}
		return history, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	hist := r.inMemHist[caseID]
	if hist == nil {
		return []CaseHistory{}, nil
	}
	return hist, nil
}
