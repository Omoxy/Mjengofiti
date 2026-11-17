# MJENGOFITI: STAGING DEPLOYMENT PLAN AND HOSTING ARCHITECTURE

* Document Identifier: DEPLOY-STAGING-2026-V1
* Location: `docs/deployment/STAGING_DEPLOYMENT_PLAN.md`
* Target Application: MjengoFiti Geospatial Platform
* Target Branch: `master`
* Primary Goal: Establish a reliable, isolated, and cost effective staging environment prior to public production release

---

## 1. PROPOSED HOSTING ARCHITECTURE

MjengoFiti comprises three distinct application tiers that must be hosted with clear boundary separation:
1. Geospatial Database: PostgreSQL 16 with PostGIS 3.4 extension, GiST spatial indexing, and persistent volume storage.
2. Backend API Service: Compiled Go 1.26 REST microservice running in an unprivileged container on port 8080.
3. Web Frontend Client: React 19 single page application served via Nginx or a global edge CDN with HTTPS.

### Recommended Staging Architecture: Managed Cloud Containers (PaaS Option)

To minimize operational overhead while guaranteeing PostGIS compatibility and certificate verified TLS:
* Database Tier: Managed PostgreSQL with PostGIS enabled (DigitalOcean Managed Databases or Render Managed PostgreSQL).
* Backend Tier: Containerized Go API running on a managed container service with automated health check restarts.
* Frontend Tier: Static web asset hosting with global edge caching (Cloudflare Pages, Vercel, or DigitalOcean App Platform).
* Ingress and Routing: Custom subdomains `staging.mjengofiti.org` (frontend) and `api-staging.mjengofiti.org` (backend API) with automated TLS certificates.

### Alternative Staging Architecture: Dedicated Linux Host (Single VPS Option)

For development teams seeking single host simplicity matching `docker-compose.prod.yml`:
* Compute: Dedicated virtual private server (2 vCPU, 4 GB RAM, 50 GB SSD).
* Runtime: Docker Engine with Docker Compose running `mjengofiti-db`, `mjengofiti-api`, and `mjengofiti-web`.
* Reverse Proxy: Caddy or Nginx reverse proxy managing automatic Let's Encrypt TLS certificates and routing external traffic.

---

## 2. SERVICE RESPONSIBILITIES

```
+=============================================================================+
|                      STAGING SERVICE RESPONSIBILITY MATRIX                  |
+=============================================================================+
| Tier          | Service            | Role and Responsibilities              |
+=============================================================================+
| Web Client    | mjengofiti-web     | Serves static HTML/JS/CSS assets       |
|               |                    | Injects public MapTiler browser key    |
|               |                    | Handles client routing and map UI      |
+---------------+--------------------+----------------------------------------+
| Backend API   | mjengofiti-api     | Enforces JWT authentication and RBAC   |
|               |                    | Executes parameterized PostGIS queries |
|               |                    | Validates GeoJSON and enforces SSRF    |
|               |                    | Implements token bucket rate limiting  |
+---------------+--------------------+----------------------------------------+
| Database      | mjengofiti-db      | Stores building and inspection records |
|               |                    | Executes GiST indexed spatial searches |
|               |                    | Maintains schema_migrations ledger     |
+=============================================================================+
```

---

## 3. REQUIRED STAGING ENVIRONMENT VARIABLES

All production credentials must be injected dynamically via environment variables and never committed to source control.

### Backend API Variables
* `APP_ENV`: Must be set to `staging`
* `APP_PORT`: `8080`
* `APP_LOG_LEVEL`: `info`
* `APP_CORS_ORIGINS`: Explicit HTTPS URL of the staging web app, e.g. `https://staging.mjengofiti.org`
* `DB_HOST`: Private hostname or IP of the staging PostgreSQL instance
* `DB_PORT`: `5432`
* `DB_NAME`: `mjengofiti_staging`
* `DB_USER`: `mjengofiti_app` (least privilege non superuser role)
* `DB_PASSWORD`: Cryptographically random 32 character password
* `DB_SSLMODE`: Set to `verify-full` (cloud managed DB) or `require`
* `AUTH_ENABLED`: `true` (enforced fail closed)
* `AUTH_ISSUER_URL`: `https://auth-staging.mjengofiti.org`
* `AUTH_AUDIENCE`: `mjengofiti-staging-api`
* `JWT_SECRET`: Cryptographically generated 64 byte random string
* `RATE_LIMIT_REQUESTS_PER_MINUTE`: `120`
* `RATE_LIMIT_BURST`: `20`

### Frontend Web Variables
* `VITE_API_URL`: `https://api-staging.mjengofiti.org`
* `VITE_MAPTILER_KEY`: Public MapTiler browser key restricted in MapTiler Cloud to `https://staging.mjengofiti.org/*`

---

## 4. DATABASE INITIALIZATION AND MIGRATION PROCEDURE

1. Provision Database Instance:
   Ensure PostgreSQL 16 is installed and accessible. Enable PostGIS before running migrations:
   ```sql
   CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
   CREATE EXTENSION IF NOT EXISTS postgis;
   ```

