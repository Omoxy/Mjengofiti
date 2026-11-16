import { useState, useEffect, useCallback } from 'react';
import { api, type InspectionCase } from '../api/client';

export function useInspections(initialStatus?: string) {
  const [cases, setCases] = useState<InspectionCase[]>([]);
  const [total, setTotal] = useState(0);
  const [statusFilter, setStatusFilter] = useState<string>(initialStatus || 'all');
  const [isLoading, setIsLoading] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchCases = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const filterParam = statusFilter !== 'all' ? statusFilter : undefined;
      const res = await api.getInspectionCases(filterParam);
      setCases(res.cases || []);
      setTotal(res.total || 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load inspection cases');
    } finally {
      setIsLoading(false);
    }
  }, [statusFilter]);

  useEffect(() => {
    fetchCases();
  }, [fetchCases]);

  const createCase = useCallback(
    async (data: Partial<InspectionCase>): Promise<InspectionCase> => {
      setIsSubmitting(true);
      try {
        const created = await api.createInspectionCase(data);
        setCases((prev) => [created, ...prev]);
        setTotal((prev) => prev + 1);
        return created;
      } finally {
        setIsSubmitting(false);
      }
    },
    []
  );

  const updateCaseStatus = useCallback(
    async (id: string, newStatus: string): Promise<InspectionCase> => {
      setIsSubmitting(true);
      try {
        const updated = await api.updateInspectionCaseStatus(id, newStatus);
        setCases((prev) =>
          prev.map((c) => (c.id === id ? { ...c, status: updated.status } : c))
        );
        return updated;
      } finally {
        setIsSubmitting(false);
      }
    },
    []
  );

  return {
    cases,
    total,
    statusFilter,
    setStatusFilter,
    isLoading,
    isSubmitting,
    error,
    createCase,
    updateCaseStatus,
    refetch: fetchCases,
  };
}
