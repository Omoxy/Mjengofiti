package indicators

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles indicator definitions and jobs.
type Repository struct {
	pool      *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetDefinitions returns the documented, explainable indicator definitions.
func (r *Repository) GetDefinitions(ctx context.Context) ([]IndicatorDefinition, error) {
	// Built-in verified indicators for Mjengofiti
	created := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	return []IndicatorDefinition{
		{
			ID:          "ind-density-01",
			Name:        "built_up_density_index",
			DisplayName: "Neighborhood Built-Up Density Index",
			Description: "Measures spatial building footprint density within a 200m circular neighborhood radius.",
			Purpose:     "Identifies extreme spatial overcrowding that complicates emergency vehicular access, firefighting, and structural settlement monitoring.",
			CalculationMethod: "Total building footprint area within 200m buffer divided by buffer area: (SUM(area_m2) / (PI * 200^2)) * 100",
			InputSchema: map[string]any{
				"buffer_radius_m": 200,
				"min_confidence":  0.65,
				"required_layers": []string{"building_footprints"},
			},
			OutputUnit:           "percentage",
			MissingDataBehaviour: "Returns null if fewer than 50% of buildings in the buffer have verified polygon footprints.",
			Limitations:          "Does not account for vertical airspace; multi-storey towers in high-density informal settlements may share footprints with small single-storey shacks.",
			Version:              "1.1.0",
			IsActive:             true,
			CreatedAt:            created,
		},
		{
			ID:          "ind-height-aspect-02",
			Name:        "height_to_footprint_slenderness",
			DisplayName: "Building Slenderness & Height-to-Base Ratio",
			Description: "Calculates the ratio of estimated/measured building height against the effective base width (square root of ground footprint area).",
			Purpose:     "Screens for disproportionately tall, narrow multi-storey residential tenements (often constructed without geotechnical soil piling or structural setbacks in densifying sub-counties like Pipeline, Embakasi, and Kasarani).",
			CalculationMethod: "height_m / SQRT(footprint_area_sq_m)",
			InputSchema: map[string]any{
				"height_attribute":   "height_m",
				"area_attribute":     "area_sq_m",
				"critical_threshold": 3.5,
			},
			OutputUnit:           "ratio",
			MissingDataBehaviour: "Returns null and marks record as 'unverified_height' if height is not measured or satellite confidence is below 0.70.",
			Limitations:          "Satellite shadow height derivations have +/- 2.5m error margin; this indicator is strictly a screening trigger for professional visual inspection, NOT proof of structural defect.",
			Version:              "1.0.0",
			IsActive:             true,
			CreatedAt:            created,
		},
		{
			ID:          "ind-riparian-buffer-03",
			Name:        "riparian_flood_proximity",
			DisplayName: "Riparian Reserve & Flood Drainage Proximity",
			Description: "Measures Euclidean distance from the building footprint perimeter to the nearest classified natural river course or low-lying drainage depression (Nairobi, Ngong, Mathare Rivers).",
			Purpose:     "Assesses statutory compliance with Physical and Land Use Planning Act 2019 Section 69 (30m buffer) and foundation vulnerability to severe seasonal flash flooding.",
			CalculationMethod: "ST_Distance(building.geom::geography, river_centerline.geom::geography)",
			InputSchema: map[string]any{
				"buffer_zones_m":  []int{30, 50, 100},
				"river_dataset":   "nairobi_river_network_v1",
				"elevation_model": "copernicus_dem_30m",
			},
			OutputUnit:           "meters",
			MissingDataBehaviour: "Defaults to 'unassessed' if river centerline dataset has spatial gaps in the target sub-county.",
			Limitations:          "Surface distance does not measure finished floor foundation height or engineered river embankment retaining walls.",
			Version:              "1.2.0",
			IsActive:             true,
			CreatedAt:            created,
		},
		{
			ID:          "ind-data-completeness-04",
			Name:        "data_completeness_index",
			DisplayName: "Dataset Completeness & Quality Confidence Score",
			Description: "Assesses the presence and verification level of core building metadata (footprint geometry, verified height, street address, county approval record, inspection history).",
			Purpose:     "Explicitly quantifies epistemic uncertainty, ensuring inspectors know how much information is unknown or unverified before deploying to the field.",
			CalculationMethod: "Weighted composite: (GeometryValid * 0.25) + (HeightVerified * 0.25) + (ApprovalRecordFound * 0.25) + (RecentAuditWithin24M * 0.25)",
			InputSchema: map[string]any{
				"weights": map[string]float64{
					"geometry": 0.25,
					"height":   0.25,
					"approval": 0.25,
					"audit":    0.25,
				},
			},
			OutputUnit:           "score_0_to_1",
			MissingDataBehaviour: "Missing elements score 0; total score is adjusted by known dataset acquisition age.",
			Limitations:          "Low completeness indicates lack of digitized municipal records, not necessarily building illegality.",
			Version:              "1.0.0",
			IsActive:             true,
			CreatedAt:            created,
		},
	}, nil
}

// GetDefinitionByName retrieves a specific definition.
func (r *Repository) GetDefinitionByName(ctx context.Context, name string) (*IndicatorDefinition, error) {
	defs, err := r.GetDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	for _, d := range defs {
		if d.Name == name || d.ID == name {
			return &d, nil
		}
	}
	return nil, fmt.Errorf("indicator definition not found: %s", name)
}

// CreateAnalysisJob starts an analysis job.