2. Provision Application User Role:
   Create a dedicated application user with restricted privileges (avoid using database superuser accounts in staging):
   ```sql
   CREATE USER mjengofiti_app WITH PASSWORD 'GENERATED_PASSWORD_HERE';
   GRANT CONNECT ON DATABASE mjengofiti_staging TO mjengofiti_app;
   GRANT USAGE ON SCHEMA public TO mjengofiti_app;
   GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO mjengofiti_app;
   GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO mjengofiti_app;
   ```

3. Execute Schema Migrations:
   Run migrations using the compiled migration tool before starting the API:
   ```bash
   cd apps/api
   go run ./cmd/migrate up
   ```
   Verify that `schema_migrations` contains all 3 versions:
   * `1_initial_schema`
   * `2_nairobi_county_boundary`
   * `3_nairobi_boundary_provenance`

---

## 5. TLS AND REVERSE PROXY CONFIGURATION

1. Modern Transport Security:
   * Enforce TLS 1.3 as default, allowing TLS 1.2 minimum.
   * Enable HTTP Strict Transport Security (HSTS) with `max-age=31536000; includeSubDomains`.
2. Backend API Proxy Rules:
   * Proxy pass `https://api-staging.mjengofiti.org/*` to `http://127.0.0.1:8080/*`.
   * Forward headers: `X-Forwarded-For`, `X-Forwarded-Proto`, `X-Request-ID`.
   * Enforce client request body limit of 50 MB for GeoJSON uploads.
3. Frontend Proxy Rules:
   * Direct all unmatched requests to `index.html` to support HTML5 browser routing.
   * Cache static assets (`/assets/*`) with `Cache-Control: public, max-age=31536000, immutable`.
   * Prevent caching of `index.html` via `Cache-Control: no-cache, no-store, must-revalidate`.

---

## 6. BACKUP AND DISASTER RECOVERY STRATEGY

1. Automated Daily Snapshots:
   * Database backups must run daily via `pg_dump` targeting object storage (e.g. AWS S3 or Cloudflare R2).
   * Command:
     ```bash
     pg_dump -Fc -U mjengofiti_backup -d mjengofiti_staging > /backups/mjengofiti_staging_$(date +%F).dump
     ```
2. Retention Schedule:
   * Retain daily dumps for 14 days.
   * Retain weekly dumps for 8 weeks.
3. Restoration Drill:
   * Test restoration onto a clean database instance monthly:
     ```bash
     pg_restore -U mjengofiti_admin -d mjengofiti_test_restore -v /backups/target.dump
     ```
   * Execute PostGIS geometry checks (`ST_IsValid`, `ST_Centroid`) on restored records to verify spatial data integrity.

---

## 7. STEP BY STEP DEPLOYMENT SEQUENCE

1. Pre-Deployment Gate:
   * Verify all GitHub Actions CI checks have passed on `master`.
   * Confirm no open CVEs exist in dependencies.
2. Infrastructure Provisioning:
   * Provision PostgreSQL service with PostGIS enabled.
   * Verify network connectivity and TLS certificate validity.
3. Database Migration:
   * Run `migrate up` against the staging database.
   * Confirm table creation and spatial reference systems (`srid = 4326`).
4. Backend API Deployment:
   * Build container image using `apps/api/Dockerfile`.
   * Inject staging environment variables.
   * Start container service and probe `https://api-staging.mjengofiti.org/health/live`.
5. Frontend Client Deployment:
   * Build web distribution using `npm run build` with `VITE_API_URL` and `VITE_MAPTILER_KEY` injected.
   * Deploy static bundle to web hosting or Nginx container.
6. Post-Deployment Smoke Test:
   * Query `/health/ready` to verify database connection pool status.
   * Load `https://staging.mjengofiti.org/buildings` to verify MapTiler tile loading and building footprints.

---

## 8. ROLLBACK PROCEDURE

If staging verification fails or a critical error occurs:
1. Backend Service Rollback:
   * Re-deploy the previously known good container image tag.
2. Frontend Client Rollback:
   * Revert the static hosting release to the previous commit artifact.
3. Database Rollback and Data Protection:
   * Prefer verified backup restoration and forward migrations over destructive down migrations where data may be lost.
   * If database corruption or migration failure occurs, restore the database from the pre-deployment snapshot dump using `pg_restore`.
   * For schema defects identified post deployment, author and deploy a forward corrective migration rather than executing destructive schema drops.
4. Incident Documentation:
   * Log failure details, error messages, and remediation actions in `docs/operations/incident-response.md`.

---

## 9. ESTIMATED OPERATIONAL COSTS (STAGING ENVIRONMENT)

Based on standard public cloud pricing without long term commitments:

* Managed PostgreSQL with PostGIS (DigitalOcean or Render): approximately \$15 to \$25 per month.
* Containerized Go Backend API (DigitalOcean Basic App or Render Web Service): approximately \$7 to \$12 per month.
* Static Frontend Web Hosting (Cloudflare Pages or Vercel Hobby): \$0 (free tier suitable for staging).
* Domain and DNS Management: approximately \$10 to \$15 per year.
* MapTiler Cloud Tiles (Free / Flex tier up to 100,000 requests per month): \$0 for staging volume.
* Total Estimated Monthly Staging Cost: approximately \$22 to \$37 per month.
