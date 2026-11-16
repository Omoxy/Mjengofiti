import React from 'react';
import { Compass, Layers, CheckSquare, Activity } from 'lucide-react';

interface GisStatusBarProps {
  cursorCoords: { lng: number; lat: number } | null;
  zoom: number;
  activeLayersCount: number;
  selectedFeaturesCount: number;
  isDegraded: boolean;
}

export const GisStatusBar: React.FC<GisStatusBarProps> = ({
  cursorCoords,
  zoom,
  activeLayersCount,
  selectedFeaturesCount,
  isDegraded,
}) => {
  return (
    <footer className="gis-status-bar" aria-label="Geospatial Viewport Telemetry">
      {/* Left: Coordinate Telemetry & Elevation */}
      <div className="gis-status-group">
        <div className="gis-status-item">
          <Compass size={11} color="var(--gis-cyan)" />
          {cursorCoords ? (
            <span>
              Lat: {cursorCoords.lat.toFixed(5)}°, Lng: {cursorCoords.lng.toFixed(5)}°
            </span>
          ) : (
            <span>Lat: -1.28639°, Lng: 36.81722° (Nairobi CBD)</span>
          )}
        </div>

        <div className="gis-status-item" style={{ color: 'var(--text-muted)' }}>
          <span>EPSG:4326 (WGS84)</span>
        </div>
      </div>

      {/* Center: Scale & Viewport Zoom */}
      <div className="gis-status-group">
        <div className="gis-status-item">
          <span>Zoom: {zoom.toFixed(1)}</span>
        </div>
      </div>

      {/* Right: Active Layers, Selection, System Status */}
      <div className="gis-status-group">
        <div className="gis-status-item">
          <Layers size={11} color="var(--gis-cyan)" />
          <span>{activeLayersCount} Active Layers</span>
        </div>

        <div className="gis-status-item">
          <CheckSquare size={11} color={selectedFeaturesCount > 0 ? 'var(--gis-cyan)' : 'var(--text-muted)'} />
          <span>{selectedFeaturesCount} Selected</span>
        </div>

        <div className="gis-status-item">
          <Activity size={11} color={isDegraded ? 'var(--gis-amber)' : 'var(--gis-emerald)'} />
          <span style={{ color: isDegraded ? 'var(--gis-amber)' : 'var(--gis-emerald)' }}>
            {isDegraded ? 'PostGIS Offline (Degraded)' : 'PostGIS Live'}
          </span>
        </div>
      </div>
    </footer>
  );
};
