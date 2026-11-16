# Mjengofiti — Architecture and Technical Design

## 1. Architectural Style: Modular Monolith
Mjengofiti is built as a single, cohesive, high-performance modular monolith with clean internal architectural boundaries. This avoids unnecessary distributed infrastructure, RPC overhead, or microservice operational complexity while maintaining clear domain boundaries.

```
+-------------------------------------------------------------------+
|               React 19 + TypeScript Frontend (apps/web)          |
|  - MapLibre GL JS (Geospatial)      - TanStack React Query        |
|  - 6 Connected Workspaces           - Recharts & Lucide Icons     |
|  - Accessible Editorial UI System   - React Router v7             |
+-------------------------------------------------------------------+
                                  |
                                  | REST / JSON / GeoJSON
                                  v
+-------------------------------------------------------------------+
|                    Go 1.26 Backend (apps/api)                     |
|  - Chi Router & Middleware (CORS, Rate Limiting, Audit, Security) |
|  - Domain Packages:                                               |
|      * buildings      * datasets       * inspections              |
|      * indicators     * standards      * audit                    |
|      * health         * auth / RBAC    * boundaries               |
+-------------------------------------------------------------------+
                                  |
                                  | pgx Connection Pool / SQL
                                  v
+-------------------------------------------------------------------+
|               PostgreSQL 15+ with PostGIS Extension               |
|  - Spatial Indexing (GiST on geometry)                            |
|  - Reversible SQL Migrations                                      |
|  - Strict Check Constraints & Foreign Keys                        |
+-------------------------------------------------------------------+
```

---

## 2. Component Boundaries and Domain Model

### 2.1 Backend Packages (`apps/api/internal/`)
- `config`: Environment configuration loader (`.env` parsing with validation).
- `database`: PostgreSQL connection pool wrapper via `jackc/pgx/v5/pgxpool`.
- `http/middleware`:
  - Request ID injection (`X-Request-ID`).
  - Structured request logging with duration and status code.
  - Security headers (HSTS, X-Frame-Options, CSP, Referrer-Policy).
  - CORS with configurable allowed origins.
  - Panic recovery with structured error logging.
  - Auth & RBAC token verification.
- `health`: Liveness (`/health/live`), readiness (`/health/ready`), and diagnostic status (`/api/v1/system/status`).
- `buildings`: Spatial queries (bounding box, radius, text search), building details, observations, and footprint geometry retrieval.
- `datasets`: Catalogue management, dataset versions, validation reports, ingestion job tracking, and promotion lifecycle.
- `indicators`: Screening indicator definitions, execution engine, uncertainty quantification, and sensitivity records.
- `inspections`: Referral queue, case management, professional findings, assignments, and state transitions.
- `standards`: Kenyan regulatory compliance clauses register, verified statutory references, and rule bindings.
- `audit`: Immutable event logger tracking actor, action, resource, IP address, and payload metadata.
- `boundaries`: Serves the verified geoBoundaries Kenya ADM1 2020 Nairobi feature. It is an alternative, not a statutory official boundary; the legacy coarse record remains preserved separately.

### 2.2 Database Layer (`apps/api/migrations/`)
- Canonical migrations are maintained in `apps/api/migrations/` using standard forward (`.up.sql`) and rollback (`.down.sql`) files.
- Migrations are embedded directly into the Go executable via `//go:embed *.sql` in `apps/api/migrations/migrations.go`, eliminating fragile relative-path dependencies across runtime environments.
- The migration runner CLI lives at `apps/api/cmd/migrate` and is orchestrated via `make migrate`, `make migrate-status`, and `make migrate-down`.
- All spatial columns use `GEOMETRY(GEOMETRY, 4326)` with spatial GiST indexes.
- Strict database constraints enforce data integrity (e.g., `chk_priority`, `chk_status` on inspection cases).

---

## 3. Data Flow Patterns

### 3.1 Map Viewport Bounding Box Query
1. User pans/zooms map in `apps/web`.
2. MapLibre fires `moveend` event with bounding box coordinates `[minLng, minLat, maxLng, maxLat]`.
3. TanStack Query sends `GET /api/v1/buildings?bbox=minLng,minLat,maxLng,maxLat&limit=500`.
4. Go API validates bounds (ensuring coordinates fall within Kenya geographic envelope and query area is bounded).
5. Repository executes parameterised query using PostGIS `ST_Intersects(geom, ST_MakeEnvelope($1, $2, $3, $4, 4326))`.
6. Results returned as JSON list with coordinates, footprint properties, and data quality flags.

### 3.2 Dataset Lifecycle Pipeline
```
[External Source (OSM/Google/MS/NCA)]
                 |
                 v
        [1. Registration]   --> Stored in `datasets` table (status: draft)
                 |
                 v
        [2. Ingestion Job]  --> Logged in `ingestion_jobs` (records parsed/failed)
                 |
                 v
        [3. Validation]     --> `validation_reports` (geometry checks, CRS, bounds)
                 |
                 v
        [4. Promotion]      --> Version promoted to operational status
                 |
                 v
     [Operational Queries]  --> Rendered on live map and indicators
```

### 3.3 Inspection Case State Machine
```
[Referral / Ingestion]
         |
         v
       (open)
         |  [Triage Action]
         v
     (triaged)
         |  [Assignment to Inspector]
         v
     (assigned)
         |  [Site Visit Commenced]
         v
    (in_progress)
         |  [Findings Submitted]
         v
   (pending_review)
      /         \  [Supervisor Decision]
     v           v
  (closed)   (reopened)
```

---

## 4. Security Architecture
- **Stateless Authentication:** Bearer token (JWT) containing user ID, role, and permissions claims.
- **RBAC Matrix:** Middleware verifies user permissions before allowing state-changing operations (`POST`, `PATCH`, `DELETE`).
- **SQL Injection Prevention:** 100% parameterised SQL queries through `pgx`. No dynamic string concatenation.
- **Degraded Mode Resilience:** When the database is offline or unreachable, the Go API continues running in degraded mode, providing honest health and diagnostic responses rather than crashing.
