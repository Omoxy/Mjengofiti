package ingestion

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)



// Service orchestrates dataset registration, ingestion, quality validation, and promotion.
type Service struct {
	db          *pgxpool.Pool
	logger      *slog.Logger
	validator   Validator
	mu          sync.RWMutex
	sources     map[string]Source
	jobs        map[string]*IngestionJob
	reports     map[string]*ValidationReport
	promotions  map[string]*PromotionResult
	artifacts   map[string]*AcquiredArtifact
	rawPayloads map[string][]byte
}

// NewService creates a new Ingestion Service instance.
func NewService(db *pgxpool.Pool, logger *slog.Logger) *Service {
	s := &Service{
		db:          db,
		logger:      logger,
		validator:   NewValidator(),
		sources:     make(map[string]Source),
		jobs:        make(map[string]*IngestionJob),
		reports:     make(map[string]*ValidationReport),
		promotions:  make(map[string]*PromotionResult),
		artifacts:   make(map[string]*AcquiredArtifact),
		rawPayloads: make(map[string][]byte),
	}

	s.seedAuthoritativeSources()
	return s
}


func (s *Service) seedAuthoritativeSources() {
	sources := []Source{
		{
			ID:          "src-ms-kenya",
			Name:        "Microsoft Kenya Building Footprints",
			Publisher:   "Microsoft Maps",
			SourceURL:   "https://github.com/microsoft/KenyaNigeriaBuildingFootprints",
			Format:      "geojson",
			ExpectedCRS: "EPSG:4326",
			Licence:     "ODbL 1.0",
			LicenceURL:  "https://opendatacommons.org/licenses/odbl/",
			Description: "Comprehensive building footprints extracted from high-resolution satellite imagery across Nairobi County.",
			CreatedAt:   time.Now().UTC(),
		},
		{
			ID:          "src-google-openbldg",
			Name:        "Google Open Buildings (Temporal V3)",
			Publisher:   "Google Research",
			SourceURL:   "https://sites.research.google/gr/open-buildings/temporal",
			Format:      "csv",
			ExpectedCRS: "EPSG:4326",
			Licence:     "CC-BY 4.0",
			LicenceURL:  "https://creativecommons.org/licenses/by/4.0/",
			Description: "Building footprints and roof-height estimations generated with machine learning for urban density profiling.",
			CreatedAt:   time.Now().UTC(),
		},
		{
			ID:          "src-osm-kenya",
			Name:        "OpenStreetMap Kenya Building & Infrastructure Layers",
			Publisher:   "OpenStreetMap Contributors",
			SourceURL:   "https://download.geofabrik.de/africa/kenya.html",
			Format:      "osm_pbf",
			ExpectedCRS: "EPSG:4326",
			Licence:     "ODbL",
			LicenceURL:  "https://www.openstreetmap.org/copyright",
			Description: "Crowdsourced infrastructure, road access networks, and facility attribution for Nairobi County.",
			CreatedAt:   time.Now().UTC(),
		},
		{
			ID:          "src-copernicus-dem",
			Name:        "Copernicus Digital Elevation Model (COP-DEM 30m)",
			Publisher:   "European Space Agency (ESA) / Copernicus",
			SourceURL:   "https://dataspace.copernicus.eu/",
			Format:      "cog",
			ExpectedCRS: "EPSG:4326",
			Licence:     "Copernicus Open Access Licence",
			LicenceURL:  "https://dataspace.copernicus.eu/terms-and-conditions",
			Description: "30-meter global elevation raster for hydrological slope and floodplain risk analysis.",
			CreatedAt:   time.Now().UTC(),
		},
		{
			ID:          "src-chirps-precip",
			Name:        "CHIRPS Historical & Daily Precipitation Monitoring",
			Publisher:   "Climate Hazards Center / UCSB",
			SourceURL:   "https://data.chc.ucsb.edu/products/CHIRPS-2.0/",
			Format:      "geotiff",
			ExpectedCRS: "EPSG:4326",
			Licence:     "Public Domain / US Federal Open Data",
			LicenceURL:  "https://www.chc.ucsb.edu/data/chirps",
			Description: "High-resolution rainfall estimates for Nairobi flood hazard modeling and drainage stress indexes.",
			CreatedAt:   time.Now().UTC(),
		},
		{
			ID:          "src-nca-register",
			Name:        "National Construction Authority Contractor & Site Registers",
			Publisher:   "National Construction Authority (NCA Kenya)",
			SourceURL:   "https://nca.go.ke/registers/",
			Format:      "csv",
			ExpectedCRS: "EPSG:4326",
			Licence:     "Kenya Open Data Initiative",
			LicenceURL:  "https://nca.go.ke/terms",
			Description: "Official regulatory registration records of accredited contractors, site supervisors, and compliance notices.",
			CreatedAt:   time.Now().UTC(),
		},
	}

	for _, src := range sources {
		s.sources[src.ID] = src
	}

}

// ListSources returns all registered ingestion sources.
func (s *Service) ListSources() []Source {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Source, 0, len(s.sources))
	for _, src := range s.sources {
		result = append(result, src)
	}
	return result
}

