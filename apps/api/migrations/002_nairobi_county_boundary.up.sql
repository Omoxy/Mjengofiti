-- Add the pinned geoBoundaries Kenya ADM1 Nairobi feature without overwriting the legacy coarse record.
INSERT INTO boundaries (
    id, name, boundary_type, geom, source, source_url, acquired_at, properties
) VALUES (
    'c2000000-0000-0000-0000-000000000001',
    'Nairobi City County (geoBoundaries ADM1 2020)',
    'county',
    ST_Multi(ST_SetSRID(ST_GeomFromGeoJSON(($1::jsonb #> '{features,0,geometry}')::text), 4326)),
    'geoBoundaries / RCMRD GeoPortal / Africa GeoPortal',
    'https://github.com/wmgeolab/geoBoundaries/raw/9469f09592ced973a3448cf66b6100b741b64c0d/releaseData/gbOpen/KEN/ADM1/geoBoundaries-KEN-ADM1.geojson',
    '2026-11-09T00:00:00Z',
    ($1::jsonb #> '{features,0,properties}') || jsonb_build_object('record_status', 'verified_alternative')
);