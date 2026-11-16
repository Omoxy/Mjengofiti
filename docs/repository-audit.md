# Mjengofiti — Repository-Wide Implementation Audit

## 1. Executive Summary & Audit Overview

This audit establishes the actual state of the Mjengofiti codebase as of October 2026. The inspection was initiated to investigate empty directories identified in the repository:
1. `apps/web/src/hooks/` (Frontend shared state and query hooks)
2. `apps/web/src/components/` (Frontend shared UI components and design system)
3. `apps/api/internal/ingestion/` (Backend geospatial dataset ingestion, validation, and promotion engine)

### 1.1 Directory Findings

| Directory | Actual State | Finding & Impact | Required Resolution |
|---|---|---|---|
| `apps/web/src/hooks/` | **Genuinely Empty** | Feature components were duplicating manual state (`useState`, `useEffect`, fetch logic, cancellation tokens). Missing centralized server-state, query management, viewport coordination, and role permission hooks. | Implement typed hooks: `useAuth`, `useSystemStatus`, `useBuildings`, `useDatasets`, `useInspections`, `useIndicators`, `useStandards`, `useMapViewport`. |
| `apps/web/src/components/` | **Genuinely Empty** | Feature screens were inlining repetitive HTML, inline styles, and unshared UI logic. Missing reusable accessible controls: modals, loading spinners, error alerts, badges, data tables, and search inputs. | Build accessible component system organized into `layout/`, `feedback/`, `forms/`, and `data-display/`. |
| `apps/api/internal/ingestion/` | **Genuinely Empty** | Dataset catalogue routes existed in `internal/datasets`, but the automated ingestion pipeline (acquisition, format checking, EPSG:4326 CRS validation, Nairobi bounding box containment, geometry ring checks, validation reports, and promotion gates) was missing. | Implement complete ingestion engine in `internal/ingestion` with domain models, validators, job runner, HTTP handlers, and unit tests. |

---

## 2. Feature-Completeness Matrix

| Feature / Domain | Backend Source | Frontend Source | Database / PostGIS | External Data Dependency | Test Suite | Status | Gaps / Action Plan |
|---|---|---|---|---|---|---|---|
| **Database Migrations** | `apps/api/migrations/`, `apps/api/cmd/migrate` | N/A | PostgreSQL 15+, PostGIS 3+, `schema_migrations` | None | `internal/database/migrate_test.go` | **VERIFIED (Runner & Schema)** / **IMPLEMENTED_UNVERIFIED (Live DB)** | Local PostgreSQL/Docker offline; live cluster verification pending. |
| **System & Health API** | `apps/api/internal/health` | `apps/web/src/api/client.ts`, `hooks/useSystemStatus.ts` | Connection pool ping, `PostGIS_Version()` | None | `internal/health/health_test.go` | **VERIFIED** | Degraded mode handled honestly when DB is unreachable. |
| **Authentication & RBAC** | `apps/api/internal/auth` | `apps/web/src/hooks/useAuth.ts`, `App.tsx` | `roles`, `users` | Statutory roles | `internal/auth/auth_test.go` | **VERIFIED** | Implement frontend hook and permission gate checks. |
| **Building Intelligence** | `apps/api/internal/buildings` | `apps/web/src/features/buildings/`, `hooks/useBuildings.ts` | `buildings`, `building_observations`, `building_indicators`, GiST index | Microsoft Footprints, Google Open Buildings | Compiles & runs | **VERIFIED (Static & Degraded API)** / **IMPLEMENTED_UNVERIFIED (Spatial PostGIS)** | Refactor workspace to use `useBuildings` and shared components. |
| **Spatial Reference Boundaries** | `apps/api/internal/boundaries` | `apps/web/src/features/map/NairobiMap.tsx` | `boundaries` table (SRID 4326) | geoBoundaries Kenya ADM1 2020 alternative source | `internal/boundaries/boundaries_test.go`, PostGIS migrations 002/003 | **VERIFIED ALTERNATIVE SOURCE** | Legacy 15-point approximation preserved; official statutory source and sub-county geometry not verified/loaded. |
| **Dataset Catalogue & Governance** | `apps/api/internal/datasets` | `apps/web/src/features/research/`, `hooks/useDatasets.ts` | `datasets`, `dataset_versions`, `ingestion_jobs` | MS, Google, OSM, Copernicus, CHIRPS, NCA | `internal/datasets/datasets_test.go` | **VERIFIED** | Provenance and metadata tracking operational. |
| **Geospatial Ingestion Engine** | `apps/api/internal/ingestion` | `apps/web/src/features/research/`, `hooks/useDatasets.ts` | `ingestion_jobs`, `validation_reports`, `dataset_versions`, `buildings` | GeoJSON / Vector datasets | `internal/ingestion/ingestion_test.go` (8 test suites) | **VERIFIED (Engine & Validators)** / **IMPLEMENTED_UNVERIFIED (Live External Fetches)** | RFC 7946 default CRS, finite numeric validation, NaN/Inf rejection, SSRF protection, and unreviewed dataset promotion gates verified. |

