-- Correction script for Mjengofiti spatial data provenance
-- Accurately labels manually constructed sample polygons and corrects projected areas

BEGIN;

-- 1. Update Dataset Classification
UPDATE datasets
SET
    name = 'Nairobi Sample Building Stock Development Baseline',
    description = 'Illustrative building footprint test dataset comprising manually constructed rectangular sample bounding polygons for development, UI, and workflow testing. Not authoritative cadastral records.',
    licence = 'Open Database Licence Test Sample',
    status = 'development_sample'
WHERE id = 'a1000000-0000-0000-0000-000000000001';

UPDATE dataset_versions
SET
    validation_status = 'sample_unverified',
    promotion_status = 'development_only',
    notes = 'Synthetic bounding polygons around approximate landmark locations for spatial UI testing. Height and area values are unverified estimates.'
WHERE dataset_id = 'a1000000-0000-0000-0000-000000000001';

-- 2. Update Buildings with True Projected UTM 37S Areas and Truthful Provenance
UPDATE buildings
SET
    area_sq_m = ROUND(ST_Area(ST_Transform(geom, 32737))::numeric, 2),
    height_method = 'Unverified Architectural Estimate',
    match_method = 'Manual Bounding Box Approximation',
    match_confidence = 0.35,
    data_quality = 'unverified_sample'
WHERE source_dataset_id = 'a1000000-0000-0000-0000-000000000001';

-- 3. Update Observations to Reflect Illustrative Nature
UPDATE building_observations
SET
    method = 'Illustrative Sample Test Data',
    confidence = 0.30,
    limitations = 'Manually generated test fixture for UI verification only. Not an actual forensic or municipal inspection report.'
WHERE source_version = '2026.1';

-- 4. Update Indicators to Reflect Illustrative Nature
UPDATE building_indicators
SET
    calculation_method = 'Simulated baseline calculation for development testing',
    uncertainty = 'High: Unverified Test Sample',
    limitations = 'Values are synthetic demonstration figures and must not be used for municipal planning or structural certification.'
WHERE version = '1.0.0';

-- 5. Update Boundary Notes
UPDATE boundaries
SET
    properties = jsonb_set(
        properties,
        '{notes}',
        '"Coarse 14 vertex boundary approximation for viewport framing. Official statutory land area is 696 sq km per KNBS 2019 Census."'
    )
WHERE id = 'c1000000-0000-0000-0000-000000000001';

COMMIT;
