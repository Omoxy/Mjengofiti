# Mjengofiti — Backup and Disaster Recovery Runbook

## 1. Database Backup Strategy
All core building records, inspection findings, audit logs, and dataset catalogue records reside in PostgreSQL.

### 1.1 Automated Daily Backups
```bash
pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -F c -b -v -f /backups/mjengofiti_$(date +%Y%m%d_%H%M%S).dump
```

### 1.2 Point-In-Time Recovery (PITR)
- Enable WAL archiving in `postgresql.conf`:
  ```
  wal_level = replica
  archive_mode = on
  archive_command = 'test ! -f /wal_archives/%f && cp %p /wal_archives/%f'
  ```

---

## 2. Restoration Procedures

### 2.1 Full Restore from Dump
```bash
# 1. Terminate active application connections
# 2. Restore schema and data
pg_restore -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c -v /backups/mjengofiti_target.dump

# 3. Verify spatial index integrity
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "REINDEX INDEX idx_buildings_geom;"
```
