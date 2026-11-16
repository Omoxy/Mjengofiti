import React, { useState, useEffect, useMemo } from 'react';
import { NairobiMap } from '../map/NairobiMap';
import { BuildingDetailPanel } from './BuildingDetailPanel';
import { LayerContentsPanel, type LayerState } from './LayerContentsPanel';
import { BuildingDataTable } from './BuildingDataTable';
import { GisToolRail, type GisActiveTool } from '../../components/gis/GisToolRail';
import { GisStatusBar } from '../../components/gis/GisStatusBar';
import {
  BasemapGallery,
  MapLegend,
  BookmarksPanel,
  NAIROBI_BOOKMARKS,
} from '../../components/gis/GisFlyouts';
import { useBuildings } from '../../hooks';
import { api } from '../../api/client';
import type { BuildingDetail } from '../../api/client';

interface BuildingIntelligenceProps {
  searchTarget?: string;
  onResetSearchTarget?: () => void;
  onBuildingSelectNotify?: (building: any) => void;
  onStartInspection?: (building: any) => void;
}

export const BuildingIntelligence: React.FC<BuildingIntelligenceProps> = ({
  searchTarget,
  onResetSearchTarget,
  onBuildingSelectNotify,
  onStartInspection,
}) => {
  const [viewMode, setViewMode] = useState<'map' | 'table'>('map');
  const [activeTool, setActiveTool] = useState<GisActiveTool>(null);
  const [currentBasemap, setCurrentBasemap] = useState<string>('streets');
  const [targetView, setTargetView] = useState<{ center: [number, number]; zoom: number } | null>(null);
  const [cursorCoords, setCursorCoords] = useState<{ lng: number; lat: number } | null>(null);
  const [currentZoom, setCurrentZoom] = useState<number>(11.5);
  const [bbox, setBBox] = useState<[number, number, number, number] | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  // Handle global search extent jumps
  useEffect(() => {
    if (!searchTarget) return;
    const queryLower = searchTarget.toLowerCase().trim();
    setSearchQuery(searchTarget);

    // Check if coordinates format: e.g. "-1.2864, 36.8172" or "36.8172, -1.2864"
    const coordMatch = queryLower.match(/^(-?\d+\.?\d*)[,\s]+(-?\d+\.?\d*)$/);
    if (coordMatch) {
      const v1 = parseFloat(coordMatch[1]);
      const v2 = parseFloat(coordMatch[2]);
      if (v1 < 0 && v2 > 0) {
        setTargetView({ center: [v2, v1], zoom: 15 });
        onResetSearchTarget?.();
        return;
      } else if (v2 < 0 && v1 > 0) {
        setTargetView({ center: [v1, v2], zoom: 15 });
        onResetSearchTarget?.();
        return;
      }
    }

    // Match Nairobi spatial bookmarks
    const found = NAIROBI_BOOKMARKS.find(
      (b) =>
        b.name.toLowerCase().includes(queryLower) ||
        b.desc.toLowerCase().includes(queryLower)
    );
    if (found) {
      setTargetView({ center: found.center, zoom: found.zoom });
    }
    onResetSearchTarget?.();
  }, [searchTarget, onResetSearchTarget]);

  // Sourced Layer States
  const [layers, setLayers] = useState<LayerState[]>([
    {
      id: 'ms-footprints',
      name: 'Microsoft Kenya Building Footprints',
      group: 'buildings',
      visible: false,
      opacity: 0.9,
      status: 'candidate',
      source: 'Microsoft Maps / Bing Imagery',
      license: 'ODbL v1.0',
      featureCount: 'Evaluated catalogue entry (8 sample records in local PostGIS)',
      srid: 'EPSG:4326',
    },
    {
      id: 'google-openbldg',
      name: 'Google Open Buildings Temporal (v3)',
      group: 'buildings',
      visible: false,
      opacity: 0.8,
      status: 'candidate',
      source: 'Google Research',
      license: 'CC BY 4.0',
      featureCount: 'Evaluated catalogue entry (candidate for ingestion)',
      srid: 'EPSG:4326',
    },
    {
      id: 'osm-buildings',
      name: 'OpenStreetMap Kenya Buildings',
      group: 'buildings',
      visible: false,
      opacity: 0.8,
      status: 'candidate',
      source: 'Geofabrik / OSM Contributors',
      license: 'ODbL',
      srid: 'EPSG:4326',
    },
    {
      id: 'nairobi-county-boundary',
      name: 'Nairobi County Boundary (geoBoundaries Alternative)',
      group: 'admin',
      visible: true,
      opacity: 1.0,
      status: 'verified-alternative',
      source: 'geoBoundaries / RCMRD GeoPortal / Africa GeoPortal (2020 ADM1)',
      license: 'Public Domain',
      srid: 'EPSG:4326',
      mapLayers: [
        { id: 'nairobi-boundary-glow', opacityProperty: 'line-opacity', baseOpacity: 0.16 },
        { id: 'nairobi-boundary-line', opacityProperty: 'line-opacity', baseOpacity: 1 },
        { id: 'nairobi-boundary-fill', opacityProperty: 'fill-opacity', baseOpacity: 0.18 },
      ],
    },
    {
      id: 'sub-counties',
      name: 'Nairobi Sub-County Boundaries',
      group: 'admin',
      visible: false,
      opacity: 0.8,
      status: 'offline',
      source: 'No sub-county geometry loaded',
      license: 'Not provided',
      srid: 'Not loaded',
    },
    {
      id: 'wards',
      name: 'Electoral Wards (85 Units)',
      group: 'admin',
      visible: false,
      opacity: 0.6,
      status: 'candidate',
      source: 'IEBC Boundary Review',
      license: 'Statutory Gazette',
      srid: 'EPSG:4326',
    },
    {
      id: 'cop-dem',
      name: 'Copernicus DEM 30m Slope & Elevation',
      group: 'environment',
      visible: false,
      opacity: 0.7,
      status: 'candidate',
      source: 'European Space Agency (ESA)',
      license: 'Copernicus Open Access',
      srid: 'EPSG:4326',
    },
    {
      id: 'riparian-buffer',
      name: '30m River Riparian Reserve Buffer',
      group: 'environment',
      visible: false,
      opacity: 0.85,
      status: 'candidate',
      source: 'PLUPA 2019 / WRMA',
      license: 'Statutory Reserve (Sec 69)',
      srid: 'EPSG:4326',
    },
    {
      id: 'chirps-rainfall',
      name: 'CHIRPS v3.0 Rainfall Anomaly Grid',
      group: 'environment',
      visible: false,
      opacity: 0.6,
      status: 'candidate',
      source: 'UCSB Climate Hazards Center',
      license: 'Public Domain',
      srid: 'EPSG:4326',
    },
    {
      id: 'postgis-buildings',
      name: 'PostGIS building records',
      group: 'buildings',
      visible: true,
      opacity: 0.9,
      status: 'api',
      source: 'Mjengofiti Go API / PostgreSQL + PostGIS',
      license: 'See record-level source attribution',
      srid: 'EPSG:4326',
      mapLayers: [{ id: 'buildings-points', opacityProperty: 'circle-opacity', baseOpacity: 0.9 }],
    },
  ]);

  const {
    buildings,
    total,
    selectedBuilding,
    isLoading,
    isLoadingDetail,
    isDegraded,
    selectBuilding,
    clearSelection,
    refetch,
  } = useBuildings({
    bbox,
    query: searchQuery,
    limit: 50,
  });

  const handleToggleLayer = (id: string) => {
    setLayers((prev) =>
      prev.map((l) => (l.id === id ? { ...l, visible: !l.visible } : l))
    );
  };

  const handleChangeOpacity = (id: string, opacity: number) => {
    setLayers((prev) =>
      prev.map((l) => (l.id === id ? { ...l, opacity } : l))
    );
  };

  const handleSelectBookmark = (center: [number, number], zoom: number) => {
    setTargetView({ center, zoom });
  };

  const handleReferForInspection = async (buildingItem: any, priority: string, reason: string) => {
    try {
      await api.createInspectionCase({
        building_id: buildingItem.id,
        title: `Screening Triage Inspection: ${buildingItem.external_id || buildingItem.id.substring(0, 8)}`,
        description: reason,
        priority: priority as 'low' | 'medium' | 'high' | 'critical',
        status: 'open',
        referral_reason: reason,
        evidence_refs: [`dataset:${buildingItem.source_name || 'ms-footprints'}`],
        assigned_to: 'NCA Field Inspectorate',
      });
      console.info('Inspection referral successfully created in Go backend:', {
        buildingId: buildingItem.id,
        priority,
      });
    } catch (err) {
      console.error('Failed to create inspection referral:', err);
    }
  };

  const mapLayerSettings = useMemo(
    () => Object.fromEntries(
      layers.flatMap((layer) => (layer.mapLayers ?? []).map((mapLayer) => [
        mapLayer.id,
        {
          visible: layer.visible,
          opacity: layer.opacity,
          opacityProperty: mapLayer.opacityProperty,
          baseOpacity: mapLayer.baseOpacity,
        },
      ] as const)),
    ),
    [layers],
  );
  const activeLayersCount = layers.filter((layer) => layer.visible && layer.mapLayers?.length).length;

  return (
    <div className="building-intelligence-workspace" style={{ position: 'relative', width: '100%', height: '100%', overflow: 'hidden' }}>
      {/* Central Full-Canvas Map View vs Attribute Table */}
      <div style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}>
        {viewMode === 'map' ? (
          <NairobiMap
            onBBoxChange={setBBox}
            basemap={currentBasemap}
            buildings={buildings}
            mapLayerSettings={mapLayerSettings}
            onCursorMove={setCursorCoords}
            onZoomChange={setCurrentZoom}
            targetView={targetView}
            selectedBuildingId={selectedBuilding?.id}
            onBuildingSelect={async (id) => {
              const b = buildings.find((x) => x.id === id);
              if (b) {
                await selectBuilding(b);
                setActiveTool(null);
                onBuildingSelectNotify?.(b);
              } else {
                try {
                  const detail = await api.getBuildingDetail(id);
                  selectBuilding(detail as any);
                  setActiveTool(null);
                  onBuildingSelectNotify?.(detail);
                } catch (e) {
                  console.warn('Could not fetch detail for building', id, e);
                }
              }
            }}
            isMeasuring={activeTool === 'measure'}
            onStopMeasuring={() => setActiveTool(null)}
          />
        ) : (
          <div style={{ paddingTop: '72px', height: '100%', boxSizing: 'border-box' }}>
            <BuildingDataTable
              buildings={buildings}
              total={total}
              selectedBuildingId={selectedBuilding?.id}
              onSelectBuilding={(building) => {
                selectBuilding(building);
                setActiveTool(null);
              }}
              isLoading={isLoading}
              isDegraded={isDegraded}
              onRefresh={refetch}
            />
          </div>
        )}
      </div>

      {/* Felt Floating GIS Tool Rail */}
      <GisToolRail
        activeTool={activeTool}
        onToggleTool={(tool) => setActiveTool(tool)}
        viewMode={viewMode}
        onToggleViewMode={() => setViewMode((m) => (m === 'map' ? 'table' : 'map'))}
      />

      {/* Floating Side Drawers: Layers, Basemaps, Legend, Bookmarks, Filter */}
      {activeTool === 'layers' && (
        <LayerContentsPanel
          layers={layers}
          onToggleLayer={handleToggleLayer}
          onChangeOpacity={handleChangeOpacity}
          onClose={() => setActiveTool(null)}
        />
      )}

      {activeTool === 'basemaps' && (
        <BasemapGallery
          currentBasemap={currentBasemap}
          onSelectBasemap={(bm) => {
            setCurrentBasemap(bm);
            setActiveTool(null);
          }}
          onClose={() => setActiveTool(null)}
        />
      )}

      {activeTool === 'legend' && (
        <MapLegend onClose={() => setActiveTool(null)} />
      )}

      {activeTool === 'bookmarks' && (
        <BookmarksPanel
          onSelectBookmark={(center, zoom) => {
            handleSelectBookmark(center, zoom);
            setActiveTool(null);
          }}
          onClose={() => setActiveTool(null)}
        />
      )}

      {activeTool === 'context' ? (
        <BuildingDetailPanel
          building={null}
          isLoading={false}
          onClose={() => setActiveTool(null)}
          onJumpToBookmark={handleSelectBookmark}
          bbox={bbox}
        />
      ) : selectedBuilding ? (
        <BuildingDetailPanel
          building={selectedBuilding as BuildingDetail}
          isLoading={isLoadingDetail}
          onClose={clearSelection}
          onReferForInspection={handleReferForInspection}
          onStartInspection={onStartInspection}
          onJumpToBookmark={handleSelectBookmark}
          bbox={bbox}
        />
      ) : null}

      {/* Bottom Floating Map Status Pill */}
      <GisStatusBar
        cursorCoords={cursorCoords}
        zoom={currentZoom}
        activeLayersCount={activeLayersCount}
        selectedFeaturesCount={selectedBuilding ? 1 : 0}
        isDegraded={isDegraded}
      />
    </div>
  );
};
