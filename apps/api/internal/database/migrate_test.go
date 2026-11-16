package database_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/mjengofiti/api/internal/database"
	"github.com/mjengofiti/api/migrations"
)

func TestLoadMigrations_FromEmbeddedFS(t *testing.T) {
	migrator := database.NewMigrator(nil, migrations.FS)

	upMigs, downMigs, err := migrator.LoadMigrations()
	if err != nil {
		t.Fatalf("LoadMigrations() returned error: %v", err)
	}

	if len(upMigs) == 0 {
		t.Fatal("expected at least one up migration in embedded migrations.FS, found 0")
	}

	if len(downMigs) == 0 {
		t.Fatal("expected at least one down migration in embedded migrations.FS, found 0")
	}

	// Verify 001 exists
	mig001Up, ok := upMigs[1]
	if !ok {
		t.Fatalf("expected version 1 in up migrations, got keys: %v", upMigs)
	}
	if mig001Up.Name != "initial_schema" {
		t.Errorf("expected name 'initial_schema', got %q", mig001Up.Name)
	}
	if mig001Up.Checksum == "" {
		t.Error("expected non-empty SHA-256 checksum for 001 up migration")
	}
	if !strings.Contains(mig001Up.Content, "CREATE TABLE buildings") {
		t.Error("001_initial_schema.up.sql does not contain CREATE TABLE buildings")
	}

	mig001Down, ok := downMigs[1]
	if !ok {
		t.Fatalf("expected version 1 in down migrations, got keys: %v", downMigs)
	}
	if mig001Down.Name != "initial_schema" {
		t.Errorf("expected name 'initial_schema', got %q", mig001Down.Name)
	}
	if !strings.Contains(mig001Down.Content, "DROP TABLE IF EXISTS buildings") {
		t.Error("001_initial_schema.down.sql does not contain DROP TABLE IF EXISTS buildings")
	}
}

func TestLoadMigrations_CustomMapFS(t *testing.T) {
	mockFS := fstest.MapFS{
		"001_init.up.sql":      &fstest.MapFile{Data: []byte("CREATE TABLE t1 (id int);")},
		"001_init.down.sql":    &fstest.MapFile{Data: []byte("DROP TABLE t1;")},
		"002_add_col.up.sql":   &fstest.MapFile{Data: []byte("ALTER TABLE t1 ADD COLUMN name text;")},
		"002_add_col.down.sql": &fstest.MapFile{Data: []byte("ALTER TABLE t1 DROP COLUMN name;")},
		"readme.md":            &fstest.MapFile{Data: []byte("# Ignored file")},
		"invalid_name.sql":     &fstest.MapFile{Data: []byte("SELECT 1;")},
	}

	migrator := database.NewMigrator(nil, mockFS)
	upMigs, downMigs, err := migrator.LoadMigrations()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(upMigs) != 2 {
		t.Errorf("expected 2 up migrations, got %d", len(upMigs))
	}
	if len(downMigs) != 2 {
		t.Errorf("expected 2 down migrations, got %d", len(downMigs))
	}

	if upMigs[1].Name != "init" || upMigs[2].Name != "add_col" {
		t.Errorf("unexpected names: %s, %s", upMigs[1].Name, upMigs[2].Name)
	}
}

func TestMigrationSchemaConsistency(t *testing.T) {
	migrator := database.NewMigrator(nil, migrations.FS)
	upMigs, downMigs, err := migrator.LoadMigrations()
	if err != nil {
		t.Fatalf("LoadMigrations() error: %v", err)
	}

	// Verify all up migrations have matching down migrations
	for v, upMig := range upMigs {
		downMig, hasDown := downMigs[v]
		if !hasDown {
			t.Errorf("migration version %d (%s) has no matching .down.sql rollback file", v, upMig.Name)
		} else if downMig.Name != upMig.Name {
			t.Errorf("migration version %d name mismatch: up=%q, down=%q", v, upMig.Name, downMig.Name)
		}
	}

	// Verify essential Mjengofiti domain tables in initial schema
	initialSchema := upMigs[1].Content
	requiredTables := []string{
		"roles",
		"users",
		"audit_events",
		"datasets",
		"dataset_versions",
		"ingestion_jobs",
		"validation_reports",
		"buildings",
		"building_observations",
		"building_indicators",
		"indicator_definitions",
		"inspection_cases",
		"case_history",
		"case_findings",
		"standards",
		"boundaries",
	}

	for _, table := range requiredTables {
		createPattern := "CREATE TABLE " + table
		if !strings.Contains(initialSchema, createPattern) {
			t.Errorf("001_initial_schema.up.sql is missing table: %s", table)
		}
	}
}
