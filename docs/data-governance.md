# Mjengofiti — Data Governance, Quality, and Provenance Policy

## 1. Principles of Data Governance
1. **Verifiable Provenance:** Every building footprint, elevation raster, rainfall observation, and regulatory clause registered in Mjengofiti must have an explicit publisher, source URL, acquisition date, cryptographic checksum, and licence declaration.
2. **Immutability of Ingested Datasets:** Raw ingested datasets are archived as read-only versions. Any transformation produces a new version with an audit log.
3. **Promotion Quality Gates:** Datasets start in `draft` or `pending` status. A dataset can only be promoted to `promoted` (live operational use) after passing automated validation.
4. **Honest Quality Flags:** If an attribute is missing (such as building height or contractor registration number), it must be recorded as `null` or `unverified`. Fabricating data or guessing values is strictly forbidden.

---

## 2. Ingestion and Validation Lifecycle

```
[Register Source] -> [Acquire & Checksum] -> [Run Validation Report] -> [Supervisor Review] -> [Promote / Reject]
```

### 2.1 Automated Validation Criteria
1. **Coordinate Reference System (CRS):** Must be valid EPSG:4326 (WGS84).
2. **Geometry Integrity:** PostGIS `ST_IsValid(geom)` must equal `true`. Self-intersecting rings or open polygons are rejected.
3. **Spatial Envelope:** Geometries are screened against the verified Nairobi ADM1 geometry (`ST_Intersects(geom, nairobi_boundary)`). The current geoBoundaries source is an alternative, not a statutory official boundary; this screen is not a legal determination.
4. **Duplicate Detection:** Spatial overlap > 90% between footprints from the same source is flagged as duplicate candidates.
5. **Report Generation:** Generates a structured JSON validation report detailing:
   - Total records
   - Valid geometry count
   - Invalid geometry count
   - Duplicate count
   - Spatial bounds validity

### 2.2 Data Retention and Retirement
- **Active Operational Datasets:** Served to map layers and used for indicator calculations.
- **Superseded Datasets:** Archived with status `retired`. Historical indicators retain references to the exact version used during calculation for reproducibility.
- **Incident & Inspection Records:** Permanent retention under Kenya National Archives and Documentation Service requirements.
