package datasets

import (
	"time"
)

// Dataset represents a registered data source in the catalogue.
type Dataset struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Publisher    string    `json:"publisher"`
	SourceURL    string    `json:"source_url"`
	Licence      string    `json:"licence"`
	LicenceURL   *string   `json:"licence_url,omitempty"`
	DataType     string    `json:"data_type"`
	Format       string    `json:"format"`
	SRID         int       `json:"srid"`
	GeoCoverage  string    `json:"geo_coverage"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// DatasetVersion represents a specific version of a dataset.
type DatasetVersion struct {
	ID            string     `json:"id"`
	DatasetID     string     `json:"dataset_id"`
	Version       string     `json:"version"`
	AcquisitionAt time.Time  `json:"acquisition_at"`
	Checksum      *string    `json:"checksum,omitempty"`
	FileSize      *int64     `json:"file_size,omitempty"`
	FeatureCount  *int       `json:"feature_count,omitempty"`
	GeoCoverage   *string    `json:"geo_coverage_pct,omitempty"`
	ValidationStatus string  `json:"validation_status"`
	PromotionStatus  string  `json:"promotion_status"`
	PromotedAt    *time.Time `json:"promoted_at,omitempty"`
	PromotedBy    *string    `json:"promoted_by,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// IngestionJob tracks a data ingestion operation.
type IngestionJob struct {
	ID              string     `json:"id"`
	DatasetID       string     `json:"dataset_id"`
	DatasetVersion  *string    `json:"dataset_version,omitempty"`
	Status          string     `json:"status"`
	JobType         string     `json:"job_type"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	RecordsTotal    *int       `json:"records_total,omitempty"`
	RecordsIngested *int       `json:"records_ingested,omitempty"`
	RecordsFailed   *int       `json:"records_failed,omitempty"`
	ErrorLog        *string    `json:"error_log,omitempty"`
	CreatedBy       string     `json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ValidationReport holds the result of validating a dataset version.
type ValidationReport struct {
	ID               string    `json:"id"`
	DatasetVersionID string    `json:"dataset_version_id"`
	ValidGeometries  int       `json:"valid_geometries"`
	InvalidGeometries int      `json:"invalid_geometries"`
	DuplicateCount   int       `json:"duplicate_count"`
	MissingValues    int       `json:"missing_values"`
	CRSValid         bool      `json:"crs_valid"`
	BoundsValid      bool      `json:"bounds_valid"`
	OverallStatus    string    `json:"overall_status"`
	Details          string    `json:"details"`
	CreatedAt        time.Time `json:"created_at"`
}

// Dataset statuses.
const (
	DatasetStatusDraft      = "draft"
	DatasetStatusRegistered = "registered"
	DatasetStatusActive     = "active"
	DatasetStatusRetired    = "retired"
)

// Validation statuses.
const (
	ValidationPending  = "pending"
	ValidationPassed   = "passed"
	ValidationFailed   = "failed"
	ValidationWarnings = "passed_with_warnings"
)

// Promotion statuses.
const (
	PromotionNone     = "none"
	PromotionPending  = "pending_review"
	PromotionApproved = "approved"
	PromotionRejected = "rejected"
)

// Ingestion job statuses.
const (
	JobStatusPending    = "pending"
	JobStatusRunning    = "running"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
	JobStatusCancelled  = "cancelled"
)
