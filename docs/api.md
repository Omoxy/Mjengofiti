# Mjengofiti — REST API Specification

## Overview
- **Base URL:** `/api/v1`
- **Protocol:** HTTP/1.1 or HTTP/2 over TLS
- **Data Formats:** JSON (`application/json`), GeoJSON (`application/geo+json`)
- **Authentication:** Bearer token (JWT) via `Authorization: Bearer <token>`

---

## Endpoints Summary

### 1. Health & System
| Method | Path | Description | Auth Required |
|---|---|---|---|
| `GET` | `/health/live` | Liveness probe (returns 200 OK) | No |
| `GET` | `/health/ready` | Readiness probe (verifies DB connection) | No |
| `GET` | `/api/v1/system/status` | Detailed system and dependency status | No |
| `GET` | `/api/v1/config/map` | MapLibre basemap, center, and zoom config | No |

### 2. Spatial & Boundaries
| Method | Path | Description | Auth Required |
|---|---|---|---|
| `GET` | `/api/v1/boundaries` | Nairobi County administrative boundary GeoJSON | No |
| `GET` | `/api/v1/coverage` | Building count and spatial dataset coverage | No |

### 3. Buildings
| Method | Path | Description | Auth Required |
|---|---|---|---|
| `GET` | `/api/v1/buildings` | Query buildings by bounding box (`bbox=minLng,minLat,maxLng,maxLat`) or text (`q=...`) | Optional |
| `GET` | `/api/v1/buildings/{id}` | Retrieve building profile with sourced facts, observations, and indicators | Optional |
| `GET` | `/api/v1/buildings/{id}/indicators` | Retrieve analytical indicators for a specific building | Optional |
| `GET` | `/api/v1/buildings/{id}/evidence` | Sourced observations, sensor records, and footprint provenance | Optional |
| `GET` | `/api/v1/buildings/{id}/history` | Historical observations and structural records | Optional |

### 4. Datasets & Ingestion
| Method | Path | Description | Auth Required |
|---|---|---|---|
| `GET` | `/api/v1/datasets` | List registered spatial and environmental datasets | Optional |
| `POST` | `/api/v1/datasets` | Register a new dataset entry | Yes (Data Steward, Admin) |
| `GET` | `/api/v1/datasets/{id}` | Retrieve dataset details, versions, and validation history | Optional |
| `POST` | `/api/v1/ingestion-jobs` | Trigger an ingestion job for a dataset version | Yes (Data Steward, Admin) |
| `GET` | `/api/v1/jobs/{id}` | Retrieve ingestion job progress, logs, and record metrics | Yes |
| `POST` | `/api/v1/datasets/{id}/promotions` | Promote validated dataset version to operational status | Yes (Admin, Data Steward) |

### 5. Analysis & Indicators
| Method | Path | Description | Auth Required |
|---|---|---|---|
| `GET` | `/api/v1/indicators` | List all codified screening indicator definitions | Optional |
| `GET` | `/api/v1/indicators/{id}` | Retrieve calculation formula, input schemas, and limitations | Optional |
| `POST` | `/api/v1/analysis-jobs` | Queue calculation job for indicator updates | Yes (Data Researcher, Admin) |
| `GET` | `/api/v1/analysis-jobs/{id}` | Inspect analysis job progress and results | Yes |

### 6. Inspections
| Method | Path | Description | Auth Required |
|---|---|---|---|
| `GET` | `/api/v1/inspection-cases` | List inspection referral queue with filter by status/priority | Yes (Inspector, Supervisor) |
| `POST` | `/api/v1/inspection-cases` | Create a new inspection referral case | Yes (Supervisor, Inspector) |
| `GET` | `/api/v1/inspection-cases/{id}` | Retrieve case details, assigned officer, evidence, and findings | Yes |
| `PATCH` | `/api/v1/inspection-cases/{id}` | Update case status (enforcing state machine constraints) | Yes |
| `POST` | `/api/v1/inspection-cases/{id}/assignments` | Assign case to an authorized building inspector | Yes (Supervisor) |
| `POST` | `/api/v1/inspection-cases/{id}/findings` | Author and submit professional structural findings | Yes (Inspector) |
| `GET` | `/api/v1/inspection-cases/{id}/history` | Immutable audit history of case transitions | Yes |

### 7. Governance & Standards
| Method | Path | Description | Auth Required |
|---|---|---|---|
| `GET` | `/api/v1/standards` | List verified statutory clauses and code references | Optional |
| `GET` | `/api/v1/audit-events` | Query security and operational audit trail | Yes (Admin) |
