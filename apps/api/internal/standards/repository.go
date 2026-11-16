package standards

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository manages reviewed standards records.
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetAuthoritativeStandards returns only records stored in the reviewed standards register.
func (r *Repository) GetAuthoritativeStandards(ctx context.Context) ([]Standard, error) {
	if r.pool == nil {
		return []Standard{}, nil
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, authority, document_title, version, publication_date,
		       legal_status, clause_ref, clause_title, requirement,
		       interpretation, software_rule, reviewed_by, reviewed_at
		FROM standards
		ORDER BY document_title, clause_ref
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Standard, 0)
	for rows.Next() {
		var standard Standard
		if err := rows.Scan(
			&standard.ID, &standard.Authority, &standard.DocumentTitle, &standard.Version, &standard.PublicationAt,
			&standard.LegalStatus, &standard.ClauseRef, &standard.ClauseTitle, &standard.Requirement,
			&standard.Interpretation, &standard.SoftwareRule, &standard.ReviewedBy, &standard.ReviewedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, standard)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
