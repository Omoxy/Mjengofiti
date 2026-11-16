-- Seed spatial sample records for Nairobi County building stock
-- These records are manually constructed synthetic test polygons for workflow testing and UI development.
-- They are explicitly labeled as unverified samples with true projected UTM Zone 37S footprint areas.

BEGIN;

-- 1. Core Dataset
INSERT INTO datasets (
    id, name, description, publisher, source_url, licence, licence_url, data_type, format, srid, geo_coverage, status
) VALUES (
    'a1000000-0000-0000-0000-000000000001',
    'Nairobi Sample Building Stock Development Baseline',
    'Illustrative building footprint test dataset comprising manually constructed rectangular sample bounding polygons for development, UI, and workflow testing. Not authoritative cadastral records.',
    'Mjengofiti Engineering Test Bed',
    'https://github.com/mjengofiti',
    'Open Database Licence Test Sample',
    'https://opendatacommons.org/licenses/odbl/',
    'vector',
    'geojson',
    4326,
    'Nairobi County',
    'development_sample'
) ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    status = EXCLUDED.status;

INSERT INTO dataset_versions (
    id, dataset_id, version, acquisition_at, checksum, file_size, feature_count, geo_coverage_pct, validation_status, promotion_status, promoted_at, promoted_by, notes
) VALUES (
    'b1000000-0000-0000-0000-000000000001',
    'a1000000-0000-0000-0000-000000000001',
    '2026.1',
    NOW(),
    'sha256:7e9f3b14a908c62d88192a5431bf90e4f16bc12093847291a27456bcdef98123',
    1048576,
    8,
    'unverified_sample',
    'sample_unverified',
    'development_only',
    NOW(),
    'lead_gis_steward',
    'Synthetic bounding polygons around approximate landmark locations for spatial UI testing. Height and area values are unverified estimates.'
) ON CONFLICT (dataset_id, version) DO UPDATE SET
    validation_status = EXCLUDED.validation_status,
    promotion_status = EXCLUDED.promotion_status,
    notes = EXCLUDED.notes;

