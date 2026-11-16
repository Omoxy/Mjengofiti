package buildings

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides data access for buildings.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new buildings repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// FindByID retrieves a single building by its internal ID.
func (r *Repository) FindByID(ctx context.Context, id string) (*BuildingDetail, error) {
	query := `
		SELECT
			b.id, b.external_id, b.source_dataset_id, b.source_name,
			ST_X(ST_Centroid(b.geom)) as longitude,
			ST_Y(ST_Centroid(b.geom)) as latitude,
			b.area_sq_m, b.height_m, b.height_method,
			b.match_method, b.match_confidence, b.data_quality,
			b.created_at, b.updated_at,
			ST_AsGeoJSON(b.geom) as geojson
		FROM buildings b
		WHERE b.id = $1
	`

	var bd BuildingDetail
	err := r.db.QueryRow(ctx, query, id).Scan(
		&bd.ID, &bd.ExternalID, &bd.SourceDatasetID, &bd.SourceName,
		&bd.Longitude, &bd.Latitude,
		&bd.AreaSqM, &bd.HeightM, &bd.HeightMethod,
		&bd.MatchMethod, &bd.MatchConfidence, &bd.DataQuality,
		&bd.CreatedAt, &bd.UpdatedAt,
		&bd.GeoJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("finding building %s: %w", id, err)
	}

	// Load observations
	bd.Observations, err = r.findObservations(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("loading observations for building %s: %w", id, err)
	}

	// Load indicators
	bd.Indicators, err = r.findIndicators(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("loading indicators for building %s: %w", id, err)
	}

	return &bd, nil
}

// FindByBBox retrieves buildings within a bounding box.
func (r *Repository) FindByBBox(ctx context.Context, q BBoxQuery) (*ListResult, error) {
	if q.Limit <= 0 || q.Limit > MaxBBoxLimit {
		q.Limit = DefaultLimit
	}
	if q.Offset < 0 {
		q.Offset = 0
	}

	envelope := fmt.Sprintf(
		"ST_MakeEnvelope(%f, %f, %f, %f, 4326)",
		q.MinLng, q.MinLat, q.MaxLng, q.MaxLat,
	)

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM buildings b
		WHERE b.geom && %s
	`, envelope)

	var total int
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, fmt.Errorf("counting buildings in bbox: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT
			b.id, b.external_id, b.source_dataset_id, b.source_name,
			ST_X(ST_Centroid(b.geom)) as longitude,
			ST_Y(ST_Centroid(b.geom)) as latitude,
			b.area_sq_m, b.height_m, b.height_method,
			b.match_method, b.match_confidence, b.data_quality,
			b.created_at, b.updated_at
		FROM buildings b
		WHERE b.geom && %s
		ORDER BY b.id
		LIMIT $1 OFFSET $2
	`, envelope)

	rows, err := r.db.Query(ctx, query, q.Limit, q.Offset)
	if err != nil {
		return nil, fmt.Errorf("querying buildings in bbox: %w", err)
	}
	defer rows.Close()

	var buildings []Building
	for rows.Next() {
		var b Building
		if err := rows.Scan(
			&b.ID, &b.ExternalID, &b.SourceDatasetID, &b.SourceName,
			&b.Longitude, &b.Latitude,
			&b.AreaSqM, &b.HeightM, &b.HeightMethod,
			&b.MatchMethod, &b.MatchConfidence, &b.DataQuality,
			&b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning building row: %w", err)
		}
		buildings = append(buildings, b)
	}

	if buildings == nil {
		buildings = []Building{}
	}

	return &ListResult{
		Buildings: buildings,
		Total:     total,
		Limit:     q.Limit,
		Offset:    q.Offset,
	}, nil
}

