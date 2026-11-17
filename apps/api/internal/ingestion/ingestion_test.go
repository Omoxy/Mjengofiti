package ingestion_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mjengofiti/api/internal/auth"
	"github.com/mjengofiti/api/internal/ingestion"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestCreateJobUsesAuthenticatedActor(t *testing.T) {
	service := ingestion.NewService(nil, newTestLogger())
	handler := ingestion.NewHandler(service, newTestLogger())
	request := httptest.NewRequest(http.MethodPost, "/api/v1/ingestion/jobs", strings.NewReader(`{"dataset_id":"ds-test","created_by":"spoofed-user"}`))
	request = request.WithContext(auth.WithUser(request.Context(), &auth.User{ID: "verified-subject", Role: auth.RoleDataSteward}))
	response := httptest.NewRecorder()

	handler.CreateJob(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", response.Code, response.Body.String())
	}
	var job ingestion.IngestionJob
	if err := json.NewDecoder(response.Body).Decode(&job); err != nil {
		t.Fatalf("decode job: %v", err)
	}
	if job.CreatedBy != "verified-subject" {
		t.Fatalf("expected authenticated subject as actor, got %q", job.CreatedBy)
	}
}

func TestRegisterSourceRejectsOversizedRequest(t *testing.T) {
	service := ingestion.NewService(nil, newTestLogger())
	handler := ingestion.NewHandler(service, newTestLogger())
	body := `{"name":"` + strings.Repeat("x", 1<<20) + `","source_url":"https://example.test/data.geojson"}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/ingestion/sources", strings.NewReader(body))
	response := httptest.NewRecorder()

	handler.RegisterSource(response, request)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 for oversized body, got %d", response.Code)
	}
}

func TestAcquisitionFailureDoesNotEchoInternalDetails(t *testing.T) {
	service := ingestion.NewService(nil, newTestLogger())
	handler := ingestion.NewHandler(service, newTestLogger())
	request := httptest.NewRequest(http.MethodPost, "/api/v1/ingestion/acquire", strings.NewReader(`{"source_id":"missing-source"}`))
	response := httptest.NewRecorder()

	handler.Acquire(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", response.Code)
	}
	if strings.Contains(response.Body.String(), "missing-source") {
		t.Fatalf("response exposed an internal source identifier: %s", response.Body.String())
	}
}

func TestValidateGeoJSON_ValidNairobiPolygon(t *testing.T) {
	validator := ingestion.NewValidator()

	// Polygon located within Nairobi Central Business District (~36.817, -1.286)
	validGeoJSON := []byte(`{
		"type": "FeatureCollection",
		"crs": {
			"type": "name",
			"properties": { "name": "urn:ogc:def:crs:OGC:1.3:CRS84" }
		},
		"features": [
			{
				"type": "Feature",
				"id": "bldg-101",
				"properties": { "source": "microsoft", "area_sq_m": 145.2 },
				"geometry": {
					"type": "Polygon",
					"coordinates": [
						[
							[36.8170, -1.2860],
							[36.8175, -1.2860],
							[36.8175, -1.2865],
							[36.8170, -1.2865],
							[36.8170, -1.2860]
						]
					]
				}
			}
		]
	}`)

	report, err := validator.ValidateGeoJSON("ver-test-01", validGeoJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.OverallStatus != "passed" {
		t.Errorf("expected status 'passed', got %q (issues: %v)", report.OverallStatus, report.Issues)
	}
	if report.ValidGeometries != 1 {
		t.Errorf("expected 1 valid geometry, got %d", report.ValidGeometries)
	}
	if !report.CRSValid {
		t.Error("expected CRSValid to be true")
	}
	if !report.BoundsValid {
		t.Error("expected BoundsValid to be true")
	}
}

func TestValidateGeoJSON_UnclosedPolygonRing(t *testing.T) {
	validator := ingestion.NewValidator()

	// Unclosed ring: end point does not match start point
	unclosedGeoJSON := []byte(`{
		"type": "FeatureCollection",
		"features": [
			{
				"type": "Feature",
				"geometry": {
					"type": "Polygon",
					"coordinates": [
						[
							[36.8170, -1.2860],
							[36.8175, -1.2860],
							[36.8175, -1.2865],
							[36.8170, -1.2865]
						]
					]
				}
			}
		]
	}`)

	report, err := validator.ValidateGeoJSON("ver-unclosed", unclosedGeoJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.OverallStatus != "failed" {
		t.Errorf("expected status 'failed', got %q", report.OverallStatus)
	}
	if report.InvalidGeometries != 1 {
		t.Errorf("expected 1 invalid geometry, got %d", report.InvalidGeometries)
	}
}

func TestValidateGeoJSON_OutOfBoundsCoordinates(t *testing.T) {
	validator := ingestion.NewValidator()

	// Coordinates in London, UK (~ -0.12, 51.5) outside Nairobi bounds
	outOfBoundsGeoJSON := []byte(`{
		"type": "FeatureCollection",
		"features": [
			{
				"type": "Feature",
				"geometry": {
					"type": "Polygon",
					"coordinates": [
						[
							[-0.120, 51.500],
							[-0.115, 51.500],
							[-0.115, 51.505],
							[-0.120, 51.505],
							[-0.120, 51.500]
						]
					]
				}
			}
		]
	}`)

	report, err := validator.ValidateGeoJSON("ver-oob", outOfBoundsGeoJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.OverallStatus != "failed" {
		t.Errorf("expected status 'failed' for out of bounds coordinates, got %q", report.OverallStatus)
	}
	if report.BoundsValid {
		t.Error("expected BoundsValid to be false")
	}
}

func TestIngestionService_JobAndPromotionLifecycle(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	svc := ingestion.NewService(nil, logger)

	// Verify authoritative sources seeded
	sources := svc.ListSources()
	if len(sources) < 3 {
		t.Fatalf("expected at least 3 default sources, got %d", len(sources))
	}

	// Create Ingestion Job
	job, err := svc.CreateJob(ctx, "src-ms-kenya", "2025.1", "test-user")
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}
	if job.Status != "pending" {
		t.Errorf("expected job status 'pending', got %q", job.Status)
	}

	fetchedJob, err := svc.GetJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("failed to fetch job: %v", err)
	}
	if fetchedJob.DatasetID != "src-ms-kenya" {
		t.Errorf("expected dataset ID 'src-ms-kenya', got %q", fetchedJob.DatasetID)
	}

	// Validate a failing dataset
	badGeoJSON := []byte(`{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"Point","coordinates":[0.0, 0.0]}}]}`)
	_, err = svc.ValidateDataset(ctx, "ver-failing", badGeoJSON)
	if err != nil {
		t.Fatalf("validation call returned error: %v", err)
	}

	// Attempting to promote failed version without force should be rejected
	_, err = svc.PromoteVersion(ctx, ingestion.PromotionRequest{
		DatasetID:  "src-ms-kenya",
		Version:    "ver-failing",
		PromotedBy: "Lead Data Steward",
		Force:      false,
	})
	if err == nil {
		t.Fatal("expected promotion of failed dataset to be rejected, but it succeeded")
	}

	// Attempting to promote failed version with force but NO notes should be rejected
	_, err = svc.PromoteVersion(ctx, ingestion.PromotionRequest{
		DatasetID:  "src-ms-kenya",
		Version:    "ver-failing",
		PromotedBy: "Lead Data Steward",
		Force:      true,
		Notes:      "",
	})
	if err == nil {
		t.Fatal("expected forced promotion without notes to be rejected")
	}

	// Promotion of unreviewed dataset should be rejected
	_, err = svc.PromoteVersion(ctx, ingestion.PromotionRequest{
		DatasetID:  "src-ms-kenya",
		Version:    "ver-never-validated",
		PromotedBy: "Lead Data Steward",
	})
	if err == nil {
		t.Fatal("expected unreviewed dataset promotion to be rejected")
	}

	// Validate a passing dataset
	goodGeoJSON := []byte(`{
		"type": "FeatureCollection",
		"features": [
			{
				"type": "Feature",
				"properties": { "source": "ms", "area_sq_m": 120.5 },
				"geometry": {
					"type": "Polygon",
					"coordinates": [[[36.81, -1.28], [36.82, -1.28], [36.82, -1.29], [36.81, -1.29], [36.81, -1.28]]]
				}
			}
		]
	}`)
	_, err = svc.ValidateDataset(ctx, "ver-passing", goodGeoJSON)
	if err != nil {
		t.Fatalf("validation call returned error: %v", err)
	}

	// Promotion of passing dataset should succeed and record features count
	promoResult, err := svc.PromoteVersion(ctx, ingestion.PromotionRequest{
		DatasetID:  "src-ms-kenya",
		Version:    "ver-passing",
		PromotedBy: "Lead Data Steward",
	})
	if err != nil {
		t.Fatalf("expected promotion to succeed, got error: %v", err)
	}
	if promoResult.Status != "promoted" {
		t.Errorf("expected promotion status 'promoted', got %q", promoResult.Status)
	}
	if promoResult.FeaturesPromoted != 1 {
		t.Errorf("expected 1 feature promoted, got %d", promoResult.FeaturesPromoted)
	}
}

func TestValidateGeoJSON_RFC7946DefaultCRS(t *testing.T) {
	validator := ingestion.NewValidator()

	// RFC 7946 document without any 'crs' member (defaults to WGS84)
	noCRSGeoJSON := []byte(`{
		"type": "FeatureCollection",
		"features": [
			{
				"type": "Feature",
				"properties": { "source": "test" },
				"geometry": {
					"type": "Polygon",
					"coordinates": [
						[
							[36.8170, -1.2860],
							[36.8175, -1.2860],
							[36.8175, -1.2865],
							[36.8170, -1.2865],
							[36.8170, -1.2860]
						]
					]
				}
			}
		]
	}`)

	report, err := validator.ValidateGeoJSON("ver-no-crs", noCRSGeoJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !report.CRSValid {
		t.Error("expected CRS to be valid by default under RFC 7946")
	}
	if report.OverallStatus != "passed" {
		t.Errorf("expected status 'passed', got %q", report.OverallStatus)
	}
	if report.BoundaryMethod != "envelope_screening" {
		t.Errorf("expected boundary method 'envelope_screening', got %q", report.BoundaryMethod)
	}
	if report.SpatialLimitation == "" {
		t.Error("expected spatial limitation disclosure to be present")
	}
}

func TestValidateGeoJSON_EmptyFeatureCollection(t *testing.T) {
	validator := ingestion.NewValidator()

	emptyGeoJSON := []byte(`{
		"type": "FeatureCollection",
		"features": []
	}`)

	report, err := validator.ValidateGeoJSON("ver-empty", emptyGeoJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.OverallStatus != "passed_with_warnings" {
		t.Errorf("expected status 'passed_with_warnings' for empty collection, got %q", report.OverallStatus)
	}
	if len(report.Warnings) == 0 {
		t.Error("expected warning about empty feature collection")
	}
}

func TestValidateGeoJSON_InvalidAndMalformedCoordinates(t *testing.T) {
	validator := ingestion.NewValidator()

	// 1D coordinate array (missing latitude)
	malformed1D := []byte(`{
		"type": "FeatureCollection",
		"features": [
			{
				"type": "Feature",
				"geometry": {
					"type": "Point",
					"coordinates": [36.8]
				}
			}
		]
	}`)

	report, err := validator.ValidateGeoJSON("ver-1d", malformed1D)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.OverallStatus != "failed" {
		t.Errorf("expected failed status for 1D coordinate, got %q", report.OverallStatus)
	}
	if report.InvalidGeometries != 1 {
		t.Errorf("expected 1 invalid geometry, got %d", report.InvalidGeometries)
	}

	// Out of global geographic coordinate range (lat = 195.0)
	outOfRangeGlobal := []byte(`{
		"type": "FeatureCollection",
		"features": [
			{
				"type": "Feature",
				"geometry": {
					"type": "Point",
					"coordinates": [36.8, 195.0]
				}
			}
		]
	}`)

	reportGlobal, err := validator.ValidateGeoJSON("ver-out-of-range", outOfRangeGlobal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reportGlobal.OverallStatus != "failed" {
		t.Errorf("expected failed status for out-of-range global coordinate, got %q", reportGlobal.OverallStatus)
	}
}

func TestAcquireSource_SSRFProtection(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	svc := ingestion.NewService(nil, logger)

	testCases := []struct {
		id        string
		name      string
		targetURL string
	}{
		{"src-malicious-local", "Local Host Loopback Target", "http://127.0.0.1:8080/data.json"},
		{"src-malicious-meta", "Cloud Metadata Target", "http://169.254.169.254/latest/meta-data"},
		{"src-malicious-priv", "RFC 1918 Private Target", "http://10.0.0.1:8080/data.json"},
	}

	for _, tc := range testCases {
		source := ingestion.Source{
			ID:          tc.id,
			Name:        tc.name,
			SourceURL:   tc.targetURL,
			Format:      "geojson",
			ExpectedCRS: "EPSG:4326",
		}

		_, err := svc.RegisterSource(source)
		if err != nil {
			t.Fatalf("failed to register source %s: %v", tc.id, err)
		}

		_, err = svc.AcquireSource(ctx, tc.id, "ver-ssrf-test")
		if err == nil {
			t.Fatalf("expected AcquireSource to reject %s with SSRF error, but it succeeded", tc.targetURL)
		}
		if !strings.Contains(err.Error(), "SSRF prevention") && !strings.Contains(err.Error(), "insecure source URL") {
			t.Errorf("expected SSRF error substring for %s, got: %v", tc.targetURL, err)
		}
	}
}
