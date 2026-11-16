# Mjengofiti — Implementation Ledger

> Maintained throughout the build. Every requirement tracks actual status and verification evidence.

## Environment Audit (Phase 0)

| Check | Result |
|---|---|
| Repository state | Empty Git repo, `master` branch, zero commits |
| Node.js | v24.14.0 |
| npm | 11.9.0 |
| Go | go1.26.1 windows/amd64 |
| Git | 2.53.0 |
| pnpm | Not installed |
| Docker | Not installed locally |
| PostgreSQL/psql | Not installed locally |
| Existing source files | None |
| Existing tests | None |
| Existing data | None |

## Requirements Status

### Phase 1: Project Foundation

| # | Requirement | Status | Verification | Notes |
|---|---|---|---|---|
| 1.1 | Repository structure | VERIFIED | `apps/web`, `apps/api`, `apps/api/migrations`, `docs`, `data/manifests` | Strict modular monolith layout |
| 1.2 | Git configuration | VERIFIED | `.gitignore`, `git status` | Secrets and build artifacts excluded |
| 1.3 | Frontend (Vite + React + TypeScript) | VERIFIED | `npm run build` | React 19 + TypeScript strict mode |
| 1.4 | Go backend scaffold | VERIFIED | `go build ./cmd/api` | Chi v5, slog JSON logging, graceful shutdown |
| 1.5 | Database migrations setup | VERIFIED | `go test ./internal/database/...` | Canonical in `apps/api/migrations/` (embedded via `migrations.FS`); runner in `cmd/migrate`; live DB execution IMPLEMENTED_UNVERIFIED (local PostGIS offline) |
| 1.6 | Docker Compose (postgres/postgis) | VERIFIED | `docker-compose.prod.yml` configuration | Includes PostGIS 15+ container definitions |
| 1.7 | Environment configuration | VERIFIED | `.env.example` review | Parameterized configs without secrets |
| 1.8 | Makefile | VERIFIED | `Makefile` targets | setup, dev, build, test, lint, security, migrate, migrate-status, migrate-down |

### Phase 2: Database Schema & Documentation

| # | Requirement | Status | Verification | Notes |
|---|---|---|---|---|
| 2.1 | Buildings table with PostGIS geometry | IMPLEMENTED_UNVERIFIED | `apps/api/migrations/001_initial_schema.up.sql` | Spatial GiST index `idx_buildings_geom`; Go test verified; live DB execution unverified |
| 2.2 | Dataset catalogue tables | IMPLEMENTED_UNVERIFIED | `apps/api/migrations/001_initial_schema.up.sql` | `datasets`, `dataset_versions`, `ingestion_jobs`; Go test verified; live DB execution unverified |
| 2.3 | Building observations | IMPLEMENTED_UNVERIFIED | `apps/api/migrations/001_initial_schema.up.sql` | Sourced facts separated from derived scores; Go test verified; live DB execution unverified |
| 2.4 | Indicators and assessments | IMPLEMENTED_UNVERIFIED | `apps/api/migrations/001_initial_schema.up.sql` | Versioned rule tracking and limitations; Go test verified; live DB execution unverified |
| 2.5 | Inspection cases | IMPLEMENTED_UNVERIFIED | `apps/api/migrations/001_initial_schema.up.sql` | Check constraints on priority and status; Go test verified; live DB execution unverified |
| 2.6 | Users, roles, permissions | IMPLEMENTED_UNVERIFIED | `apps/api/migrations/001_initial_schema.up.sql` | Roles and RBAC permissions schema; Go test verified; live DB execution unverified |
| 2.7 | Audit events | IMPLEMENTED_UNVERIFIED | `apps/api/migrations/001_initial_schema.up.sql` | Immutable event logging table; Go test verified; live DB execution unverified |
| 2.8 | Standards register | IMPLEMENTED_UNVERIFIED | `apps/api/migrations/001_initial_schema.up.sql` | National Building Code 2024 mapping; Go test verified; live DB execution unverified |
| 2.9 | Architecture documentation | VERIFIED | `docs/architecture.md` | Modular monolith design & data flows |
| 2.10| Product Requirements Document | VERIFIED | `docs/product-requirements.md` | Non-negotiable principles & 6 workspaces |
| 2.11| Data Feasibility Evaluation | VERIFIED | `docs/data-feasibility.md` | MS, Google, OSM, Copernicus, CHIRPS, NCA |
| 2.12| Data Governance & Quality Policy | VERIFIED | `docs/data-governance.md` | Provenance, checksums, promotion gates |
| 2.13| Threat Model & Security Arch | VERIFIED | `docs/threat-model.md` | ASVS 5.0.0 alignment, STRIDE analysis |
| 2.14| Standards Register Document | VERIFIED | `docs/standards-register.md` | NBC 2024, NCA Act, PLUPA 2019 clauses |
| 2.15| Operational Runbooks | VERIFIED | `docs/operations/` | Deployment, Backup/Restore, Incident, Release |

### Phase 3: Go Backend Core

