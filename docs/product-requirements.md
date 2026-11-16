# Mjengofiti — Product Requirements Document (PRD)

## 1. Product Overview
**Mjengofiti** is a production-grade building intelligence, geospatial research, and inspection prioritisation platform designed specifically for Nairobi County, Kenya.

Its primary purpose is to enhance evidence-based building inspection planning through authoritative geospatial data, traceable indicators, and rigorous inspection workflows.

### 1.1 Core Problems Addressed
1. **Unplanned & Non-compliant Developments:** Nairobi County faces rapid urban densification with fragmented tracking of building structures, heights, and structural compliance.
2. **Opaque Inspection Triage:** Building inspections by county and national authorities (e.g., National Construction Authority - NCA, Nairobi City County Government) often lack centralised, explainable screening evidence.
3. **Data Quality & Provenance Deficits:** Available geospatial datasets (OpenStreetMap, Google Open Buildings, Microsoft Footprints) vary significantly in accuracy, currency, and coverage.
4. **Lack of Explainability:** Prioritisation algorithms often act as "black boxes" or falsely claim to predict building collapse without geotechnical or structural engineering ground truth.

### 1.2 Non-Negotiable Product Principles
- **Observed vs Derived Separation:** Observed facts (e.g., satellite-derived footprint), derived measurements (e.g., polygon area), modelled indicators (e.g., density screening score), and professional findings (e.g., engineer's on-site report) are strictly partitioned in the data model and UI.
- **No Fictional Collapse Prediction:** The system is explicitly a screening and prioritisation tool. It never claims to predict building collapse or certify structural integrity without qualified physical engineering inspection.
- **Truthful Data State:** Missing data is presented honestly. Empty states and degraded offline states reflect reality rather than displaying synthetic or fabricated records.
- **Auditability & Traceability:** Every indicator must trace back to input dataset versions, calculation methods, and timestamps. Every state transition in an inspection case is immutably logged.

---

## 2. User Roles and Personas

| Role | Description | Primary Permissions |
|---|---|---|
| **Read-only Viewer** | Public or cross-departmental observer | View building intelligence map, view dataset metadata, view standards register |
| **Data Researcher** | GIS analyst or academic researcher | View and query datasets, inspect indicators, export aggregated statistics |
| **Data Steward** | GIS / Data engineer managing catalogues | Register datasets, upload versions, run validation reports, manage promotions |
| **Inspector** | Qualified building inspector (NCA / County) | View assigned cases, record inspection findings, upload evidence references |
| **Inspection Supervisor** | Senior engineer / Directorate head | Triage referrals, assign cases, approve findings, close or reopen cases |
| **Administrator** | Systems and security engineer | User management, RBAC configuration, audit logs, system maintenance |

---

## 3. Functional Workspaces

### 3.1 Workspace A: Building Intelligence
- **Nairobi County Map:** MapLibre GL JS client rendering Nairobi extent (`[-1.286389, 36.817223]`, bounds `[36.65, -1.45, 37.10, -1.15]`).
- **Layer Controls:** Administrative boundary, building footprints, density heatmaps, terrain contours/drainage.
- **Spatial & Attribute Search:** Search by unique building ID, external source ID, sub-county/neighbourhood (e.g., Kilimani, Westlands, Eastleigh, Embakasi), and area.
- **Building Profile & Evidence Panel:**
  - Sourced geometry footprint with source dataset provenance.
  - Calculated polygon area (m²).
  - Recorded height (m) and measurement method (e.g., shadow analysis, LiDAR, reported).
  - Nearby building density context.
  - Linked inspection cases and verified incident history (if any).
  - Explicit honest indicators: "Attribute not recorded" for unmeasured parameters.
- **Accessible Table/List View:** Alternative non-map tabular search interface meeting WCAG 2.2 AA.

### 3.2 Workspace B: Research & Data
- **Dataset Catalogue:** Comprehensive register of spatial and tabular datasets covering Nairobi County.
- **Metadata Register:** Publisher, official source URL, licence, acquisition date, CRS (EPSG:4326 / EPSG:3857), checksum, and coverage percentage.
- **Ingestion & Validation Pipeline:**
  - Ingestion job tracking (pending, running, completed, failed) with error logs.
  - Validation reports: Geometry validity (ST_IsValid), duplicate detection, missing attribute counts, spatial boundary compliance.
  - Version history and promotion controls (draft -> validated -> promoted -> retired).

### 3.3 Workspace C: Analysis & Methodology
- **Indicator Definitions:** Transparent documentation of every screening indicator:
  - Built-up density index (buildings per hectare).
  - Height-to-footprint ratio screening.
  - Drainage/flood susceptibility context (slope and elevation derived).
  - Data completeness index.
- **Calculation Provenance:** Input datasets, version tags, formula, calculation timestamp, missing input handling.
- **Sensitivity & Limitation Analysis:** Clear display of method limitations, spatial resolution boundaries, and false-positive risks.

### 3.4 Workspace D: Inspections
- **Referral Queue:** Triage list of buildings flagged for inspection review with priority levels (Low, Medium, High, Critical).
- **Case Lifecycle State Machine:**
  - States: `open` → `triaged` → `assigned` → `in_progress` → `pending_review` → `closed` (or `reopened`).
- **Case Details:** Building context, referral rationale, attached evidence references, assigned officer, target due date.
- **Inspection Findings Entry:** Professional findings record with structural observations, compliance status against standards, recommended corrective actions.
- **Case Audit Trail:** Chronological log of who modified the case, previous and new status, and action timestamps.

### 3.5 Workspace E: Security & Governance
- **Audit Trail Viewer:** Searchable audit events with actor, action, resource type, resource ID, IP address, and timestamp.
- **Standards Register:** Kenyan building codes and regulatory statutory instruments:
  - National Building Code 2024.
  - National Construction Authority (NCA) Act.
  - Physical and Land Use Planning Act 2019.
  - Section/clause references, requirements, and software enforcement mapping.
- **Access Control Matrix:** Role-Based Access Control (RBAC) permission inspector.

### 3.6 Workspace F: Administration
- **System Health:** Liveness, readiness, PostgreSQL/PostGIS connection status, memory utilization, API uptime.
- **Data Source Connectors:** Configuration status for external data feeds (Copernicus, OSM, Google, Microsoft).
- **Operational Controls:** Maintenance triggers, cache clearing, log level controls.

---

## 4. Non-Functional Requirements
1. **Performance:** Spatial queries bounded to viewport bounding boxes; API response times under 200ms for p95 bounded bbox queries.
2. **Security:** OWASP ASVS 5.0.0 alignment; parameterised SQL queries (pgx); JWT/Bearer token authentication with RBAC enforcement; CSRF and security headers (CSP, HSTS, X-Content-Type-Options).
3. **Accessibility:** WCAG 2.2 AA compliance; keyboard navigation, ARIA live regions for map updates, high-contrast text ratios.
4. **Maintainability:** Modular monolith Go backend; strict TypeScript React frontend; version-controlled reversible SQL migrations.
