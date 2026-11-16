# Mjengofiti: Security Review and Vulnerability Assessment

## 1. Executive Summary

This independent security review establishes the vulnerability posture and security controls of the Mjengofiti platform across backend microservices, geospatial ingestion pipelines, database access layers, and web interfaces.

The review prioritizes threat vectors specific to municipal geospatial building platforms, including Server-Side Request Forgery (SSRF), DNS rebinding attacks, privilege escalation, unauthorized dataset promotion, data tampering, and Denial of Service (DoS) through spatial query abuse.

---

## 2. Threat Vector Analysis and Remediations

### 2.1 Ingestion SSRF and DNS Rebinding Defense (P0 Remediation)

- **Vulnerability:** External dataset ingestion endpoints allowing users to specify remote URLs could be exploited to target internal network services, cloud provider instance metadata services (such as AWS and GCP metadata endpoints at 169.254.169.254), or internal database ports via DNS rebinding.
- **Implemented Mitigation:**
  1. **Strict URL Syntax Validation:** `ValidateURLSyntax` enforces explicit `http` or `https` schemes, prohibits embedded URL credentials (`user:pass@`), validates destination port numbers (1 to 65535), and blocks localhost by name.
  2. **Connect-Time IP Validation:** `NewSafeHTTPClient` implements a custom `DialContext` function. Rather than performing a detached pre-lookup, destination IP addresses are resolved and verified immediately during connection establishment.
  3. **Comprehensive Prohibited CIDR Filter:** The dialer cross-checks all resolved IPs against internal network blocks, including loopback (`127.0.0.0/8`, `::1`), RFC 1918 private subnets (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`), link-local and cloud metadata (`169.254.0.0/16`, `fe80::/10`), Carrier-Grade NAT (`100.64.0.0/10`), and multicast ranges (`224.0.0.0/4`, `ff00::/8`).
  4. **IPv4-Mapped IPv6 Normalization:** Any IPv6 address wrapping an IPv4 representation (e.g. `::ffff:127.0.0.1` or `::ffff:169.254.169.254`) is normalized to its IPv4 form before CIDR evaluation.
  5. **DNS Rebinding Elimination:** The dialer binds the outbound TCP socket directly to the specific IP address validated during address resolution, preventing Time-of-Check to Time-of-Use (TOCTOU) address swapping.
  6. **Redirect Verification:** HTTP redirects are constrained to a maximum of 3 hops and re-validated through the URL syntax validator and safe dialer on each transition.
  7. **Streaming Download Bounds:** `ReadLimitedPayload` constrains streaming responses to an explicit limit (50 MB), preventing memory exhaustion and zip-bomb style payload denial of service.

### 2.2 Dataset Promotion Authorization and Integrity (P1 Remediation)

- **Vulnerability:** Unreviewed or corrupted geospatial datasets being elevated to operational status without formal quality audit or steward oversight.
- **Implemented Mitigation:**
  1. `PromoteVersion` in `internal/ingestion/service.go` enforces a mandatory validation report gate. Any dataset version lacking a completed validation report is rejected.
  2. Failed dataset promotions require explicit steward authority (`force: true`) and a mandatory documented reason in `notes`.
  3. Feature insertion into the operational PostGIS `buildings` table executes with parameterized spatial geometry constructors (`ST_SetSRID(ST_GeomFromGeoJSON($1), 4326)`).

### 2.3 SQL Injection and Spatial Query Abuse (P1 Verified)

- **Vulnerability:** Manipulation of spatial bounding box parameters or search strings to inject malicious SQL fragments or exhaust database memory.
- **Implemented Mitigation:**
  1. All database queries in `internal/buildings`, `internal/datasets`, `internal/inspections`, and `internal/standards` use parameterized placeholders (`$1, $2, ...`).
  2. Bounding box coordinates are strictly parsed into float64 types, with bounded query envelopes (`ST_MakeEnvelope`) and hard upper pagination limits (`MaxBBoxLimit = 5000`, `DefaultLimit = 50`).

### 2.4 Cross-Origin Resource Sharing (CORS) and Security Headers (P2 Verified)

- **Vulnerability:** Cross-site data exposure and clickjacking.
- **Implemented Mitigation:**
  1. Middleware sets `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `X-XSS-Protection: 0`, and strict referrer policies.
  2. CORS headers strictly whitelist configured origins (e.g. `http://localhost:5173`) and reject wildcard allowances on authenticated routes.

---

## 3. Verified Security Controls Matrix

| Control Category | Implemented Mechanism | Verification Evidence | Status |
|---|---|---|---|
| Ingestion SSRF Protection | Custom `DialContext` with CIDR and metadata blocking | `ssrf_test.go` (12 test suites passed) | VERIFIED |
| DNS Rebinding Defense | Socket address pinned to verified IP at connect time | `TestSafeHTTPClient_ConnectTimeSSRFRejection` | VERIFIED |
| Redirect Target Validation | `CheckRedirect` re-validating URL syntax and target IP | `TestSafeHTTPClient_RedirectHandling` | VERIFIED |
| Payload Size Throttling | `ReadLimitedPayload` stream reader (50 MB limit) | `TestReadLimitedPayload_StreamingSizeEnforcement` | VERIFIED |
| Geospatial RFC 7946 CRS | Default WGS 84 (`EPSG:4326`) and explicit CRS checks | `TestValidateGeoJSON_RFC7946DefaultCRS` | VERIFIED |
| Finite Coordinate Audit | `validatePosition` rejecting `NaN`, `Inf`, and out-of-range floats | `TestValidateGeoJSON_InvalidAndMalformedCoordinates` | VERIFIED |
| Unreviewed Promotion Gate | Validation report existence check before operational use | `TestIngestionService_JobAndPromotionLifecycle` | VERIFIED |
| SQL Injection Protection | Parameterized `pgx` queries across all repositories | Repository inspection and unit tests | VERIFIED |
| Security Headers & CORS | Middleware applying nosniff, DENY frame options, origin whitelist | Middleware test and inspection | VERIFIED |
| Audit Trail Logging | `audit_events` immutable database table | `internal/audit/audit_test.go` | VERIFIED |

---

## 4. Remaining Production Security Recommendations

1. **Production Reverse Proxy Rate Limiting:** Enforce token-bucket rate limits (e.g. 100 requests per minute per IP) at the Cloudflare or Nginx ingress layer for public endpoints.
2. **PostgreSQL Row-Level Security (RLS):** Apply RLS policies to restrict field inspector case updates strictly to cases assigned to their registered user identifier.
3. **Automated Secret Rotation:** Rotate JWT signing keys and database connection credentials on a scheduled 90-day cycle using AWS Secrets Manager or HashiCorp Vault.
