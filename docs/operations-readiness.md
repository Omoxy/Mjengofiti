# Mjengofiti: Operations Readiness and Deployment Runbook

## 1. Operational Overview

Mjengofiti is architected as a modular monolith providing high performance, operational simplicity, and straightforward deployment.

- **Backend Runtime:** Go compiled binary (Go 1.22+), stateless HTTP server utilizing Chi v5 router.
- **Frontend Runtime:** React 19 single-page application compiled to static HTML, CSS, and JavaScript via Vite, served through a CDN or Nginx container.
- **Spatial Database:** PostgreSQL 15+ with PostGIS 3+ extension.
- **Service Communications:** Clean REST API over JSON with standard RFC 7807 problem details.

---

## 2. Health and Readiness Probe Semantics

Mjengofiti provides distinct endpoints for container orchestrators (Kubernetes, AWS ECS, or Docker Compose):

### 2.1 Liveness Probe (`GET /health/live`)
- **Purpose:** Indicates whether the Go HTTP server process is running and able to handle socket traffic.
- **Response:** HTTP 200 OK with `{"status":"ok","timestamp":"..."}`.
- **Action on Failure:** Orchestrator must restart the container instance.

### 2.2 Readiness Probe (`GET /health/ready`)
- **Purpose:** Indicates whether the application can service user traffic requiring database connectivity.
- **Behavior:**
  - When PostgreSQL/PostGIS connection pool responds to ping: HTTP 200 OK.
  - When database is unreachable: HTTP 503 Service Unavailable with `{"status":"degraded","error":"database_unreachable"}`.
- **Action on Failure:** Traffic router removes instance from active load balancer pool until connection recovery.

### 2.3 Detailed System Telemetry (`GET /api/v1/system/status`)
- **Purpose:** Provides administrative insight into database latency, PostGIS extension version, active connections, and memory statistics.
- **Accessibility:** Protected administrative endpoint for monitoring agents (Prometheus or Datadog).

---

## 3. Graceful Startup and Shutdown

1. **Signal Handling:** The application listens for `SIGINT` (Ctrl+C) and `SIGTERM` (container termination signal).
2. **Grace Period:** Upon receiving a termination signal, the HTTP server allows up to 10 seconds for active requests to complete before closing listeners.
3. **Connection Teardown:** Open database connection pools and background ingestion workers are flushed and closed cleanly.

---

## 4. Database Migrations Runbook

### 4.1 Automated Startup Migrations
Migrations are embedded directly in the API binary via `//go:embed *.sql`. During container initialization, the entry point executes embedded migrations before launching the HTTP server.

### 4.2 Migration CLI Usage
For administrative or CI/CD pipeline operations, the standalone migration CLI is located in `apps/api/cmd/migrate`:

```bash
# Apply pending migrations
./migrate up

# Check current migration version and status
./migrate status

# Roll back the latest migration step
./migrate down -steps 1
```

### 4.3 Disaster Recovery and Rollback Semantics
- Every migration must provide both an `*.up.sql` and a reversible `*.down.sql` script.
- The `schema_migrations` table records applied version IDs and SHA-256 checksums to detect schema divergence.
- Rollbacks must be verified against disposable test environments before execution against production clusters.

---

## 5. Configuration and Environment Variables

| Variable Name | Required | Default Value | Purpose |
|---|---|---|---|
| `PORT` | No | `8080` | HTTP port on which the Go API listens. |
| `ENV` | No | `development` | Deployment environment (`development`, `staging`, `production`). |
| `DATABASE_HOST` | Yes | `localhost` | PostgreSQL host address. |
| `DATABASE_PORT` | No | `5432` | PostgreSQL port number. |
| `DATABASE_NAME` | Yes | `mjengofiti` | Primary database name. |
| `DATABASE_USER` | Yes | `postgres` | Database authentication username. |
| `DATABASE_PASSWORD` | Yes | `""` | Database authentication password (Tier 1 secret). |
| `DATABASE_SSLMODE` | No | `disable` | SSL connection mode (`require` or `verify-full` in production). |
| `CORS_ORIGINS` | No | `http://localhost:5173` | Comma-separated list of permitted web origins. |
| `TILE_URL` | No | `https://tile.openstreetmap.org/{z}/{x}/{y}.png` | Base raster or vector tile URL for MapLibre client. |

---

## 6. Staging and Production Acceptance Gates

Before promoting any build to operational production status, verify:

1. **Database Migration Verification:** Migrations apply cleanly (`up`), roll back cleanly (`down`), and re-apply cleanly (`up`) on an isolated PostgreSQL/PostGIS instance.
2. **Static Code Quality:** Zero warnings in `go vet ./...` and `npm run lint`.
3. **Unit and Integration Test Pass:** 100% passing tests across backend services and frontend components.
4. **Production Build Success:** `tsc -b && vite build` completes with zero type errors.
5. **SSRF and Connect-Time Defense:** All tests in `internal/ingestion/ssrf_test.go` pass.
6. **Provenance Integrity:** Ingested building datasets reference verified sources with valid licenses.
