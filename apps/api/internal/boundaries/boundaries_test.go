package boundaries_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mjengofiti/api/internal/boundaries"
)

func TestBoundaryGeoJSON(t *testing.T) {
	h := boundaries.NewHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/boundaries", nil)
	w := httptest.NewRecorder()

	h.HandleNairobiBoundary(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/geo+json" {
		t.Errorf("expected application/geo+json, got %s", contentType)
	}

	var fc struct {
		Type     string `json:"type"`
		Features []struct {
			Properties struct {
				Name                string `json:"shapeName"`
				ISO                 string `json:"shapeISO"`
				Group               string `json:"shapeGroup"`
				AdminType           string `json:"shapeType"`
				DatasetName         string `json:"boundary_dataset_name"`
				BoundaryIdentifier  string `json:"boundary_identifier"`
				SourceName          string `json:"boundary_source_name"`
				SourceOrganisation  string `json:"boundary_source_organisation"`
				SourceURL           string `json:"boundary_source_url"`
				DownloadURL         string `json:"boundary_download_url"`
				AdminLevel          string `json:"boundary_admin_level"`
				RetrievedAt         string `json:"boundary_retrieved_at"`
				License             string `json:"boundary_license"`
				CoordinateReference string `json:"coordinate_reference_system"`
				Validation          struct {
					Valid      bool   `json:"is_valid"`
					Empty      bool   `json:"is_empty"`
					Type       string `json:"geometry_type"`
					SRID       int    `json:"srid"`
					VertexCount int    `json:"vertex_count"`
				} `json:"postgis_validation"`
			} `json:"properties"`
			Geometry struct {
				Type        string      `json:"type"`
				Coordinates [][][]float64 `json:"coordinates"`
			} `json:"geometry"`
		} `json:"features"`
	}
	if err := json.NewDecoder(w.Body).Decode(&fc); err != nil {
		t.Fatalf("failed decoding GeoJSON: %v", err)
	}

	if fc.Type != "FeatureCollection" {
		t.Fatalf("expected FeatureCollection, got %q", fc.Type)
	}
	if len(fc.Features) != 1 {
		t.Fatalf("expected one county feature, got %d", len(fc.Features))
	}
	feature := fc.Features[0]
	if feature.Properties.Name != "Nairobi" || feature.Properties.ISO != "KE-30" || feature.Properties.Group != "KEN" || feature.Properties.AdminType != "ADM1" {
		t.Fatalf("unexpected source feature identity: %+v", feature.Properties)
	}
	if feature.Properties.DatasetName != "geoBoundaries Kenya ADM1 2020" ||
		feature.Properties.BoundaryIdentifier != "KEN-ADM1-32016919" ||
		feature.Properties.SourceName != "geoBoundaries" ||
		feature.Properties.SourceOrganisation != "RCMRD GeoPortal, Africa GeoPortal" ||
		feature.Properties.AdminLevel != "ADM1 (county)" ||
		feature.Properties.RetrievedAt != "2026-11-09" ||
		feature.Properties.License != "Public Domain" ||
		feature.Properties.CoordinateReference == "" {
		t.Fatalf("incomplete boundary provenance metadata: %+v", feature.Properties)
	}
	if feature.Properties.SourceURL == "" || feature.Properties.DownloadURL == "" {
		t.Fatal("boundary provenance is missing its source URL")
	}
	if feature.Geometry.Type != "Polygon" || len(feature.Geometry.Coordinates) != 1 {
		t.Fatalf("expected one Polygon geometry, got %s", feature.Geometry.Type)
	}
	ring := feature.Geometry.Coordinates[0]
	if len(ring) != 1490 {
		t.Fatalf("expected full 1490-vertex source ring, got %d vertices", len(ring))
	}
	if ring[0][0] != ring[len(ring)-1][0] || ring[0][1] != ring[len(ring)-1][1] {
		t.Fatal("source polygon ring is not closed")
	}
	if !feature.Properties.Validation.Valid || feature.Properties.Validation.Empty ||
		feature.Properties.Validation.Type != "MULTIPOLYGON" ||
		feature.Properties.Validation.SRID != 4326 ||
		feature.Properties.Validation.VertexCount != 1490 {
		t.Fatalf("unexpected PostGIS validation metadata: %+v", feature.Properties.Validation)
	}
}
