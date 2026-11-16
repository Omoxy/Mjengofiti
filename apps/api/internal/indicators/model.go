package indicators

import "time"

// IndicatorDefinition describes an explainable screening indicator.
type IndicatorDefinition struct {
	ID                   string         `json:"id"`
	Name                 string         `json:"name"`
	DisplayName          string         `json:"display_name"`
	Description          string         `json:"description"`
	Purpose              string         `json:"purpose"`
	CalculationMethod    string         `json:"calculation_method"`
	InputSchema          map[string]any `json:"input_schema"`
	OutputUnit           string         `json:"output_unit"`
	MissingDataBehaviour string         `json:"missing_data_behaviour"`
	Limitations          string         `json:"limitations"`
	Version              string         `json:"version"`
	IsActive             bool           `json:"is_active"`
	CreatedAt            time.Time      `json:"created_at"`
}

// IndicatorResult represents a computed indicator for a building.
type IndicatorResult struct {
	ID                string    `json:"id"`
	BuildingID        string    `json:"building_id"`
	IndicatorName     string    `json:"indicator_name"`
	IndicatorVersion  string    `json:"indicator_version"`
	Score             float64   `json:"score"`
	Unit              string    `json:"unit"`
	CalculationMethod string    `json:"calculation_method"`
	InputDatasets     []string  `json:"input_datasets"`
	MissingInputs     []string  `json:"missing_inputs,omitempty"`
	Confidence        string    `json:"confidence"` // high, moderate, low, unverified
	Limitations       string    `json:"limitations"`
	CalculatedAt      time.Time `json:"calculated_at"`
}

// AnalysisJob describes a background calculation job.
type AnalysisJob struct {
	ID             string     `json:"id"`
	IndicatorName  string     `json:"indicator_name"`
	Status         string     `json:"status"` // pending, running, completed, failed
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	TotalBuildings int        `json:"total_buildings"`
	ProcessedCount int        `json:"processed_count"`
	ErrorLog       string     `json:"error_log,omitempty"`
}

// ListDefinitionsResponse wraps definitions response.
type ListDefinitionsResponse struct {
	Definitions []IndicatorDefinition `json:"definitions"`
	Total       int                   `json:"total"`
	Disclaimer  string                `json:"disclaimer"`
}