// GetSource retrieves a registered source by ID.
func (s *Service) GetSource(id string) (*Source, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	src, ok := s.sources[id]
	if !ok {
		return nil, fmt.Errorf("source %q not found", id)
	}
	return &src, nil
}

// RegisterSource registers a new external data source.
func (s *Service) RegisterSource(src Source) (*Source, error) {
	if src.Name == "" || src.SourceURL == "" {
		return nil, fmt.Errorf("name and source_url are required")
	}
	if src.ID == "" {
		src.ID = "src-" + uuid.New().String()[:8]
	}
	if src.ExpectedCRS == "" {
		src.ExpectedCRS = "EPSG:4326"
	}
	src.CreatedAt = time.Now().UTC()

	s.mu.Lock()
	s.sources[src.ID] = src
	s.mu.Unlock()

	s.logger.Info("registered new ingestion source", slog.String("id", src.ID), slog.String("name", src.Name))
	return &src, nil
}

// CreateJob enqueues a dataset ingestion or processing job.
func (s *Service) CreateJob(ctx context.Context, datasetID, version, createdBy string) (*IngestionJob, error) {
	if datasetID == "" {
		return nil, fmt.Errorf("dataset_id is required")
	}
	if createdBy == "" {
		createdBy = "system"
	}
	if version == "" {
		version = "1.0.0"
	}

	job := &IngestionJob{
		ID:             "job-" + uuid.New().String()[:8],
		DatasetID:      datasetID,
		DatasetVersion: version,
		Status:         "pending",
		JobType:        "full_ingestion",
		CreatedBy:      createdBy,
		CreatedAt:      time.Now().UTC(),
	}

	s.mu.Lock()
	s.jobs[job.ID] = job
	s.mu.Unlock()

	// Persist to database if pool is reachable
	if s.db != nil {
		go func(j IngestionJob) {
			bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, _ = s.db.Exec(bgCtx, `
				INSERT INTO ingestion_jobs (id, dataset_id, dataset_version, status, job_type, created_by, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				ON CONFLICT (id) DO NOTHING
			`, j.ID, j.DatasetID, j.DatasetVersion, j.Status, j.JobType, j.CreatedBy, j.CreatedAt)
		}(*job)
	}

	s.logger.Info("created ingestion job", slog.String("job_id", job.ID), slog.String("dataset", datasetID))
	return job, nil
}

// ListJobs returns all ingestion jobs.
func (s *Service) ListJobs(ctx context.Context) ([]IngestionJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]IngestionJob, 0, len(s.jobs))
	for _, job := range s.jobs {
		result = append(result, *job)
	}
	return result, nil
}

// GetJob retrieves an ingestion job by ID.
func (s *Service) GetJob(ctx context.Context, id string) (*IngestionJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.jobs[id]
	if !ok {
		return nil, fmt.Errorf("job %q not found", id)
	}
	return job, nil
}

