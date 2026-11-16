import React, { useEffect, useRef, useState, useMemo } from 'react';
import * as maplibregl from 'maplibre-gl';
import { api, type Building, type MapConfig } from '../../api/client';
import { AlertCircle, Compass, Ruler, RotateCcw, X, Check } from 'lucide-react';

interface NairobiMapProps {
  onBBoxChange?: (bbox: [number, number, number, number]) => void;
  onBuildingSelect?: (buildingId: string) => void;
  selectedBuildingId?: string | null;
  selectedCoords?: [number, number] | null;
  basemap?: string;
  onCursorMove?: (coords: { lng: number; lat: number }) => void;
  onZoomChange?: (zoom: number) => void;
  targetView?: { center: [number, number]; zoom: number } | null;
  isMeasuring?: boolean;
  onStopMeasuring?: () => void;
  buildings?: Building[];
  mapLayerSettings: Record<string, {
    visible: boolean;
    opacity: number;
    opacityProperty: 'circle-opacity' | 'line-opacity' | 'fill-opacity';
    baseOpacity: number;
  }>;
  showMapTilerNotice?: boolean;
}

const BASEMAP_TILES: Record<string, { tiles: string[]; maxzoom: number; attribution: string }> = {
  'dark-matter': {
    tiles: ['https://a.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}@2x.png'],
    maxzoom: 19,
    attribution: '&copy; OpenStreetMap contributors &copy; CARTO',
  },
  'positron': {
    tiles: ['https://a.basemaps.cartocdn.com/light_all/{z}/{x}/{y}@2x.png'],
    maxzoom: 19,
    attribution: '&copy; OpenStreetMap contributors &copy; CARTO',
  },
  'osm': {
    tiles: ['https://tile.openstreetmap.org/{z}/{x}/{y}.png'],
    maxzoom: 19,
    attribution: '&copy; OpenStreetMap contributors',
  },
  'satellite': {
    tiles: ['https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}'],
    maxzoom: 18,
    attribution: 'Tiles &copy; Esri',
  },
};

// Geodesic calculations on WGS84 sphere
function haversineDistance(c1: [number, number], c2: [number, number]): number {
  const R = 6371008.8; // Earth mean radius in meters
  const lat1 = (c1[1] * Math.PI) / 180;
  const lat2 = (c2[1] * Math.PI) / 180;
  const dLat = ((c2[1] - c1[1]) * Math.PI) / 180;
  const dLng = ((c2[0] - c1[0]) * Math.PI) / 180;
  const a =
    Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos(lat1) * Math.cos(lat2) * Math.sin(dLng / 2) * Math.sin(dLng / 2);
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  return R * c;
}

function sphericalPolygonArea(coords: Array<[number, number]>): number {
  const R = 6371008.8;
  if (coords.length < 3) return 0;
  let total = 0;
  const len = coords.length;
  for (let i = 0; i < len; i++) {
    const p1 = coords[i];
    const p2 = coords[(i + 1) % len];
    const p1Lat = (p1[1] * Math.PI) / 180;
    const p2Lat = (p2[1] * Math.PI) / 180;
    const dLng = ((p2[0] - p1[0]) * Math.PI) / 180;
    total += dLng * (2 + Math.sin(p1Lat) + Math.sin(p2Lat));
  }
  return Math.abs((total * R * R) / 4);
}