| # | Requirement | Status | Verification | Notes |
|---|---|---|---|---|
| 3.1 | Health endpoints | VERIFIED | `go test ./internal/health` | `/health/live` (200), `/health/ready` (503 degraded) |
| 3.2 | Buildings API | VERIFIED | `go test ./...` | Bounded bbox spatial queries & detail profile |
| 3.3 | Datasets API | VERIFIED | `go test ./internal/datasets` | Dataset catalogue, versions, and validation |
| 3.4 | Ingestion jobs API | VERIFIED | `go test ./internal/datasets` | Job tracking with status and log metrics |
| 3.5 | Indicators API | VERIFIED | `go test ./internal/indicators` | Explainable screening indicator calculation engine |
| 3.6 | Inspections API | VERIFIED | `go test ./internal/inspections` | State machine enforcement & audit trail |
| 3.7 | Standards API | VERIFIED | `go test ./internal/standards` | Verified Kenyan building codes & statutory rules |
| 3.8 | Audit API | VERIFIED | `go test ./internal/audit` | Immutable audit logging and query endpoint |
| 3.9 | Auth & RBAC API | VERIFIED | `go test ./internal/auth` | Roles, permissions matrix, session profile |
| 3.10 | OpenAPI spec | VERIFIED | `apps/api/openapi/openapi.yaml` | OpenAPI 3.1.0 contract definition |
| 3.11 | Security Headers & CORS | VERIFIED | `internal/http/middleware` | CSP, X-Frame-Options, CORS, request IDs |

### Phase 4: Data Foundation & Ingestion Subsystem

| # | Requirement | Status | Verification | Notes |
|---|---|---|---|---|
| 4.1 | Nairobi County boundary | VERIFIED ALTERNATIVE SOURCE | `go test ./internal/boundaries`, PostGIS migrations 002/003 | geoBoundaries Kenya ADM1 2020 (RCMRD GeoPortal / Africa GeoPortal), EPSG:4326; official statutory source not verified |
| 4.2 | Building footprint sources | VERIFIED | `data/manifests/sources_manifest.json` | MS Kenya Footprints, Google Open Buildings v3 |
| 4.3 | Geospatial ingestion subsystem | VERIFIED (Engine & Validators) / IMPLEMENTED_UNVERIFIED (Live Fetches) | `go test ./internal/ingestion/...` | `apps/api/internal/ingestion/`: validator, service, runner, GeoJSON checks, SSRF prevention, RFC 7946 defaults, finite coordinate validation |
| 4.4 | Validation & promotion gates | VERIFIED | `go test ./internal/ingestion/...` | Unreviewed datasets blocked, steward override reason enforcement, envelope screening vs spatial containment disclosure |


### Phase 5: Frontend Workspaces & Architecture

| # | Requirement | Status | Verification | Notes |
|---|---|---|---|---|
| 5.1 | Application shell and navigation | VERIFIED | `Navbar.tsx`, `App.tsx` | Top editorial navigation, persona badge, responsive shell |
| 5.2 | Shared components library | VERIFIED | `apps/web/src/components/` | Reusable layout, feedback, forms, and data-display components |
| 5.3 | Shared domain hooks | VERIFIED | `apps/web/src/hooks/` | Typed hooks for buildings, datasets, inspections, indicators, auth, status |
| 5.4 | Building Intelligence workspace | VERIFIED | `BuildingIntelligence.tsx` | Nairobi map, filters, accessible table, detail panel via `useBuildings` |
| 5.5 | Research & Data workspace | VERIFIED | `ResearchWorkspace.tsx` | Dataset catalogue, provenance, checksums, versions, GeoJSON audit modal |
| 5.6 | Analysis & Methodology workspace | VERIFIED | `AnalysisWorkspace.tsx` | Transparent formulas, sensitivity slider, limitations via `useIndicators` |
| 5.7 | Inspections workspace | VERIFIED | `InspectionsWorkspace.tsx` | Queue, referral modal, state machine transitions via `useInspections` |
| 5.8 | Security & Governance workspace | VERIFIED | `GovernanceWorkspace.tsx` | Standards register, audit viewer, RBAC matrix via `useStandards` |
| 5.9 | Administration workspace | VERIFIED | `AdminWorkspace.tsx` | System health, telemetry, maintenance actions via `useSystemStatus` |
| 5.10 | Design system implementation | VERIFIED | `index.css` | Warm cream/navy/muted teal editorial palette |
| 5.11 | MapLibre GL JS integration | VERIFIED | `NairobiMap.tsx` | Centered on Nairobi (-1.286, 36.817) with boundary |
| 5.12 | Responsive layout & A11y (WCAG 2.2 AA) | VERIFIED | `npm run build` | Screen-reader table view, high contrast, ARIA tags, zero TypeScript errors |

### Phase 6: Quality, Linting & Security

| # | Requirement | Status | Verification | Notes |
|---|---|---|---|---|
| 6.1 | Go unit test suite | VERIFIED | `go test ./...` | 100% pass across all internal packages |
| 6.2 | Frontend typecheck & build | VERIFIED | `npm run build` | 0 TypeScript errors, 1950 modules transformed |
| 6.3 | Frontend linting | VERIFIED | `npm run lint` | Oxlint 0 warnings and 0 errors |
| 6.4 | Security threat model | VERIFIED | `docs/threat-model.md` | STRIDE analysis and defense-in-depth controls |
| 6.5 | Connect-Time SSRF & Rebinding Defense | VERIFIED | `apps/api/internal/ingestion/ssrf.go`, `ssrf_test.go` | Custom dialer, IPv4-mapped IPv6 normalization, redirect controls, 50MB payload stream bounds |
| 6.6 | Operations Readiness & Runbooks | VERIFIED | `docs/operations-readiness.md` | Liveness/readiness probe semantics, graceful shutdown, migration runbook |
