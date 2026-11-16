import { useState, useCallback } from 'react';

// Default Nairobi County map center and initial bounding box
export const NAIROBI_CENTER: [number, number] = [36.8219, -1.2921];
export const NAIROBI_DEFAULT_ZOOM = 12;

export interface MapViewport {
  center: [number, number];
  zoom: number;
  bbox: [number, number, number, number] | null;
}

export function useMapViewport(
  initialCenter: [number, number] = NAIROBI_CENTER,
  initialZoom: number = NAIROBI_DEFAULT_ZOOM
) {
  const [viewport, setViewport] = useState<MapViewport>({
    center: initialCenter,
    zoom: initialZoom,
    bbox: null,
  });

  const updateBBox = useCallback((bbox: [number, number, number, number]) => {
    setViewport((prev) => ({ ...prev, bbox }));
  }, []);

  const updateCenterZoom = useCallback(
    (center: [number, number], zoom: number) => {
      setViewport((prev) => ({ ...prev, center, zoom }));
    },
    []
  );

  const resetToNairobi = useCallback(() => {
    setViewport({
      center: NAIROBI_CENTER,
      zoom: NAIROBI_DEFAULT_ZOOM,
      bbox: null,
    });
  }, []);

  return {
    viewport,
    center: viewport.center,
    zoom: viewport.zoom,
    bbox: viewport.bbox,
    updateBBox,
    updateCenterZoom,
    resetToNairobi,
  };
}
