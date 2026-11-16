DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM boundaries
        WHERE id = 'c2000000-0000-0000-0000-000000000001'
          AND source = 'geoBoundaries / RCMRD GeoPortal / Africa GeoPortal'
          AND ST_IsValid(geom)
          AND NOT ST_IsEmpty(geom)
          AND GeometryType(geom) = 'MULTIPOLYGON'
          AND ST_SRID(geom) = 4326
    ) THEN
        RAISE EXCEPTION 'Verified geoBoundaries Nairobi ADM1 record is missing or invalid';
    END IF;
END $$;

UPDATE boundaries
SET properties = properties || jsonb_build_object(
    'boundary_dataset_name', 'geoBoundaries Kenya ADM1 2020',
    'boundary_identifier', 'KEN-ADM1-32016919',
    'source_data_update_date', '2023-01-19',
    'postgis_validation', jsonb_build_object(
        'is_valid', ST_IsValid(geom),
        'is_empty', ST_IsEmpty(geom),
        'geometry_type', GeometryType(geom),
        'srid', ST_SRID(geom),
        'vertex_count', ST_NPoints(geom),
        'bounds_wgs84', jsonb_build_array(
            ST_XMin(Box3D(geom)), ST_YMin(Box3D(geom)),
            ST_XMax(Box3D(geom)), ST_YMax(Box3D(geom))
        )
    )
)
WHERE id = 'c2000000-0000-0000-0000-000000000001'
  AND source = 'geoBoundaries / RCMRD GeoPortal / Africa GeoPortal';