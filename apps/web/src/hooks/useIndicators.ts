import { useState, useEffect, useCallback } from 'react';
import { api, type IndicatorDefinition } from '../api/client';

export function useIndicators() {
  const [indicators, setIndicators] = useState<IndicatorDefinition[]>([]);
  const [total, setTotal] = useState(0);
  const [disclaimer, setDisclaimer] = useState<string>('');
  const [isLoading, setIsLoading] = useState(false);
  const [isTriggering, setIsTriggering] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchIndicators = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await api.getIndicators();
      setIndicators(res.definitions || []);
      setTotal(res.total || 0);
      setDisclaimer(res.disclaimer || '');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load indicator definitions');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchIndicators();
  }, [fetchIndicators]);

  const triggerAnalysis = useCallback(async (indicatorName: string) => {
    setIsTriggering(true);
    try {
      return await api.triggerAnalysisJob(indicatorName);
    } finally {
      setIsTriggering(false);
    }
  }, []);

  return {
    indicators,
    total,
    disclaimer,
    isLoading,
    isTriggering,
    error,
    triggerAnalysis,
    refetch: fetchIndicators,
  };
}
