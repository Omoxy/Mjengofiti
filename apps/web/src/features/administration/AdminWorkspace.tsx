import React from 'react';
import { AlertTriangle, HardDrive, RefreshCw, Server } from 'lucide-react';
import { useSystemStatus } from '../../hooks';

export const AdminWorkspace: React.FC = () => {
  const { status, coverage, isDegraded, isLoading, error, refetch } = useSystemStatus(15000);
  const databaseConnected = status?.database.status === 'connected';
  const coverageAvailable = coverage?.status === 'available' && typeof coverage.database_records === 'number';

  return (
    <div className="workspace-page">
      <div className="workspace-header">
        <div>
          <h2 className="workspace-title">
            <Server size={20} aria-hidden="true" />
            <span>System status</span>
          </h2>
          <p className="workspace-subtitle">
            Current API and database health. Detailed infrastructure versions and host information are not shown here.
          </p>
        </div>
        <button type="button" className="btn btn-secondary" onClick={refetch} disabled={isLoading}>
          <RefreshCw size={14} className={isLoading ? 'spin' : ''} />
          <span>Refresh status</span>
        </button>
      </div>

      {error && (
        <div className="mjengo-empty-state" role="alert">
          <h3>Status unavailable</h3>
          <p>The system status service could not be reached.</p>
        </div>
      )}

      {isDegraded && !error && (
        <div className="mjengo-empty-state" role="status">
          <AlertTriangle size={18} aria-hidden="true" />
          <h3>Service degraded</h3>
          <p>One or more required services are unavailable. Live building queries may not be available.</p>
        </div>
      )}

      <div className="card-grid">
        <section className="card" aria-labelledby="api-status-title">
          <h3 id="api-status-title"><Server size={15} aria-hidden="true" /> API</h3>
          <p>{isLoading ? 'Checking' : status?.status === 'ok' ? 'Available' : status?.status ?? 'Unavailable'}</p>
          {typeof status?.uptime_seconds === 'number' && (
            <small>Process uptime: {Math.max(0, Math.floor(status.uptime_seconds))} seconds</small>
          )}
        </section>

        <section className="card" aria-labelledby="database-status-title">
          <h3 id="database-status-title"><HardDrive size={15} aria-hidden="true" /> PostgreSQL / PostGIS</h3>
          <p>{isLoading ? 'Checking' : databaseConnected ? 'Connected' : 'Unavailable'}</p>
          <small>Only coarse connection state is reported.</small>
        </section>

        <section className="card" aria-labelledby="coverage-status-title">
          <h3 id="coverage-status-title">Building records</h3>
          {coverageAvailable ? (
            <>
              <p>{coverage.database_records} database records</p>
              <small>Record count only; this is not a verified inventory or coverage estimate.</small>
            </>
          ) : (
            <p>{coverage?.note ?? 'Live record count unavailable.'}</p>
          )}
        </section>
      </div>

      <section className="card" aria-labelledby="ingestion-status-title">
        <h3 id="ingestion-status-title">Dataset ingestion</h3>
        <p>On-demand ingestion endpoints are available to authorized data stewards.</p>
        <small>No background worker health endpoint is currently implemented.</small>
      </section>
    </div>
  );
};