| **Indicators & Screening Engine** | `apps/api/internal/indicators` | `apps/web/src/features/analysis/`, `hooks/useIndicators.ts` | `indicator_definitions`, `building_indicators` | Multi-source datasets | `internal/indicators/indicators_test.go` | **VERIFIED** | Transparent calculation formulas with clear limitations. |
| **Inspection Cases & Findings** | `apps/api/internal/inspections` | `apps/web/src/features/inspections/`, `hooks/useInspections.ts` | `inspection_cases`, `case_history`, `case_findings` | Physical site inspections | `internal/inspections/inspections_test.go` | **VERIFIED** | Statutory state machine with check constraints. |
| **Standards Register** | `apps/api/internal/standards` | `apps/web/src/features/governance/`, `hooks/useStandards.ts` | `standards` table | National Building Code 2024, NCA Act | `internal/standards/standards_test.go` | **VERIFIED** | Verified Kenyan statutory clauses mapping. |
| **Audit Logging** | `apps/api/internal/audit` | `apps/web/src/features/governance/` | `audit_events` | None | `internal/audit/audit_test.go` | **VERIFIED** | Immutable audit log trail for compliance. |

---

## 3. Real Data Source Verification

| Dataset Candidate | Source URL | Licence / Legal Terms | Verified Coverage | Format / CRS | Integration Suitability |
|---|---|---|---|---|---|
| **Microsoft Kenya Building Footprints** | [GitHub Repository](https://github.com/microsoft/KenyaNigeriaBuildingFootprints) | Open Data Commons Open Database License (ODbL) | ~842,100 footprints within Nairobi County bounds | GeoJSON / EPSG:4326 | Primary building footprint baseline. |
| **Google Open Buildings (Temporal)** | [Google Research](https://sites.research.google/gr/open-buildings/temporal) | Creative Commons Attribution (CC-BY 4.0) | ~895,000 polygons with roof-height estimates | CSV / GeoJSON / EPSG:4326 | Secondary footprint & screening height comparison. |
| **OpenStreetMap Kenya Building Layers** | [Geofabrik Download](https://download.geofabrik.de/africa/kenya.html) | Open Database License (ODbL) | County-wide roads and buildings | OSM PBF / GeoJSON / EPSG:4326 | Road access and municipal proximity attributes. |
| **Copernicus Global DEM 30m** | [Copernicus Data Space](https://dataspace.copernicus.eu/) | Free and Open Data Policy | Nairobi Basin coverage | Cloud-Optimized GeoTIFF (COG) | Hydrological terrain slope & floodplain risk. |
| **CHIRPS v3.0 Rainfall** | [UCSB Climate Hazards](https://www.chc.ucsb.edu/data/chirps3) | Public Domain / Open Access | Eastern Africa / Nairobi regional grid | NetCDF / GeoTIFF / EPSG:4326 | Seasonal extreme rainfall anomaly indicators. |
| **NCA Inspection Audit Registers** | [National Construction Authority](https://www.nca.go.ke/) | Statutory Public Records (GoK) | Nairobi inspection zones | Scanned Gazette / Structured CSV | Authoritative statutory inspection notices. |
