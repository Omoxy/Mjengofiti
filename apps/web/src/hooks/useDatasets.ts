import { useState, useEffect, useCallback } from 'react';
import {
  api,
  type Dataset,
  type DatasetVersion,
  type IngestionJob,
  type IngestionSource,
  type ValidationReport,
  type PromotionResult,
} from '../api/client';

export function useDatasets() {
  const [datasets, setDatasets] = useState<Dataset[]>([]);
  const [sources, setSources] = useState<IngestionSource[]>([]);
  const [jobs, setJobs] = useState<IngestionJob[]>([]);
  const [selectedDatasetId, setSelectedDatasetId] = useState<string | null>(null);
  const [versions, setVersions] = useState<DatasetVersion[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isActionLoading, setIsActionLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchCatalogue = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const [dsRes, srcRes, jobsRes] = await Promise.all([
        api.getDatasets(),
        api.getIngestionSources().catch(() => ({ sources: [], total: 0 })),
        api.getIngestionJobs().catch(() => ({ jobs: [], total: 0 })),
      ]);
      setDatasets(dsRes.datasets || []);
      setSources(srcRes.sources || []);
      setJobs(jobsRes.jobs || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load datasets catalogue');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchCatalogue();
  }, [fetchCatalogue]);

  // Load versions when a dataset is selected
  useEffect(() => {
    if (!selectedDatasetId) {
      setVersions([]);
      return;
    }
    let isMounted = true;
    api
      .getDatasetVersions(selectedDatasetId)
      .then((res) => {
        if (isMounted) setVersions(res.versions || []);
      })
      .catch(() => {
        if (isMounted) setVersions([]);
      });
    return () => {
      isMounted = false;
    };
  }, [selectedDatasetId]);

  const createJob = useCallback(
    async (datasetId: string, version?: string) => {
      setIsActionLoading(true);
      try {
        const job = await api.createIngestionJob({
          dataset_id: datasetId,
          version,
        });
        setJobs((prev) => [job, ...prev]);
        return job;
      } finally {
        setIsActionLoading(false);
      }
    },
    []
  );

  const validateGeoJSON = useCallback(
    async (payload: unknown, versionId?: string): Promise<ValidationReport> => {
      setIsActionLoading(true);
      try {
        return await api.validateGeoJSON(payload, versionId);
      } finally {
        setIsActionLoading(false);
      }
    },
    []
  );

  const promoteDataset = useCallback(
    async (
      datasetId: string,
      version: string,
      force?: boolean
    ): Promise<PromotionResult> => {
      setIsActionLoading(true);
      try {
        const res = await api.promoteDataset({
          dataset_id: datasetId,
          version,
          force,
        });
        // Refresh catalogue and versions after promotion
        await fetchCatalogue();
        if (selectedDatasetId === datasetId) {
          const vRes = await api.getDatasetVersions(datasetId);
          setVersions(vRes.versions || []);
        }
        return res;
      } finally {
        setIsActionLoading(false);
      }
    },
    [fetchCatalogue, selectedDatasetId]
  );

  return {
    datasets,
    sources,
    jobs,
    versions,
    selectedDatasetId,
    setSelectedDatasetId,
    isLoading,
    isActionLoading,
    error,
    createJob,
    validateGeoJSON,
    promoteDataset,
    refetch: fetchCatalogue,
  };
}
