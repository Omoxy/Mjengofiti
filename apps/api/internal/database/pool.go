package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mjengofiti/api/internal/config"
)

// Pool wraps a pgx connection pool with health check capabilities.
type Pool struct {
	*pgxpool.Pool
}

// Connect creates a new database connection pool.
func Connect(ctx context.Context, cfg config.DatabaseConfig) (*Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parsing database config: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.MaxConns)
	poolCfg.MinConns = int32(cfg.MinConns)
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute
	poolCfg.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return &Pool{Pool: pool}, nil
}

// HealthCheck verifies database connectivity and PostGIS availability.
func (p *Pool) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var postgisVersion string
	err := p.QueryRow(ctx, "SELECT PostGIS_Version()").Scan(&postgisVersion)
	if err != nil {
		return fmt.Errorf("PostGIS health check failed: %w", err)
	}
	return nil
}

// PostGISVersion returns the installed PostGIS version string.
func (p *Pool) PostGISVersion(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var version string
	err := p.QueryRow(ctx, "SELECT PostGIS_Full_Version()").Scan(&version)
	if err != nil {
		return "", fmt.Errorf("querying PostGIS version: %w", err)
	}
	return version, nil
}
