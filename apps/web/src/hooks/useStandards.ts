import { useState, useEffect, useCallback } from 'react';
import { api, type Standard } from '../api/client';

export function useStandards() {
  const [standards, setStandards] = useState<Standard[]>([]);
  const [total, setTotal] = useState(0);
  const [message, setMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchStandards = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await api.getStandards();
      setStandards(res.standards || []);
      setTotal(res.total || 0);
      setMessage(res.message || '');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load statutory standards');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStandards();
  }, [fetchStandards]);

  return {
    standards,
    total,
    message,
    isLoading,
    error,
    refetch: fetchStandards,
  };
}
