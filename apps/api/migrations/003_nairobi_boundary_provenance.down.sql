UPDATE boundaries
SET properties = properties
    - 'boundary_dataset_name'
    - 'boundary_identifier'
    - 'source_data_update_date'
    - 'postgis_validation'
WHERE id = 'c2000000-0000-0000-0000-000000000001'
  AND source = 'geoBoundaries / RCMRD GeoPortal / Africa GeoPortal';