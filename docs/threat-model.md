# Mjengofiti — Threat Model & Security Architecture
**Reference:** OWASP Application Security Verification Standard (ASVS) 5.0.0

## 1. System Asset Classification
- **Tier 1 (High Sensitivity):**
  - Building inspection findings and non-compliance violation notices.
  - User authentication credentials, password hashes, and session tokens.
  - Case assignment details and inspector identities.
- **Tier 2 (Operational Data):**
  - Building intelligence attributes, external IDs, calculated indicator scores.
  - Ingestion job logs, system health telemetry.
- **Tier 3 (Public Metadata):**
  - Open dataset catalogues, standards register clauses, county boundary polygons.

---

## 2. Threat Analysis (STRIDE)

| Threat | Vulnerability / Attack Surface | Mitigations Implemented in Mjengofiti |
|---|---|---|
| **Spoofing** | Unauthenticated user masquerading as an authorized building inspector | JWT verification with standard claims (iss, sub, exp, role); RBAC middleware. |
| **Tampering** | Altering inspection findings or modifying indicator calculations | Immutable case history tables (`case_history`); audit event logging (`audit_events`); database foreign key constraints. |
| **Repudiation** | Inspector denies submitting a controversial structural finding | Every finding records `author_id`, `created_at`, and generates a tamper-evident audit record. |
| **Information Disclosure** | Exposing sensitive structural defect reports or database credentials | Strict CORS headers; parameterized SQL (zero query string concatenation); non-sensitive health check responses. |
| **Denial of Service** | Unbounded spatial queries requesting millions of polygons across Nairobi | Enforced viewport bounding box limits (`MaxBBoxLimit = 5000`); query pagination (`limit`/`offset`); rate limiting middleware. |
| **Elevation of Privilege** | Read-only user modifying dataset promotion or case status | Server-side role validation in Go handlers before any write transaction; strict route-level permission checks. |

---

## 3. Defense-in-Depth Security Controls
1. **Parameterized Queries:** All SQL calls utilize `pgx` with parameterized arguments `$1, $2, ...` to prevent SQL injection.
2. **Security Headers:** HTTP responses include:
   - `Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:;`
   - `X-Content-Type-Options: nosniff`
   - `X-Frame-Options: DENY`
   - `Referrer-Policy: strict-origin-when-cross-origin`
3. **CORS Restrictions:** Configurable origin whitelist (defaults to local development origins `http://localhost:5173`, `http://localhost:8080`).
4. **Input Validation:** Coordinates, bounding boxes, pagination offsets, and case priorities validated prior to database execution.