export const NairobiMap: React.FC<NairobiMapProps> = ({
  onBBoxChange,
  onBuildingSelect,
  selectedBuildingId: _selectedBuildingId,
  selectedCoords,
  basemap = 'positron',
  onCursorMove,
  onZoomChange,
  targetView,
  isMeasuring = false,
  onStopMeasuring,
  buildings = [],
  mapLayerSettings,
  showMapTilerNotice = true,
}) => {
  const mapContainer = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<maplibregl.Map | null>(null);
  const boundaryBoundsRef = useRef<maplibregl.LngLatBounds | null>(null);
  const [mapError, setMapError] = useState<string | null>(null);
  const [measurePoints, setMeasurePoints] = useState<Array<[number, number]>>([]);
  const [mapConfig, setMapConfig] = useState<MapConfig | null>(null);
  const [isConfigLoaded, setIsConfigLoaded] = useState(false);
  const isMeasuringRef = useRef(isMeasuring);
  const layerSettingsRef = useRef(mapLayerSettings);

  useEffect(() => {
    isMeasuringRef.current = isMeasuring;
    if (!isMeasuring) {
      setMeasurePoints([]);
    }
  }, [isMeasuring]);

  // Load project map config
  useEffect(() => {
    let isMounted = true;
    api
      .getMapConfig()
      .then((cfg) => {
        if (isMounted) {
          setMapConfig(cfg);
          setIsConfigLoaded(true);
        }
      })
      .catch(() => {
        if (isMounted) {
          setIsConfigLoaded(true);
        }
      });
    return () => {
      isMounted = false;
    };
  }, []);

  useEffect(() => {
    layerSettingsRef.current = mapLayerSettings;
    const map = mapRef.current;
    if (!map || !map.isStyleLoaded()) return;

    Object.entries(mapLayerSettings).forEach(([layerId, settings]) => {
      if (!map.getLayer(layerId)) return;
      map.setLayoutProperty(layerId, 'visibility', settings.visible ? 'visible' : 'none');
      map.setPaintProperty(layerId, settings.opacityProperty, settings.baseOpacity * settings.opacity);
    });
  }, [mapLayerSettings]);

  const isMapTilerConfigured = useMemo(() => {
    const maptilerKey = (import.meta as any).env?.VITE_MAPTILER_KEY;
    const tileUrl = (import.meta as any).env?.VITE_MAP_TILE_URL || mapConfig?.tile_url || '';
    return Boolean(maptilerKey || tileUrl.includes('maptiler.com'));
  }, [mapConfig]);

  // Compute live geodesic distance and spherical area
  const { totalDistanceMeters, totalAreaSqMeters } = useMemo(() => {
    if (measurePoints.length < 2) {
      return { totalDistanceMeters: 0, totalAreaSqMeters: 0 };
    }
    let dist = 0;
    for (let i = 0; i < measurePoints.length - 1; i++) {
      dist += haversineDistance(measurePoints[i], measurePoints[i + 1]);
    }
    const area = measurePoints.length >= 3 ? sphericalPolygonArea(measurePoints) : 0;
    return { totalDistanceMeters: dist, totalAreaSqMeters: area };
  }, [measurePoints]);

  // Initialize MapLibre
  useEffect(() => {
    if (!mapContainer.current || mapRef.current || !isConfigLoaded) return;

    let mapInstance: maplibregl.Map | null = null;
    try {
      const activeTileConfig = BASEMAP_TILES[basemap] || BASEMAP_TILES['positron'];
      const defaultCenter: [number, number] = mapConfig?.center || [36.817223, -1.286389];
      const defaultZoom = mapConfig?.zoom || 11.5;

      const maptilerKey = (import.meta as any).env?.VITE_MAPTILER_KEY;
      const getMaptilerStyle = (key: string, base: string) => {
        if (base === 'satellite') return `https://api.maptiler.com/maps/satellite/style.json?key=${key}`;
        if (base === 'basic') return `https://api.maptiler.com/maps/basic-v2/style.json?key=${key}`;
        if (base === 'dataviz') return `https://api.maptiler.com/maps/dataviz/style.json?key=${key}`;
        return `https://api.maptiler.com/maps/streets-v2/style.json?key=${key}`;
      };

      const maptilerUrl = (import.meta as any).env?.VITE_MAP_TILE_URL ||
        (maptilerKey ? getMaptilerStyle(maptilerKey, basemap) : null);

      const styleDefinition = maptilerUrl
        ? maptilerUrl
        : {
            version: 8,
            sources: {
              'basemap-tiles': {
                type: 'raster' as const,
                tiles: activeTileConfig.tiles,
                tileSize: 256,
                attribution: activeTileConfig.attribution,
              },
            },
            layers: [
              {
                id: 'basemap-tiles-layer',
                type: 'raster' as const,
                source: 'basemap-tiles',
                minzoom: 0,
                maxzoom: activeTileConfig.maxzoom,
              },
            ],
          };

      mapInstance = new maplibregl.Map({
        container: mapContainer.current,
        style: styleDefinition as any,
        center: defaultCenter,
        zoom: defaultZoom,
        maxBounds: [
          [36.4, -1.6],
          [37.3, -1.0],
        ],
      });

      mapInstance.addControl(new maplibregl.NavigationControl({ showCompass: true }), 'top-right');
      mapInstance.addControl(new maplibregl.ScaleControl({ unit: 'metric' }), 'bottom-left');

      const applyMapLayerSettings = (layerId: string) => {
        const settings = layerSettingsRef.current[layerId];
        if (!mapInstance || !settings || !mapInstance.getLayer(layerId)) return;
        mapInstance.setLayoutProperty(layerId, 'visibility', settings.visible ? 'visible' : 'none');
        mapInstance.setPaintProperty(layerId, settings.opacityProperty, settings.baseOpacity * settings.opacity);
      };

      mapInstance.on('load', async () => {
        const brandTokens = getComputedStyle(document.documentElement);
        const brandMaroon = brandTokens.getPropertyValue('--brand-maroon').trim();
        const brandWhite = brandTokens.getPropertyValue('--brand-white').trim();
        const brandCream = brandTokens.getPropertyValue('--brand-cream').trim();

        mapInstance?.getStyle().layers.forEach((layer) => {
          const layerName = layer.id.toLowerCase();
          if (layer.type === 'background') {
            mapInstance?.setPaintProperty(layer.id, 'background-color', brandCream);
          } else if (layer.type === 'symbol' && layer.layout?.['text-field']) {
            mapInstance?.setPaintProperty(layer.id, 'text-color', brandMaroon);
            mapInstance?.setPaintProperty(layer.id, 'text-halo-color', brandWhite);
            mapInstance?.setPaintProperty(layer.id, 'text-halo-width', 1);
          } else if (layer.type === 'fill' && layerName.includes('water')) {
            mapInstance?.setPaintProperty(layer.id, 'fill-color', brandWhite);
          } else if (layer.type === 'fill' && /land|park|building/.test(layerName)) {
            mapInstance?.setPaintProperty(layer.id, 'fill-color', brandCream);
          } else if (layer.type === 'line' && /road|highway|transport/.test(layerName)) {
            mapInstance?.setPaintProperty(layer.id, 'line-color', brandWhite);
          }
        });

        try {
          const boundaries = await api.getBoundaries();
          if (boundaries && mapInstance && !mapInstance.getSource('nairobi-boundary')) {
            const boundaryBounds = new maplibregl.LngLatBounds();
            const extendBoundaryCoordinates = (coordinates: unknown): void => {
              if (!Array.isArray(coordinates)) return;
              if (coordinates.length >= 2 && typeof coordinates[0] === 'number' && typeof coordinates[1] === 'number') {
                boundaryBounds.extend([coordinates[0], coordinates[1]]);
                return;
              }
              coordinates.forEach(extendBoundaryCoordinates);
            };

            boundaries.features?.forEach((feature: { geometry?: { coordinates?: unknown } }) => {
              extendBoundaryCoordinates(feature.geometry?.coordinates);
            });
            if (!boundaryBounds.isEmpty()) {
              boundaryBoundsRef.current = boundaryBounds;
              mapInstance.fitBounds(boundaryBounds, {
                padding: { top: 84, right: 72, bottom: 76, left: 72 },
                maxZoom: 9.5,
                duration: 0,
              });
            }

            mapInstance.addSource('nairobi-boundary', {
              type: 'geojson',
              data: boundaries,
            });

            // A restrained under-stroke keeps the boundary legible over road detail.
            mapInstance.addLayer({
              id: 'nairobi-boundary-glow',
              type: 'line',
              source: 'nairobi-boundary',
              paint: {
                'line-color': brandMaroon,
                'line-width': 2.6,
                'line-opacity': 0.16,
              },
            });

            // Clean boundary line
            mapInstance.addLayer({
              id: 'nairobi-boundary-line',
              type: 'line',
              source: 'nairobi-boundary',
              paint: {
                'line-color': brandMaroon,
                'line-width': 1.8,
                'line-opacity': 0.96,
              },
            });

            // Soft interior county wash
            mapInstance.addLayer({
              id: 'nairobi-boundary-fill',
              type: 'fill',
              source: 'nairobi-boundary',
              paint: {
                'fill-color': brandCream,
                'fill-opacity': 0.18,
              },
            });

            applyMapLayerSettings('nairobi-boundary-glow');
            applyMapLayerSettings('nairobi-boundary-line');
            applyMapLayerSettings('nairobi-boundary-fill');
          }
        } catch (err) {
          console.warn('Could not load administrative boundary layer:', err);
        }

        // Buildings source and layers
        if (mapInstance && !mapInstance.getSource('buildings-source')) {
          mapInstance.addSource('buildings-source', {
            type: 'geojson',
            data: { type: 'FeatureCollection', features: [] },
          });

          mapInstance.addLayer({
            id: 'buildings-points',
            type: 'circle',
            source: 'buildings-source',
            paint: {
              'circle-radius': 6,
              'circle-color': brandMaroon,
              'circle-stroke-width': 2,
              'circle-stroke-color': brandWhite,
              'circle-opacity': 0.9,
            },
          });
          applyMapLayerSettings('buildings-points');

          mapInstance.on('click', 'buildings-points', (e) => {
            if (e.features && e.features[0] && e.features[0].properties) {
              const bId = e.features[0].properties.id;
              if (bId && onBuildingSelect) {
                onBuildingSelect(bId);
              }
            }
          });

          mapInstance.on('mouseenter', 'buildings-points', () => {
            if (mapInstance) mapInstance.getCanvas().style.cursor = 'pointer';
          });
          mapInstance.on('mouseleave', 'buildings-points', () => {
            if (mapInstance) mapInstance.getCanvas().style.cursor = '';
          });
        }

        // Initialize measurement source and layers
        if (mapInstance && !mapInstance.getSource('measurement-source')) {
          mapInstance.addSource('measurement-source', {
            type: 'geojson',
            data: { type: 'FeatureCollection', features: [] },
          });

          // Measurement polygon fill
          mapInstance.addLayer({
            id: 'measurement-fill',
            type: 'fill',
            source: 'measurement-source',
            filter: ['==', '$type', 'Polygon'],
            paint: {
              'fill-color': brandMaroon,
              'fill-opacity': 0.15,
            },
          });

          // Measurement line
          mapInstance.addLayer({
            id: 'measurement-line',
            type: 'line',
            source: 'measurement-source',
            filter: ['any', ['==', '$type', 'LineString'], ['==', '$type', 'Polygon']],
            paint: {
              'line-color': brandMaroon,
              'line-width': 2.2,
              'line-dasharray': [2, 2],
            },
          });

          // Measurement points
          mapInstance.addLayer({
            id: 'measurement-points',
            type: 'circle',
            source: 'measurement-source',
            filter: ['==', '$type', 'Point'],
            paint: {
              'circle-radius': 5,
              'circle-color': brandMaroon,
              'circle-stroke-width': 2,
              'circle-stroke-color': brandWhite,
            },
          });
        }

        if (mapInstance && onBBoxChange) {
          const bounds = mapInstance.getBounds();
          onBBoxChange([bounds.getWest(), bounds.getSouth(), bounds.getEast(), bounds.getNorth()]);
        }
      });

      // Viewport movement listeners
      mapInstance.on('moveend', () => {
        if (mapInstance && onBBoxChange) {
          const bounds = mapInstance.getBounds();
          onBBoxChange([bounds.getWest(), bounds.getSouth(), bounds.getEast(), bounds.getNorth()]);
        }
        if (mapInstance && onZoomChange) {
          onZoomChange(mapInstance.getZoom());
        }
      });

      // Cursor coordinates tracking
      mapInstance.on('mousemove', (e) => {
        if (onCursorMove) {
          onCursorMove({ lng: e.lngLat.lng, lat: e.lngLat.lat });
        }
      });

      // Interactive Map Click (Measuring)
      mapInstance.on('click', (e) => {
        if (isMeasuringRef.current) {
          const pt: [number, number] = [e.lngLat.lng, e.lngLat.lat];
          setMeasurePoints((prev) => [...prev, pt]);
        }
      });

      mapRef.current = mapInstance;
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'WebGL or map initialization error';
      setTimeout(() => setMapError(msg), 0);
    }

    return () => {
      if (mapRef.current) {
        mapRef.current.remove();
        mapRef.current = null;
      }
    };
  }, [isConfigLoaded, basemap]);

  // Update building records on map
  useEffect(() => {
    const map = mapRef.current;
    if (!map || !map.isStyleLoaded()) return;

    const source = map.getSource('buildings-source') as maplibregl.GeoJSONSource | undefined;
    if (!source) return;

    const features = (buildings || [])
      .filter((b) => b.longitude && b.latitude)
      .map((b) => ({
        type: 'Feature' as const,
        properties: { id: b.id, name: b.external_id || b.id },
        geometry: {
          type: 'Point' as const,
          coordinates: [b.longitude, b.latitude],
        },
      }));

    source.setData({
      type: 'FeatureCollection',
      features,
    });
  }, [buildings]);

  // Update measurement GeoJSON source
  useEffect(() => {
    const map = mapRef.current;
    if (!map || !map.isStyleLoaded()) return;

    const source = map.getSource('measurement-source') as maplibregl.GeoJSONSource | undefined;
    if (!source) return;

    const features: any[] = [];

    measurePoints.forEach((pt, i) => {
      features.push({
        type: 'Feature',
        properties: { index: i + 1 },
        geometry: { type: 'Point', coordinates: pt },
      });
    });

    if (measurePoints.length >= 2) {
      features.push({
        type: 'Feature',
        properties: {},
        geometry: { type: 'LineString', coordinates: measurePoints },
      });
    }

    if (measurePoints.length >= 3) {
      features.push({
        type: 'Feature',
        properties: {},
        geometry: {
          type: 'Polygon',
          coordinates: [[...measurePoints, measurePoints[0]]],
        },
      });
    }

    source.setData({
      type: 'FeatureCollection',
      features,
    });
  }, [measurePoints]);

  // Handle Target View FlyTo
  useEffect(() => {
    if (targetView && mapRef.current) {
      mapRef.current.flyTo({
        center: targetView.center,
        zoom: targetView.zoom,
        essential: true,
        duration: 1200,
      });
    }
  }, [targetView]);

  // Handle Selected Building Pan
  useEffect(() => {
    if (selectedCoords && mapRef.current) {
      mapRef.current.flyTo({
        center: selectedCoords,
        zoom: 16,
        essential: true,
      });
    }
  }, [selectedCoords]);

  // Handle Panel Resizing using ResizeObserver
  useEffect(() => {
    if (!mapContainer.current) return;
    const observer = new ResizeObserver(() => {
      mapRef.current?.resize();
    });
    observer.observe(mapContainer.current);
    return () => observer.disconnect();
  }, []);

  const handleResetNairobiExtent = () => {
    const map = mapRef.current;
    const boundaryBounds = boundaryBoundsRef.current;
    if (map && boundaryBounds && !boundaryBounds.isEmpty()) {
      map.fitBounds(boundaryBounds, {
        padding: { top: 84, right: 72, bottom: 76, left: 72 },
        maxZoom: 9.5,
        duration: 500,
      });
    } else if (map) {
      map.flyTo({
        center: mapConfig?.center || [36.817223, -1.286389],
        zoom: mapConfig?.zoom || 11.5,
        essential: true,
      });
    }
  };

  const handleUndoMeasure = () => {
    setMeasurePoints((prev) => prev.slice(0, prev.length - 1));
  };

  const handleClearMeasure = () => {
    setMeasurePoints([]);
  };

  return (
    <div
      style={{
        position: 'relative',
        width: '100%',
        height: '100%',
        overflow: 'hidden',
        cursor: isMeasuring ? 'crosshair' : 'default',
      }}
    >
      {mapError ? (
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            height: '100%',
            padding: '2rem',
            textAlign: 'center',
            backgroundColor: 'var(--cream)',
            color: 'var(--ink)',
          }}
        >
          <AlertCircle size={36} color="var(--wine)" style={{ marginBottom: '1rem' }} />
          <h3 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.5rem' }}>
            Interactive Map Canvas Notice
          </h3>
          <p style={{ color: 'var(--muted)', fontSize: '0.9rem', maxWidth: '420px', marginBottom: '1rem' }}>
            {mapError}. You can continue searching and filtering buildings using the structured records list.
          </p>
        </div>
      ) : (
        <>
          <div ref={mapContainer} style={{ width: '100%', height: '100%' }} />

          {/* Extent Quick Control */}
          <div
            style={{
              position: 'absolute',
              top: '14px',
              left: '14px',
              zIndex: 10,
              display: 'flex',
              gap: '6px',
            }}
          >
            <button
              type="button"
              onClick={handleResetNairobiExtent}
              className="mjengo-btn alt mjengo-btn-sm"
              title="Reset view to verified Nairobi County source extent"
              style={{
                boxShadow: 'var(--shadow-md)',
              }}
            >
              <Compass size={13} color="var(--wine)" />
              <span>Nairobi Extent</span>
            </button>
          </div>

          {/* Honest MapTiler Setup Note Card */}
          {showMapTilerNotice && (
            <div className="map-setup-note" role="region" aria-label="Map configuration notice">
              <b>MapTiler</b>
              {isMapTilerConfigured ? (
                <p>Style configured</p>
              ) : (
                <p>Style not configured</p>
              )}
            </div>
          )}

          {/* Interactive Geodesic Measurement HUD Card */}
          {isMeasuring && (
            <div
              className="felt-hud-card"
              role="region"
              aria-label="Geospatial Measurement Controls"
              style={{
                backgroundColor: 'var(--brand-white)',
                color: 'var(--ink)',
                borderColor: 'var(--line)',
                boxShadow: 'var(--shadow-floating)',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '6px', color: 'var(--wine)' }}>
                <Ruler size={14} />
                <span style={{ fontWeight: 700, fontSize: '0.75rem', letterSpacing: '0.02em' }}>
                  Geodesic Measure
                </span>
              </div>

              <div style={{ width: '1px', height: '16px', backgroundColor: 'var(--line)' }} />

              <div style={{ display: 'flex', alignItems: 'center', gap: '14px', fontSize: '0.75rem' }}>
                <div>
                  <span style={{ color: 'var(--muted)', marginRight: '4px' }}>Distance:</span>
                  <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600, color: 'var(--ink)' }}>
                    {totalDistanceMeters >= 1000
                      ? `${(totalDistanceMeters / 1000).toFixed(2)} km`
                      : `${totalDistanceMeters.toFixed(1)} m`}
                  </span>
                </div>

                {measurePoints.length >= 3 && (
                  <div>
                    <span style={{ color: 'var(--muted)', marginRight: '4px' }}>Area:</span>
                    <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600, color: 'var(--wine)' }}>
                      {totalAreaSqMeters >= 10000
                        ? `${(totalAreaSqMeters / 10000).toFixed(2)} ha`
                        : `${totalAreaSqMeters.toFixed(1)} m²`}
                    </span>
                  </div>
                )}

                <div style={{ color: 'var(--muted)', fontSize: '0.7rem' }}>
                  ({measurePoints.length} {measurePoints.length === 1 ? 'vertex' : 'vertices'})
                </div>
              </div>

              <div style={{ width: '1px', height: '16px', backgroundColor: 'var(--line)' }} />

              <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                <button
                  type="button"
                  className="mjengo-btn alt mjengo-btn-sm"
                  onClick={handleUndoMeasure}
                  disabled={measurePoints.length === 0}
                  title="Undo last clicked vertex"
                >
                  <RotateCcw size={11} />
                  <span>Undo</span>
                </button>

                <button
                  type="button"
                  className="mjengo-btn alt mjengo-btn-sm"
                  onClick={handleClearMeasure}
                  disabled={measurePoints.length === 0}
                  title="Clear all points"
                >
                  <span>Clear</span>
                </button>

                <button
                  type="button"
                  className="mjengo-btn mjengo-btn-sm"
                  onClick={onStopMeasuring}
                  title="Finish measuring"
                >
                  <Check size={11} />
                  <span>Done</span>
                </button>

                <button
                  type="button"
                  className="mjengo-btn-ghost"
                  onClick={onStopMeasuring}
                  title="Cancel measurement"
                  style={{ padding: '4px' }}
                >
                  <X size={13} />
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
};
