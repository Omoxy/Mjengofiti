import { useState, useEffect, useCallback } from 'react';
import { api, type Building, type BuildingDetail } from '../api/client';

interface UseBuildingsParams {
  bbox?: [number, number, number, number] | null;
  query?: string;
  limit?: number;
}

export function useBuildings({
  bbox,
  query = '',
  limit = 50,
}: UseBuildingsParams = {}) {
  const [buildings, setBuildings] = useState<Building[]>([]);
  const [total, setTotal] = useState(0);
  const [selectedBuilding, setSelectedBuilding] = useState<BuildingDetail | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingDetail, setIsLoadingDetail] = useState(false);
  const [isDegraded, setIsDegraded] = useState(false);
  const [degradedMessage, setDegradedMessage] = useState('');

  const fetchBuildings = useCallback(async () => {
    setIsLoading(true);
    try {
      const bboxParam = bbox ? bbox.join(',') : undefined;
      const res = await api.getBuildings({
        bbox: bboxParam,
        q: query || undefined,
        limit,
      });

      setBuildings(res.buildings || []);
      setTotal(res.total || 0);
      setIsDegraded(false);
      setDegradedMessage('');
    } catch (err: unknown) {
      const msg =
        err instanceof Error
          ? err.message
          : 'PostgreSQL/PostGIS spatial database is offline.';
      setIsDegraded(true);
      setDegradedMessage(msg);
      setBuildings([]);
      setTotal(0);
    } finally {
      setIsLoading(false);
    }
  }, [bbox, query, limit]);

  useEffect(() => {
    let active = true;
    const execute = async () => {
      if (active) await fetchBuildings();
    };
    execute();
    return () => {
      active = false;
    };
  }, [fetchBuildings]);

  const selectBuilding = useCallback(async (b: Building) => {
    setIsLoadingDetail(true);
    try {
      const detail = await api.getBuildingDetail(b.id);
      setSelectedBuilding(detail);
    } catch {
      // Degraded fallback: construct detail from available building attributes
      setSelectedBuilding({
        ...b,
        observations: [],
        indicators: [],
      });
    } finally {
      setIsLoadingDetail(false);
    }
  }, []);

  const clearSelection = useCallback(() => {
    setSelectedBuilding(null);
  }, []);

  return {
    buildings,
    total,
    selectedBuilding,
    isLoading,
    isLoadingDetail,
    isDegraded,
    degradedMessage,
    selectBuilding,
    clearSelection,
    refetch: fetchBuildings,
  };
}
