package audit

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository manages audit trail records.
type Repository struct {
	pool    *pgxpool.Pool
	mu      sync.RWMutex
	inMem   []Event
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	repo := &Repository{
		pool: pool,
	}
	return repo
}

// Log records an audit event.
func (r *Repository) Log(ctx context.Context, e Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	r.inMem = append([]Event{e}, r.inMem...)
	return nil
}

// List returns recent audit events.
func (r *Repository) List(ctx context.Context, limit, offset int) (*ListResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	total := len(r.inMem)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}

	return &ListResponse{
		Events: r.inMem[offset:end],
		Total:  total,
	}, nil
}
