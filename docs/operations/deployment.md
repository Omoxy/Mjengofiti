# Mjengofiti Deployment Runbook

## 1. Operational Overview
Mjengofiti operates as an integrated spatial intelligence platform with three primary tiers:
1. Spatial Data Tier: PostgreSQL 16 with PostGIS 3.4 spatial engine.
2. Application API Tier: Go compiled service providing geospatial filtering, standards evaluation, and case management.
3. Web Client Tier: React 19 application utilizing MapLibre GL for client rendering.

## 2. Local Development Environment
Prerequisites:
- Node.js 20 or higher
- Go 1.22 or higher
- Docker Desktop with active Linux engine

### 2.1 Database Initialization
Launch the PostGIS database container:
```bash
docker compose up -d db
```

Apply database migrations:
```bash
cd apps/api
go run ./cmd/migrate up
```

Seed authentic Nairobi spatial baseline data:
```bash
docker exec -i mjengofiti-db psql -U mjengofiti -d mjengofiti < apps/api/migrations/seed_nairobi_spatial.sql
```

### 2.2 Running Application Services
Start the Go backend server:
```bash
cd apps/api
go run ./cmd/api
```
The API listens on port 8080 and exposes `/health/live`, `/health/ready`, and `/api/v1` routes.

Start the React development server:
```bash
cd apps/web
npm run dev
```
The web client starts at `http://localhost:5173` with automated proxy routing to the API on port 8080.

## 3. Production Deployment with Docker Compose
For containerized staging or single node production hosts:

1. Configure environment variables in `.env`:
   - Set strong `DB_PASSWORD`
   - Configure `VITE_MAPTILER_KEY` with your production MapTiler key
   - Confirm `MAP_TILE_URL`

2. Launch all services:
```bash
docker compose -f docker-compose.prod.yml up -d --build
```

3. Verification Probes:
```bash
curl -f http://localhost/health/ready
curl -f http://localhost/api/v1/coverage
```

## 4. Rollback and Disaster Recovery
- Standalone migrations support reversible rollbacks: `go run ./cmd/migrate down -steps 1`.
- PostGIS database volumes reside in the named volume `mjengofiti_pgdata`.
- Automated backups should target `pg_dump -Fc -U mjengofiti -d mjengofiti`.
