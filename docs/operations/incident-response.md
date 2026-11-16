# Mjengofiti — Incident Response and Escalation Runbook

## 1. Incident Classification

| Severity | Definition | Target Response | Target Resolution |
|---|---|---|---|
| **P1 — Critical** | Complete platform outage; database unavailable; inspection case corruption; security breach | < 15 mins | < 2 hours |
| **P2 — Major** | Geospatial map rendering failure; ingestion pipeline failure; auth errors for inspection officers | < 1 hour | < 8 hours |
| **P3 — Minor** | Slow query response; non-critical metadata discrepancy in dataset catalogue | < 4 hours | < 24 hours |
| **P4 — Low** | UI cosmetic issue; minor typo in standards clause documentation | Next sprint | Next scheduled release |

---

## 2. Emergency Operational Procedures

### 2.1 Database Unavailability (Degraded Mode Activation)
- The Go backend automatically detects connection pool loss and switches to degraded mode.
- Health endpoint `/health/ready` returns HTTP 503.
- Map and static boundary capabilities remain functional if client-side cached.

### 2.2 Security Incident (Compromised Credential or Token Leak)
1. Invalidate active JWT signing keys.
2. Force credential rotation for all system accounts.
3. Query `audit_events` table for unauthorized actions using compromised subject ID.
4. Issue incident notification and preserve forensic audit logs.
