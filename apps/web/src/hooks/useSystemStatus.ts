import { useState, useEffect, useCallback } from 'react';
import { api, type SystemStatus } from '../api/client';

export interface CoverageInfo {
  database_records?: number;
  coverage: string;
  status?: string;
  note?: string;
}

export function useSystemStatus(pollIntervalMs?: number) {
  const [status, setStatus] = useState<SystemStatus | null>(null);
  const [coverage, setCoverage] = useState<CoverageInfo | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchStatus = useCallback(async () => {
    try {
      const [statusRes, coverageRes] = await Promise.all([
        api.getSystemStatus().catch(() => null),
        api.getCoverage().catch(() => null),
      ]);
      if (statusRes) setStatus(statusRes);
      if (coverageRes) setCoverage(coverageRes);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'System health check unreachable');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStatus();
    if (pollIntervalMs && pollIntervalMs > 0) {
      const timer = setInterval(fetchStatus, pollIntervalMs);
      return () => clearInterval(timer);
    }
  }, [fetchStatus, pollIntervalMs]);

  const isDegraded =
    status?.database?.status === 'degraded' ||
    coverage?.status === 'database_offline' ||
    status?.status === 'degraded';

  return {
    status,
    coverage,
    isDegraded,
    isLoading,
    error,
    refetch: fetchStatus,
  };
}