// Search finds buildings matching text and attribute criteria.
func (r *Repository) Search(ctx context.Context, q SearchQuery) (*ListResult, error) {
	if q.Limit <= 0 || q.Limit > MaxSearchLimit {
		q.Limit = DefaultLimit
	}
	if q.Offset < 0 {
		q.Offset = 0
	}

	// Build parameterised query
	where := "WHERE 1=1"
	args := []any{}
	argIdx := 1

	if q.Text != "" {
		where += fmt.Sprintf(" AND (b.external_id ILIKE $%d OR b.source_name ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+q.Text+"%")
		argIdx++
	}

	if q.SourceName != nil {
		where += fmt.Sprintf(" AND b.source_name = $%d", argIdx)
		args = append(args, *q.SourceName)
		argIdx++
	}

	if q.MinArea != nil {
		where += fmt.Sprintf(" AND b.area_sq_m >= $%d", argIdx)
		args = append(args, *q.MinArea)
		argIdx++
	}

	if q.MaxArea != nil {
		where += fmt.Sprintf(" AND b.area_sq_m <= $%d", argIdx)
		args = append(args, *q.MaxArea)
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM buildings b %s", where)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("counting search results: %w", err)
	}

	selectQuery := fmt.Sprintf(`
		SELECT
			b.id, b.external_id, b.source_dataset_id, b.source_name,
			ST_X(ST_Centroid(b.geom)) as longitude,
			ST_Y(ST_Centroid(b.geom)) as latitude,
			b.area_sq_m, b.height_m, b.height_method,
			b.match_method, b.match_confidence, b.data_quality,
			b.created_at, b.updated_at
		FROM buildings b
		%s
		ORDER BY b.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)

	args = append(args, q.Limit, q.Offset)

	rows, err := r.db.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("searching buildings: %w", err)
	}
	defer rows.Close()

	var buildings []Building
	for rows.Next() {
		var b Building
		if err := rows.Scan(
			&b.ID, &b.ExternalID, &b.SourceDatasetID, &b.SourceName,
			&b.Longitude, &b.Latitude,
			&b.AreaSqM, &b.HeightM, &b.HeightMethod,
			&b.MatchMethod, &b.MatchConfidence, &b.DataQuality,
			&b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning search result: %w", err)
		}
		buildings = append(buildings, b)
	}

	if buildings == nil {
		buildings = []Building{}
	}

	return &ListResult{
		Buildings: buildings,
		Total:     total,
		Limit:     q.Limit,
		Offset:    q.Offset,
	}, nil
}

func (r *Repository) findObservations(ctx context.Context, buildingID string) ([]Observation, error) {
	query := `
		SELECT id, building_id, type, value, unit, observation_date,
		       source_version, method, confidence, limitations, created_at
		FROM building_observations
		WHERE building_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, buildingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var obs []Observation
	for rows.Next() {
		var o Observation
		if err := rows.Scan(
			&o.ID, &o.BuildingID, &o.Type, &o.Value, &o.Unit, &o.ObservationDate,
			&o.SourceVersion, &o.Method, &o.Confidence, &o.Limitations, &o.CreatedAt,
		); err != nil {
			return nil, err
		}
		obs = append(obs, o)
	}

	if obs == nil {
		obs = []Observation{}
	}
	return obs, nil
}

func (r *Repository) findIndicators(ctx context.Context, buildingID string) ([]Indicator, error) {
	query := `
		SELECT id, building_id, indicator_type, version, value, unit,
		       calculation_method, input_datasets, missing_inputs,
		       uncertainty, limitations, calculated_at
		FROM building_indicators
		WHERE building_id = $1
		ORDER BY calculated_at DESC
	`
	rows, err := r.db.Query(ctx, query, buildingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inds []Indicator
	for rows.Next() {
		var ind Indicator
		if err := rows.Scan(
			&ind.ID, &ind.BuildingID, &ind.IndicatorType, &ind.Version,
			&ind.Value, &ind.Unit, &ind.CalculationMethod,
			&ind.InputDatasets, &ind.MissingInputs,
			&ind.Uncertainty, &ind.Limitations, &ind.CalculatedAt,
		); err != nil {
			return nil, err
		}
		inds = append(inds, ind)
	}

	if inds == nil {
		inds = []Indicator{}
	}
	return inds, nil
}
