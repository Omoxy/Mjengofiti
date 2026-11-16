package database

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var migrationFilePattern = regexp.MustCompile(`^(\d+)_([a-zA-Z0-9_\-]+)\.(up|down)\.sql$`)

// Migration represents a single database migration script.
type Migration struct {
	Version   int64
	Name      string
	Direction string // "up" or "down"
	Filename  string
	Content   string
	Checksum  string
}

// MigrationStatus reports the application state of a migration.
type MigrationStatus struct {
	Version   int64      `json:"version"`
	Name      string     `json:"name"`
	Applied   bool       `json:"applied"`
	AppliedAt *time.Time `json:"applied_at,omitempty"`
	Checksum  string     `json:"checksum,omitempty"`
}

// Migrator manages and applies database migrations.
type Migrator struct {
	pool *pgxpool.Pool
	fsys fs.FS
}

// NewMigrator returns a new Migrator instance.
func NewMigrator(pool *pgxpool.Pool, fsys fs.FS) *Migrator {
	return &Migrator{
		pool: pool,
		fsys: fsys,
	}
}

// EnsureSchemaMigrationsTable creates the migration tracking table if it does not exist.
func (m *Migrator) EnsureSchemaMigrationsTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version BIGINT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		checksum VARCHAR(64) NOT NULL
	);`
	_, err := m.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}
	return nil
}

// LoadMigrations scans the filesystem for valid migration files.
func (m *Migrator) LoadMigrations() (map[int64]*Migration, map[int64]*Migration, error) {
	upMigrations := make(map[int64]*Migration)
	downMigrations := make(map[int64]*Migration)

	entries, err := fs.ReadDir(m.fsys, ".")
	if err != nil {
		return nil, nil, fmt.Errorf("reading migrations directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := migrationFilePattern.FindStringSubmatch(entry.Name())
		if len(matches) != 4 {
			continue
		}

		version, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing migration version from %s: %w", entry.Name(), err)
		}

		name := matches[2]
		direction := matches[3]

		contentBytes, err := fs.ReadFile(m.fsys, entry.Name())
		if err != nil {
			return nil, nil, fmt.Errorf("reading migration file %s: %w", entry.Name(), err)
		}

		hasher := sha256.New()
		hasher.Write(contentBytes)
		checksum := hex.EncodeToString(hasher.Sum(nil))

		mig := &Migration{
			Version:   version,
			Name:      name,
			Direction: direction,
			Filename:  entry.Name(),
			Content:   string(contentBytes),
			Checksum:  checksum,
		}

		if direction == "up" {
			upMigrations[version] = mig
		} else if direction == "down" {
			downMigrations[version] = mig
		}
	}

	return upMigrations, downMigrations, nil
}

// Up applies all pending forward migrations in ascending version order.
func (m *Migrator) Up(ctx context.Context) ([]int64, error) {
	if err := m.EnsureSchemaMigrationsTable(ctx); err != nil {
		return nil, err
	}

	upMigrations, _, err := m.LoadMigrations()
	if err != nil {
		return nil, err
	}

	rows, err := m.pool.Query(ctx, "SELECT version, checksum FROM schema_migrations ORDER BY version ASC")
	if err != nil {
		return nil, fmt.Errorf("querying applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int64]string)
	for rows.Next() {
		var v int64
		var chk string
		if err := rows.Scan(&v, &chk); err != nil {
			return nil, fmt.Errorf("scanning applied migration: %w", err)
		}
		applied[v] = chk
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("iterating applied migrations: %w", rows.Err())
	}

	// Sort available versions
	versions := make([]int64, 0, len(upMigrations))
	for v := range upMigrations {
		versions = append(versions, v)
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })

	var appliedNow []int64
	for _, v := range versions {
		if _, ok := applied[v]; ok {
			continue // already applied
		}

		mig := upMigrations[v]
		if err := m.applyMigration(ctx, mig); err != nil {
			return appliedNow, fmt.Errorf("applying migration %s: %w", mig.Filename, err)
		}
		appliedNow = append(appliedNow, v)
	}

	return appliedNow, nil
}

func (m *Migrator) applyMigration(ctx context.Context, mig *Migration) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) // safe if already committed

	var args []any
	if mig.Name == "nairobi_county_boundary" {
		payload, err := fs.ReadFile(m.fsys, "nairobi_county.geojson")
		if err != nil {
			return fmt.Errorf("reading verified Nairobi boundary asset: %w", err)
		}
		var source struct {
			Type     string `json:"type"`
			Features []struct {
				Properties struct {
					Name      string `json:"shapeName"`
					Group     string `json:"shapeGroup"`
					AdminType string `json:"shapeType"`
				} `json:"properties"`
			} `json:"features"`
		}
		if err := json.Unmarshal(payload, &source); err != nil {
			return fmt.Errorf("decoding verified Nairobi boundary asset: %w", err)
		}
		if source.Type != "FeatureCollection" || len(source.Features) != 1 ||
			source.Features[0].Properties.Name != "Nairobi" ||
			source.Features[0].Properties.Group != "KEN" ||
			source.Features[0].Properties.AdminType != "ADM1" {
			return errors.New("Nairobi boundary migration requires exactly one Kenya ADM1 Nairobi feature")
		}
		args = append(args, string(payload))
	}

	if _, err := tx.Exec(ctx, mig.Content, args...); err != nil {
		return fmt.Errorf("executing SQL: %w", err)
	}
	if mig.Name == "nairobi_county_boundary" {
		var valid, empty bool
		var geometryType string
		var srid, points int
		var minLng, minLat, maxLng, maxLat float64
		err := tx.QueryRow(ctx, `
			SELECT ST_IsValid(geom), ST_IsEmpty(geom), GeometryType(geom), ST_SRID(geom), ST_NPoints(geom),
			       ST_XMin(Box3D(geom)), ST_YMin(Box3D(geom)), ST_XMax(Box3D(geom)), ST_YMax(Box3D(geom))
			FROM boundaries WHERE id = 'c2000000-0000-0000-0000-000000000001'
		`).Scan(&valid, &empty, &geometryType, &srid, &points, &minLng, &minLat, &maxLng, &maxLat)
		if err != nil {
			return fmt.Errorf("validating promoted Nairobi boundary: %w", err)
		}
		if !valid || empty || geometryType != "MULTIPOLYGON" || srid != 4326 || points < 1000 ||
			minLng < 36.65 || minLng > 36.67 || minLat < -1.46 || minLat > -1.43 ||
			maxLng < 37.09 || maxLng > 37.12 || maxLat < -1.18 || maxLat > -1.14 {
			return fmt.Errorf("promoted Nairobi boundary failed geometry checks: valid=%t empty=%t type=%s srid=%d points=%d bounds=[%f,%f,%f,%f]", valid, empty, geometryType, srid, points, minLng, minLat, maxLng, maxLat)
		}
	}

	// Record in schema_migrations
	recordQuery := `
	INSERT INTO schema_migrations (version, name, applied_at, checksum)
	VALUES ($1, $2, NOW(), $3);`
	if _, err := tx.Exec(ctx, recordQuery, mig.Version, mig.Name, mig.Checksum); err != nil {
		return fmt.Errorf("recording migration history: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

// Down rolls back the last applied migration.
func (m *Migrator) Down(ctx context.Context, steps int) ([]int64, error) {
	if steps <= 0 {
		steps = 1
	}

	if err := m.EnsureSchemaMigrationsTable(ctx); err != nil {
		return nil, err
	}

	_, downMigrations, err := m.LoadMigrations()
	if err != nil {
		return nil, err
	}

	rows, err := m.pool.Query(ctx, "SELECT version, name FROM schema_migrations ORDER BY version DESC")
	if err != nil {
		return nil, fmt.Errorf("querying applied migrations for rollback: %w", err)
	}
	defer rows.Close()

	type appliedRec struct {
		version int64
		name    string
	}
	var appliedList []appliedRec
	for rows.Next() {
		var rec appliedRec
		if err := rows.Scan(&rec.version, &rec.name); err != nil {
			return nil, fmt.Errorf("scanning applied migration: %w", err)
		}
		appliedList = append(appliedList, rec)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("iterating applied migrations: %w", rows.Err())
	}

	if len(appliedList) == 0 {
		return nil, nil // Nothing to roll back
	}

	toRollback := steps
	if toRollback > len(appliedList) {
		toRollback = len(appliedList)
	}

	var rolledBack []int64
	for i := 0; i < toRollback; i++ {
		target := appliedList[i]
		downMig, ok := downMigrations[target.version]
		if !ok {
			return rolledBack, fmt.Errorf("missing rollback file for version %03d_%s", target.version, target.name)
		}

		if err := m.rollbackMigration(ctx, downMig); err != nil {
			return rolledBack, fmt.Errorf("rolling back migration %s: %w", downMig.Filename, err)
		}
		rolledBack = append(rolledBack, target.version)
	}

	return rolledBack, nil
}

func (m *Migrator) rollbackMigration(ctx context.Context, mig *Migration) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning rollback transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, mig.Content); err != nil {
		return fmt.Errorf("executing rollback SQL: %w", err)
	}

	deleteQuery := `DELETE FROM schema_migrations WHERE version = $1;`
	if _, err := tx.Exec(ctx, deleteQuery, mig.Version); err != nil {
		return fmt.Errorf("removing migration history: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing rollback transaction: %w", err)
	}
	return nil
}

// Status returns the status of all available migrations.
func (m *Migrator) Status(ctx context.Context) ([]MigrationStatus, error) {
	if err := m.EnsureSchemaMigrationsTable(ctx); err != nil {
		return nil, err
	}

	upMigrations, _, err := m.LoadMigrations()
	if err != nil {
		return nil, err
	}

	rows, err := m.pool.Query(ctx, "SELECT version, applied_at, checksum FROM schema_migrations")
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("querying migration statuses: %w", err)
	}
	defer rows.Close()

	type appInfo struct {
		appliedAt time.Time
		checksum  string
	}
	appliedMap := make(map[int64]appInfo)
	for rows.Next() {
		var v int64
		var info appInfo
		if err := rows.Scan(&v, &info.appliedAt, &info.checksum); err != nil {
			return nil, fmt.Errorf("scanning migration status: %w", err)
		}
		appliedMap[v] = info
	}

	// Collect all known versions from both FS and DB
	versionSet := make(map[int64]string)
	for v, mig := range upMigrations {
		versionSet[v] = mig.Name
	}
	for v := range appliedMap {
		if _, exists := versionSet[v]; !exists {
			versionSet[v] = "(unknown in files)"
		}
	}

	sortedVersions := make([]int64, 0, len(versionSet))
	for v := range versionSet {
		sortedVersions = append(sortedVersions, v)
	}
	sort.Slice(sortedVersions, func(i, j int) bool { return sortedVersions[i] < sortedVersions[j] })

	statuses := make([]MigrationStatus, 0, len(sortedVersions))
	for _, v := range sortedVersions {
		st := MigrationStatus{
			Version: v,
			Name:    versionSet[v],
		}
		if info, ok := appliedMap[v]; ok {
			st.Applied = true
			st.AppliedAt = &info.appliedAt
			st.Checksum = info.checksum
		}
		statuses = append(statuses, st)
	}

	return statuses, nil
}