-- 2. County Administrative Boundary (coarse 14-vertex boundary approximation)
INSERT INTO boundaries (
    id, name, boundary_type, geom, source, source_url, properties
) VALUES (
    'c1000000-0000-0000-0000-000000000001',
    'Nairobi City County',
    'county',
    ST_Multi(ST_SetSRID(ST_GeomFromGeoJSON('{
        "type": "Polygon",
        "coordinates": [[
            [36.652, -1.275],
            [36.705, -1.205],
            [36.782, -1.162],
            [36.850, -1.155],
            [36.930, -1.180],
            [37.010, -1.220],
            [37.085, -1.265],
            [37.102, -1.330],
            [37.075, -1.410],
            [36.995, -1.448],
            [36.910, -1.435],
            [36.820, -1.415],
            [36.735, -1.370],
            [36.670, -1.335],
            [36.652, -1.275]
        ]]
    }'), 4326)),
    'Kenya National Bureau of Statistics (KNBS) and IEBC Subnational Boundaries Coarse Approximation',
    'https://data.humdata.org/dataset/cod_ab_ken',
    '{"county_code": "047", "country": "Kenya", "area_sq_km": 696.0, "notes": "Coarse 14 vertex boundary approximation for viewport framing. Official statutory land area is 696 sq km per KNBS 2019 Census."}'::jsonb
) ON CONFLICT (id) DO UPDATE SET
    properties = EXCLUDED.properties;

-- 3. Synthetic Sample Landmark Buildings (Projected areas calculated via UTM 37S EPSG:32737)
-- Building 1: Times Tower
INSERT INTO buildings (
    id, external_id, source_dataset_id, source_name, geom, area_sq_m, height_m, height_method, match_method, match_confidence, data_quality
) VALUES (
    'd1000000-0000-0000-0000-000000000001',
    'NBI-CBD-001-TIMES',
    'a1000000-0000-0000-0000-000000000001',
    'Times Tower KRA Headquarters (Sample)',
    ST_SetSRID(ST_GeomFromGeoJSON('{
        "type": "Polygon",
        "coordinates": [[
            [36.82500, -1.29150],
            [36.82545, -1.29150],
            [36.82545, -1.29190],
            [36.82500, -1.29190],
            [36.82500, -1.29150]
        ]]
    }'), 4326),
    2216.53,
    140.0,
    'Unverified Architectural Estimate',
    'Manual Bounding Box Approximation',
    0.35,
    'unverified_sample'
) ON CONFLICT (id) DO UPDATE SET
    source_name = EXCLUDED.source_name,
    area_sq_m = EXCLUDED.area_sq_m,
    height_method = EXCLUDED.height_method,
    match_method = EXCLUDED.match_method,
    match_confidence = EXCLUDED.match_confidence,
    data_quality = EXCLUDED.data_quality;

-- Building 2: Kenyatta International Convention Centre
INSERT INTO buildings (
    id, external_id, source_dataset_id, source_name, geom, area_sq_m, height_m, height_method, match_method, match_confidence, data_quality
) VALUES (
    'd1000000-0000-0000-0000-000000000002',
    'NBI-CBD-002-KICC',
    'a1000000-0000-0000-0000-000000000001',
    'Kenyatta International Convention Centre (Sample)',
    ST_SetSRID(ST_GeomFromGeoJSON('{
        "type": "Polygon",
        "coordinates": [[
            [36.82300, -1.28840],
            [36.82360, -1.28840],
            [36.82360, -1.28890],
            [36.82300, -1.28890],
            [36.82300, -1.28840]
        ]]
    }'), 4326),
    3694.22,
    105.0,
    'Unverified Architectural Estimate',
    'Manual Bounding Box Approximation',
    0.35,
    'unverified_sample'
) ON CONFLICT (id) DO UPDATE SET
    source_name = EXCLUDED.source_name,
    area_sq_m = EXCLUDED.area_sq_m,
    height_method = EXCLUDED.height_method,
    match_method = EXCLUDED.match_method,
    match_confidence = EXCLUDED.match_confidence,
    data_quality = EXCLUDED.data_quality;

-- Building 3: Parliament of Kenya
INSERT INTO buildings (
    id, external_id, source_dataset_id, source_name, geom, area_sq_m, height_m, height_method, match_method, match_confidence, data_quality
) VALUES (
    'd1000000-0000-0000-0000-000000000003',
    'NBI-CBD-003-PARL',
    'a1000000-0000-0000-0000-000000000001',
    'Parliament of Kenya Main Chamber (Sample)',
    ST_SetSRID(ST_GeomFromGeoJSON('{
        "type": "Polygon",
        "coordinates": [[
            [36.81940, -1.28950],
            [36.82010, -1.28950],
            [36.82010, -1.29010],
            [36.81940, -1.29010],
            [36.81940, -1.28950]
        ]]
    }'), 4326),
    5171.93,
    42.0,
    'Unverified Architectural Estimate',
    'Manual Bounding Box Approximation',
    0.35,
    'unverified_sample'
) ON CONFLICT (id) DO UPDATE SET
    source_name = EXCLUDED.source_name,
    area_sq_m = EXCLUDED.area_sq_m,
    height_method = EXCLUDED.height_method,
    match_method = EXCLUDED.match_method,
    match_confidence = EXCLUDED.match_confidence,
    data_quality = EXCLUDED.data_quality;

-- Building 4: Nairobi City Hall Annex
INSERT INTO buildings (
    id, external_id, source_dataset_id, source_name, geom, area_sq_m, height_m, height_method, match_method, match_confidence, data_quality
) VALUES (
    'd1000000-0000-0000-0000-000000000004',
    'NBI-CBD-004-CITYHALL',
    'a1000000-0000-0000-0000-000000000001',
    'Nairobi City Hall Annex (Sample)',
    ST_SetSRID(ST_GeomFromGeoJSON('{
        "type": "Polygon",
        "coordinates": [[
            [36.82190, -1.28650],
            [36.82250, -1.28650],
            [36.82250, -1.28700],
            [36.82190, -1.28700],
            [36.82190, -1.28650]
        ]]
    }'), 4326),
    3694.23,
    58.0,
    'Unverified Architectural Estimate',
    'Manual Bounding Box Approximation',
    0.35,
    'unverified_sample'
) ON CONFLICT (id) DO UPDATE SET
    source_name = EXCLUDED.source_name,
    area_sq_m = EXCLUDED.area_sq_m,
    height_method = EXCLUDED.height_method,
    match_method = EXCLUDED.match_method,
    match_confidence = EXCLUDED.match_confidence,
    data_quality = EXCLUDED.data_quality;

-- Building 5: Britam Tower Commercial Highrise
INSERT INTO buildings (
    id, external_id, source_dataset_id, source_name, geom, area_sq_m, height_m, height_method, match_method, match_confidence, data_quality
) VALUES (
    'd1000000-0000-0000-0000-000000000005',
    'NBI-UPH-005-BRITAM',
    'a1000000-0000-0000-0000-000000000001',
    'Britam Tower Commercial Highrise (Sample)',
    ST_SetSRID(ST_GeomFromGeoJSON('{
        "type": "Polygon",
        "coordinates": [[
            [36.81660, -1.29940],
            [36.81720, -1.29940],
            [36.81720, -1.29990],
            [36.81660, -1.29990],
            [36.81660, -1.29940]
        ]]
    }'), 4326),
    3694.24,
    200.0,
    'Unverified Architectural Estimate',
    'Manual Bounding Box Approximation',
    0.35,
    'unverified_sample'
) ON CONFLICT (id) DO UPDATE SET
    source_name = EXCLUDED.source_name,
    area_sq_m = EXCLUDED.area_sq_m,
    height_method = EXCLUDED.height_method,
    match_method = EXCLUDED.match_method,
    match_confidence = EXCLUDED.match_confidence,
    data_quality = EXCLUDED.data_quality;

-- Building 6: UAP Old Mutual Tower
INSERT INTO buildings (
    id, external_id, source_dataset_id, source_name, geom, area_sq_m, height_m, height_method, match_method, match_confidence, data_quality
) VALUES (
    'd1000000-0000-0000-0000-000000000006',
    'NBI-UPH-006-UAP',
    'a1000000-0000-0000-0000-000000000001',
    'UAP Old Mutual Tower (Sample)',
    ST_SetSRID(ST_GeomFromGeoJSON('{
        "type": "Polygon",
        "coordinates": [[
            [36.81800, -1.29800],
            [36.81860, -1.29800],
            [36.81860, -1.29850],
            [36.81800, -1.29850],
            [36.81800, -1.29800]
        ]]
    }'), 4326),
    3694.23,
    163.0,
    'Unverified Architectural Estimate',
    'Manual Bounding Box Approximation',
    0.35,
    'unverified_sample'
) ON CONFLICT (id) DO UPDATE SET
    source_name = EXCLUDED.source_name,
    area_sq_m = EXCLUDED.area_sq_m,
    height_method = EXCLUDED.height_method,
    match_method = EXCLUDED.match_method,
    match_confidence = EXCLUDED.match_confidence,
    data_quality = EXCLUDED.data_quality;

-- Building 7: Sarit Centre Retail and Commercial Wing
INSERT INTO buildings (
    id, external_id, source_dataset_id, source_name, geom, area_sq_m, height_m, height_method, match_method, match_confidence, data_quality
) VALUES (
    'd1000000-0000-0000-0000-000000000007',
    'NBI-WST-007-SARIT',
    'a1000000-0000-0000-0000-000000000001',
    'Sarit Centre Retail Wing (Sample)',
    ST_SetSRID(ST_GeomFromGeoJSON('{
        "type": "Polygon",
        "coordinates": [[
            [36.80380, -1.26160],
            [36.80470, -1.26160],
            [36.80470, -1.26240],
            [36.80380, -1.26240],
            [36.80380, -1.26160]
        ]]
    }'), 4326),
    8866.45,
    35.0,
    'Unverified Architectural Estimate',
    'Manual Bounding Box Approximation',
    0.35,
    'unverified_sample'
) ON CONFLICT (id) DO UPDATE SET
    source_name = EXCLUDED.source_name,
    area_sq_m = EXCLUDED.area_sq_m,
    height_method = EXCLUDED.height_method,
    match_method = EXCLUDED.match_method,
    match_confidence = EXCLUDED.match_confidence,
    data_quality = EXCLUDED.data_quality;

-- Building 8: NSSF Complex
INSERT INTO buildings (
    id, external_id, source_dataset_id, source_name, geom, area_sq_m, height_m, height_method, match_method, match_confidence, data_quality
) VALUES (
    'd1000000-0000-0000-0000-000000000008',
    'NBI-COM-008-NSSF',
    'a1000000-0000-0000-0000-000000000001',
    'NSSF Building Block A (Sample)',
    ST_SetSRID(ST_GeomFromGeoJSON('{
        "type": "Polygon",
        "coordinates": [[
            [36.81120, -1.29530],
            [36.81180, -1.29530],
            [36.81180, -1.29590],
            [36.81120, -1.29590],
            [36.81120, -1.29530]
        ]]
    }'), 4326),
    4433.12,
    85.0,
    'Unverified Architectural Estimate',
    'Manual Bounding Box Approximation',
    0.35,
    'unverified_sample'
) ON CONFLICT (id) DO UPDATE SET
    source_name = EXCLUDED.source_name,
    area_sq_m = EXCLUDED.area_sq_m,
    height_method = EXCLUDED.height_method,
    match_method = EXCLUDED.match_method,
    match_confidence = EXCLUDED.match_confidence,
    data_quality = EXCLUDED.data_quality;

COMMIT;