// AcquireSource fetches the raw dataset payload from a registered source enforcing safe URL checks, SSRF prevention, and download limits.
func (s *Service) AcquireSource(ctx context.Context, sourceID, versionID string) (*AcquiredArtifact, error) {
	s.mu.RLock()
	src, ok := s.sources[sourceID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("source %q not found", sourceID)
	}
	if versionID == "" {
		versionID = "v-" + time.Now().UTC().Format("20060102150405")
	}

	if _, err := ValidateURLSyntax(src.SourceURL); err != nil {
		return nil, fmt.Errorf("insecure source URL %q: %w", src.SourceURL, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src.SourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	client := NewSafeHTTPClient(DefaultRequestTimeout)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing acquisition request for %s: %w", src.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("acquisition HTTP status %d (%s)", resp.StatusCode, resp.Status)
	}

	// Enforce streaming download limit to prevent memory exhaustion
	payload, err := ReadLimitedPayload(resp.Body, DefaultMaxDownloadBytes)
	if err != nil {
		return nil, fmt.Errorf("reading response payload: %w", err)
	}

	hash := sha256.Sum256(payload)
	checksum := hex.EncodeToString(hash[:])


	artifact := &AcquiredArtifact{
		SourceID:   sourceID,
		VersionID:  versionID,
		SourceURL:  src.SourceURL,
		Checksum:   checksum,
		ByteSize:   int64(len(payload)),
		AcquiredAt: time.Now().UTC(),
		Payload:    payload,
	}

	s.mu.Lock()
	s.artifacts[versionID] = artifact
	s.rawPayloads[versionID] = payload
	s.mu.Unlock()

	s.logger.Info("acquired source artifact",
		slog.String("source_id", sourceID),
		slog.String("version_id", versionID),
		slog.Int64("bytes", artifact.ByteSize),
		slog.String("checksum", checksum[:12]),
	)

	return artifact, nil
}

// GetArtifact retrieves an acquired artifact by version ID.
func (s *Service) GetArtifact(ctx context.Context, versionID string) (*AcquiredArtifact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	artifact, ok := s.artifacts[versionID]
	if !ok {
		return nil, fmt.Errorf("artifact for version %q not found", versionID)
	}
	return artifact, nil
}

// ValidateDataset validates a raw GeoJSON byte slice and stores the report.
func (s *Service) ValidateDataset(ctx context.Context, versionID string, rawData []byte) (*ValidationReport, error) {
	report, err := s.validator.ValidateGeoJSON(versionID, rawData)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	s.mu.Lock()
	s.reports[report.DatasetVersionID] = report
	s.rawPayloads[report.DatasetVersionID] = rawData
	s.mu.Unlock()

	// Persist report to database if pool is reachable
	if s.db != nil {
		go func(r ValidationReport) {
			bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, _ = s.db.Exec(bgCtx, `
				INSERT INTO validation_reports (
					id, dataset_version_id, valid_geometries, invalid_geometries,
					duplicate_count, missing_values, crs_valid, bounds_valid,
					overall_status, details, created_at
				) VALUES (
					uuid_generate_v4(), (SELECT id FROM dataset_versions WHERE version = $1 LIMIT 1),
					$2, $3, $4, $5, $6, $7, $8, $9, $10
				)
			`, r.DatasetVersionID, r.ValidGeometries, r.InvalidGeometries,
				r.DuplicateCount, r.MissingValues, r.CRSValid, r.BoundsValid,
				r.OverallStatus, r.Issues, r.ValidatedAt)
		}(*report)
	}

	s.logger.Info("validated dataset version",
		slog.String("version_id", versionID),
		slog.String("status", report.OverallStatus),
		slog.Int("valid_geometries", report.ValidGeometries),
		slog.Int("invalid_geometries", report.InvalidGeometries),
	)

	return report, nil
}

// GetValidationReport retrieves a stored validation report for a version.
func (s *Service) GetValidationReport(ctx context.Context, versionID string) (*ValidationReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	report, ok := s.reports[versionID]
	if !ok {
		return nil, fmt.Errorf("validation report for %q not found", versionID)
	}
	return report, nil
}

// PromoteVersion controls dataset promotion to operational status and loads spatial features.
func (s *Service) PromoteVersion(ctx context.Context, req PromotionRequest) (*PromotionResult, error) {
	if req.DatasetID == "" || req.Version == "" {
		return nil, fmt.Errorf("dataset_id and version are required")
	}
	if req.PromotedBy == "" {
		req.PromotedBy = "Authorized Data Steward"
	}

	s.mu.RLock()
	report, hasReport := s.reports[req.Version]
	rawPayload, hasPayload := s.rawPayloads[req.Version]
	s.mu.RUnlock()

	// Promotion Gate 1: unreviewed datasets without validation reports cannot enter operational use
	if !hasReport {
		return nil, fmt.Errorf("dataset version %q has not been validated; unreviewed datasets cannot be promoted to operational status", req.Version)
	}

	// Promotion Gate 2: failed datasets require explicit steward override and documented reason
	if report.OverallStatus == "failed" {
		if !req.Force {
			return nil, fmt.Errorf("dataset version %q failed quality validation and cannot be promoted without override", req.Version)
		}
		if req.Notes == "" {
			return nil, fmt.Errorf("forced promotion of failed dataset requires documented steward reason in notes")
		}
	}

	featuresCount := report.ValidGeometries

	// If database is available and we have payload, persist operational features to buildings table
	if s.db != nil && hasPayload {
		var fc GeoJSONFeatureCollection
		if err := json.Unmarshal(rawPayload, &fc); err == nil {
			for _, feat := range fc.Features {
				if feat.Geometry == nil {
					continue
				}
				geomBytes, err := json.Marshal(feat.Geometry)
				if err != nil {
					continue
				}

				extID := ""
				if feat.ID != nil {
					extID = fmt.Sprintf("%v", feat.ID)
				}
				var areaSqM float64
				if feat.Properties != nil {
					if a, ok := feat.Properties["area_sq_m"].(float64); ok {
						areaSqM = a
					}
				}

				_, _ = s.db.Exec(ctx, `
					INSERT INTO buildings (
						id, external_id, source_name, geom, area_sq_m,
						data_quality, created_at, updated_at
					) VALUES (
						uuid_generate_v4(), $1, $2, ST_SetSRID(ST_GeomFromGeoJSON($3), 4326),
						$4, 'ingested_verified', NOW(), NOW()
					)
					ON CONFLICT DO NOTHING
				`, extID, req.DatasetID, string(geomBytes), areaSqM)
			}
		}
	}

	res := &PromotionResult{
		DatasetID:        req.DatasetID,
		Version:          req.Version,
		Status:           "promoted",
		FeaturesPromoted: featuresCount,
		PromotedAt:       time.Now().UTC(),
		PromotedBy:       req.PromotedBy,
		Message:          fmt.Sprintf("Dataset version %s successfully promoted to operational status with %d features", req.Version, featuresCount),
	}

	s.mu.Lock()
	s.promotions[req.DatasetID+":"+req.Version] = res
	s.mu.Unlock()

	s.logger.Info("promoted dataset version",
		slog.String("dataset", req.DatasetID),
		slog.String("version", req.Version),
		slog.Int("features", featuresCount),
		slog.String("promoted_by", req.PromotedBy),
	)

	return res, nil
}
