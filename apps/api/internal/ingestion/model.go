package ingestion

import (
	"time"
)

// Nairobi County approximate geographic bounding box (EPSG:4326 WGS84).
const (
	NairobiMinLng = 36.65
	NairobiMaxLng = 37.10
	NairobiMinLat = -1.45
	NairobiMaxLat = -1.15
)

// Source represents an external geospatial data source configured for ingestion.
type Source struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Publisher   string    `json:"publisher"`
	SourceURL   string    `json:"source_url"`
	Format      string    `json:"format"` // "geojson", "csv", "shapefile"
	ExpectedCRS string    `json:"expected_crs"`
	Licence     string    `json:"licence"`
	LicenceURL  string    `json:"licence_url,omitempty"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// IngestionJob tracks an execution of dataset ingestion or processing.
type IngestionJob struct {
	ID              string     `json:"id"`
	DatasetID       string     `json:"dataset_id"`
	DatasetVersion  string     `json:"dataset_version"`
	Status          string     `json:"status"` // "pending", "running", "completed", "failed", "cancelled"
	JobType         string     `json:"job_type"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	RecordsTotal    int        `json:"records_total"`
	RecordsIngested int        `json:"records_ingested"`
	RecordsFailed   int        `json:"records_failed"`
	ErrorLog        string     `json:"error_log,omitempty"`
	CreatedBy       string     `json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ValidationReport contains the quality audit of an ingested dataset version.
type ValidationReport struct {
	ID                string    `json:"id"`
	DatasetVersionID  string    `json:"dataset_version_id"`
	ValidGeometries   int       `json:"valid_geometries"`
	InvalidGeometries int       `json:"invalid_geometries"`
	DuplicateCount    int       `json:"duplicate_count"`
	MissingValues     int       `json:"missing_values"`
	CRSValid          bool      `json:"crs_valid"`
	BoundsValid       bool      `json:"bounds_valid"`
	OverallStatus     string    `json:"overall_status"` // "passed", "passed_with_warnings", "failed"
	BoundaryMethod    string    `json:"boundary_method"` // "envelope_screening"
	SpatialLimitation string    `json:"spatial_limitation,omitempty"`
	Issues            []string  `json:"issues,omitempty"`
	Warnings          []string  `json:"warnings,omitempty"`
	ValidatedAt       time.Time `json:"validated_at"`
}

// AcquiredArtifact represents a fetched external dataset stored locally or in memory.
type AcquiredArtifact struct {
	SourceID    string    `json:"source_id"`
	VersionID   string    `json:"version_id"`
	SourceURL   string    `json:"source_url"`
	Checksum    string    `json:"checksum"` // SHA-256
	ByteSize    int64     `json:"byte_size"`
	AcquiredAt  time.Time `json:"acquired_at"`
	Payload     []byte    `json:"-"` // Raw payload stored safely in memory/spool
}

// PromotionRequest defines parameters to promote a validated dataset version to operational status.
type PromotionRequest struct {
	DatasetID  string `json:"dataset_id"`
	Version    string `json:"version"`
	PromotedBy string `json:"promoted_by"`
	Notes      string `json:"notes,omitempty"`
	Force      bool   `json:"force,omitempty"`
}

// PromotionResult records the outcome of a promotion action.
type PromotionResult struct {
	DatasetID        string    `json:"dataset_id"`
	Version          string    `json:"version"`
	Status           string    `json:"status"`
	FeaturesPromoted int       `json:"features_promoted"`
	PromotedAt       time.Time `json:"promoted_at"`
	PromotedBy       string    `json:"promoted_by"`
	Message          string    `json:"message"`
}
