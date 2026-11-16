# MjengoFiti

MjengoFiti is a geospatial building-intelligence and inspection-prioritisation application for Nairobi County, Kenya. It combines a map-based React client, a Go REST API, and PostgreSQL/PostGIS-backed spatial data.

## Capabilities

- Explore building records and spatial layers on an interactive map.
- Browse dataset provenance, screening indicators, standards, and inspection workflows.
- Query building and boundary data through the Go API.
- Run database migrations and automated frontend and backend checks.

## Current Data Limitations

- The Nairobi county outline uses the geoBoundaries Kenya ADM1 2020 Nairobi feature (`geoBoundaries-KEN-ADM1`, `KE-30`), distributed by geoBoundaries from the RCMRD GeoPortal/Africa GeoPortal. It is an alternative, not a verified official statutory government boundary. See [boundary source metadata](https://www.geoboundaries.org/api/current/gbOpen/KEN/ADM1/).
- Sub-county geometry is not currently loaded.
- The eight building records in the seed migration are synthetic, labelled `unverified_sample`, and are for development/UI testing only. Do not use them to make real inspection, safety, enforcement, or investment decisions.
- The boundary seed migration also contains a coarse viewport approximation; it is not an authoritative boundary source.
- Production authentication, security configuration, and deployment validation remain outstanding. This repository is not production-ready.

## Architecture and Technology

- **Web:** React 19, TypeScript 6, Vite 8, MapLibre GL JS, TanStack Query, React Router.
- **API:** Go 1.26.1, Chi, pgx, REST/JSON and GeoJSON endpoints.
- **Database:** PostgreSQL with PostGIS; SQL migrations are managed by the API migration command.
- **Maps:** MapLibre renders vector styles from MapTiler. The browser key is injected through environment configuration and is visible to browsers; restrict it to approved website origins in MapTiler Cloud.
- **Containers:** Docker Compose, with a development database service and a separate full-stack Compose file.

## Repository Layout

```text
apps/api/                 Go API, migrations, tests, and OpenAPI contract
apps/web/                 React client, tests, and static assets
data/manifests/           Dataset source and provenance registry
docs/                     Product, architecture, data, security, and operations docs
docker-compose.yml        Local PostgreSQL/PostGIS service
docker-compose.prod.yml   Container build/run example; not production-ready
Makefile                  Common development commands
```

## Prerequisites

- Node.js 20.19+ (or a supported newer LTS) and npm.
- Go 1.26.1, as specified by `apps/api/go.mod`.
- PostgreSQL with PostGIS 3.x, or Docker with the Compose plugin.
- A MapTiler key for map tiles. Use a browser key restricted to the development and production website origins; do not commit it.

## Local Setup

### Configure environment

From the repository root, copy `.env.example` to `.env` and replace the database password and JWT placeholder with locally generated values. The root `.env` is ignored by Git.

For the web client, copy `apps/web/.env.example` to `apps/web/.env.local` and set `VITE_MAPTILER_KEY` to your MapTiler browser key. The `.env.local` file is ignored. MapTiler keys are delivered to the browser, so restrict permitted website origins in MapTiler Cloud; do not treat a browser key as a server-side secret.

### Start PostgreSQL/PostGIS with Docker

After configuring `.env`, start the database service:

```sh
docker compose up -d db
```

The service uses the named volume `mjengofiti_pgdata`, published on the configured local database port (5432 by default). Docker Compose requires `DB_PASSWORD`; it has no built-in password fallback. Stopping or rebuilding containers does not remove this named volume.

Alternatively, create a local PostgreSQL database with PostGIS enabled and configure `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, and `DB_PASSWORD` in `.env`.

### Apply database migrations

With PostgreSQL/PostGIS running:

```sh
cd apps/api
go run ./cmd/migrate up
```

The migration command loads the repository-root `.env` when run from `apps/api`. `make migrate` is also available where GNU Make is installed.

### Start the API and web client

In one terminal, from the repository root:

```sh
cd apps/api
go run ./cmd/api
```

In a second terminal:

```sh
cd apps/web
npm ci
npm run dev
```

Local URLs:

- Buildings application: <http://localhost:5173/buildings>
- API: <http://localhost:8080>
- Liveness: <http://localhost:8080/health/live>
- Readiness: <http://localhost:8080/health/ready>

The Vite development server proxies `/api` and `/health` to the local API.

### Docker application images

`docker-compose.prod.yml` builds and runs the API, web client, and database. It requires `DB_PASSWORD` and `VITE_MAPTILER_KEY` in the environment. It is an example container configuration, not a production deployment recipe: production identity/authentication, TLS, secret management, network policy, backups, and operational controls still need to be designed and verified.

## Checks

Run the web commands from `apps/web`:

```sh
npm run build       # Includes TypeScript project checks
npx tsc -b          # TypeScript checks only
npm test            # Frontend unit tests
npm run lint        # Frontend lint
```

Run the API commands from `apps/api`:

```sh
go test ./...
go vet ./...
```

The migration and integration tests require a reachable PostgreSQL/PostGIS database where noted by the test. `make test` and `make lint` run the corresponding API and web checks when Make is available.

## Data Provenance

`data/manifests/sources_manifest.json` registers candidate external datasets and their stated publishers, licenses, and source URLs. Full third-party datasets are not bundled here. The initial spatial seed is explicitly development data; its eight building records are synthetic samples, not verified building inventories. Review the [data governance](docs/data-governance.md) and [data feasibility](docs/data-feasibility.md) documents before importing or operationally using any source.

## Further Documentation

- [Architecture](docs/architecture.md)
- [API contract](docs/api.md)
- [Data feasibility and limitations](docs/data-feasibility.md)
- [Data governance](docs/data-governance.md)
- [Threat model](docs/threat-model.md)
- [Operations](docs/operations/)
