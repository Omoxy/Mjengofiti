package buildings

import (
	"time"
)

// Building represents a building record with geometry and provenance.
type Building struct {
	ID              string     `json:"id"`
	ExternalID      *string    `json:"external_id,omitempty"`
	SourceDatasetID *string    `json:"source_dataset_id,omitempty"`
	SourceName      string     `json:"source_name"`
	Longitude       float64    `json:"longitude"`
	Latitude        float64    `json:"latitude"`
	AreaSqM         *float64   `json:"area_sq_m,omitempty"`
	HeightM         *float64   `json:"height_m,omitempty"`
	HeightMethod    *string    `json:"height_method,omitempty"`
	MatchMethod     *string    `json:"match_method,omitempty"`
	MatchConfidence *float64   `json:"match_confidence,omitempty"`
	DataQuality     string     `json:"data_quality"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// BuildingDetail extends Building with additional sourced observations and indicators.
type BuildingDetail struct {
	Building
	Observations []Observation `json:"observations"`
	Indicators   []Indicator   `json:"indicators"`
	GeoJSON      string        `json:"geojson,omitempty"`
}

// Observation represents a single sourced fact about a building.
type Observation struct {
	ID              string    `json:"id"`
	BuildingID      string    `json:"building_id"`
	Type            string    `json:"type"`
	Value           string    `json:"value"`
	Unit            *string   `json:"unit,omitempty"`
	ObservationDate *string   `json:"observation_date,omitempty"`
	SourceVersion   string    `json:"source_version"`
	Method          *string   `json:"method,omitempty"`
	Confidence      *float64  `json:"confidence,omitempty"`
	Limitations     *string   `json:"limitations,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// Indicator represents a derived analytical indicator for a building.
type Indicator struct {
	ID                string    `json:"id"`
	BuildingID        string    `json:"building_id"`
	IndicatorType     string    `json:"indicator_type"`
	Version           string    `json:"version"`
	Value             float64   `json:"value"`
	Unit              string    `json:"unit"`
	CalculationMethod string    `json:"calculation_method"`
	InputDatasets     []string  `json:"input_datasets"`
	MissingInputs     []string  `json:"missing_inputs,omitempty"`
	Uncertainty       *string   `json:"uncertainty,omitempty"`
	Limitations       *string   `json:"limitations,omitempty"`
	CalculatedAt      time.Time `json:"calculated_at"`
}

// BBoxQuery defines a spatial bounding box query.
type BBoxQuery struct {
	MinLng float64
	MinLat float64
	MaxLng float64
	MaxLat float64
	Limit  int
	Offset int
}

// SearchQuery defines a text and attribute search.
type SearchQuery struct {
	Text       string
	SourceName *string
	MinArea    *float64
	MaxArea    *float64
	Limit      int
	Offset     int
}

// ListResult wraps a paginated list of buildings.
type ListResult struct {
	Buildings  []Building `json:"buildings"`
	Total      int        `json:"total"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}

// MaxBBoxLimit is the maximum number of buildings returned in a single spatial query.
const MaxBBoxLimit = 5000

// MaxSearchLimit is the maximum number of buildings returned in a search.
const MaxSearchLimit = 200

// DefaultLimit is the default page size.
const DefaultLimit = 50
