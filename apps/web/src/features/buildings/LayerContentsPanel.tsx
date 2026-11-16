import React, { useState } from 'react';
import {
  Layers,
  ChevronDown,
  ChevronRight,
  ChevronUp,
  Info,
  X,
} from 'lucide-react';

export interface LayerState {
  id: string;
  name: string;
  group: 'buildings' | 'admin' | 'environment';
  visible: boolean;
  opacity: number;
  status: 'api' | 'operational' | 'candidate' | 'authoritative' | 'verified-alternative' | 'statutory' | 'offline';
  source: string;
  license: string;
  featureCount?: string;
  srid: string;
  mapLayers?: Array<{
    id: string;
    opacityProperty: 'circle-opacity' | 'line-opacity' | 'fill-opacity';
    baseOpacity: number;
  }>;
}

interface LayerContentsPanelProps {
  layers: LayerState[];
  onToggleLayer: (id: string) => void;
  onChangeOpacity: (id: string, opacity: number) => void;
  onClose: () => void;
}

export const LayerContentsPanel: React.FC<LayerContentsPanelProps> = ({
  layers,
  onToggleLayer,
  onChangeOpacity,
  onClose,
}) => {
  const [expandedGroups, setExpandedGroups] = useState<Record<string, boolean>>({
    buildings: true,
    admin: true,
    environment: true,
  });

  const [expandedDetails, setExpandedDetails] = useState<Record<string, boolean>>({});
  const [isCollapsed, setIsCollapsed] = useState(false);

  const toggleGroup = (group: string) => {
    setExpandedGroups((prev) => ({ ...prev, [group]: !prev[group] }));
  };

  const toggleDetails = (id: string) => {
    setExpandedDetails((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  const buildingLayers = layers.filter((l) => l.group === 'buildings');
  const adminLayers = layers.filter((l) => l.group === 'admin');
  const envLayers = layers.filter((l) => l.group === 'environment');

  const renderBadge = (status: LayerState['status']) => {
    switch (status) {
      case 'api':
        return <span className="badge badge-info">API records</span>;
      case 'authoritative':
        return <span className="badge badge-success">Authoritative</span>;
      case 'verified-alternative':
        return <span className="badge badge-info">Verified alternative</span>;
      case 'operational':
        return <span className="badge badge-promoted">Evaluated</span>;
      case 'candidate':
        return <span className="badge badge-info">Candidate</span>;
      case 'statutory':
        return <span className="badge badge-warning">Statutory Buffer</span>;
      case 'offline':
        return <span className="badge badge-draft">Database Offline</span>;
      default:
        return null;
    }
  };

  const renderLayerItem = (layer: LayerState) => {
    const isDetailsOpen = expandedDetails[layer.id];
    const isMapLayer = Boolean(layer.mapLayers?.length);

    return (
      <div key={layer.id} className="gis-layer-item">
        <div className="gis-layer-main-row">
          <label className="gis-layer-label">
            <input
              type="checkbox"
              checked={layer.visible}
              disabled={!isMapLayer}
              aria-label={`Show ${layer.name} on map`}
              title={isMapLayer ? `Toggle ${layer.name} visibility` : 'Catalog entry; this dataset is not loaded in the map'}
              onChange={() => onToggleLayer(layer.id)}
              style={{ accentColor: 'var(--wine)', cursor: isMapLayer ? 'pointer' : 'not-allowed' }}
            />
            <span style={{ color: layer.visible ? 'var(--text-primary)' : 'var(--text-muted)' }}>
              {layer.name}
            </span>
          </label>

          <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
            {renderBadge(layer.status)}
            {!isMapLayer && <span className="gis-layer-availability">Catalog</span>}

            <button
              type="button"
              className="btn btn-ghost btn-sm"
              onClick={() => toggleDetails(layer.id)}
              title="Layer details & provenance"
              style={{ padding: '2px 4px' }}
            >
              <Info size={11} color={isDetailsOpen ? 'var(--gis-cyan)' : 'var(--text-muted)'} />
            </button>
          </div>
        </div>

        {/* Expandable Layer Metadata & Opacity Slider */}
        {isDetailsOpen && (
          <div
            style={{
              marginTop: '4px',
              padding: '6px 8px',
              backgroundColor: 'var(--bg-surface-elevated)',
              borderRadius: 'var(--radius-xs)',
              fontSize: '0.6875rem',
              display: 'flex',
              flexDirection: 'column',
              gap: '4px',
              border: '1px solid var(--border-subtle)',
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
              <span>Source:</span>
              <span style={{ color: 'var(--text-primary)', fontWeight: 500 }}>{layer.source}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
              <span>License:</span>
              <span style={{ color: 'var(--text-primary)' }}>{layer.license}</span>
            </div>
            {layer.featureCount && (
              <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
                <span>Features:</span>
                <span style={{ color: 'var(--text-primary)', fontFamily: 'var(--font-mono)' }}>{layer.featureCount}</span>
              </div>
            )}
            <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
              <span>Spatial Reference:</span>
              <span style={{ color: 'var(--gis-cyan)', fontFamily: 'var(--font-mono)' }}>{layer.srid}</span>
            </div>

            {/* Opacity Slider */}
            {isMapLayer && <div style={{ marginTop: '4px', display: 'flex', alignItems: 'center', gap: '8px' }}>
              <span style={{ color: 'var(--text-muted)', fontSize: '0.65rem' }}>Opacity:</span>
              <input
                type="range"
                min="0"
                max="100"
                value={Math.round(layer.opacity * 100)}
                onChange={(e) => onChangeOpacity(layer.id, parseInt(e.target.value, 10) / 100)}
                style={{ flex: 1, accentColor: 'var(--gis-cyan)', height: '4px', cursor: 'pointer' }}
              />
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: '0.65rem', minWidth: '24px' }}>
                {Math.round(layer.opacity * 100)}%
              </span>
            </div>}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className={`gis-side-panel gis-layers-panel${isCollapsed ? ' is-collapsed' : ''}`}>
      <div className="gis-panel-header">
        <div className="gis-panel-title">
          <Layers size={14} color="var(--gis-cyan)" />
          <span>Map Contents & Layers</span>
        </div>
        <div className="gis-panel-actions">
          <button type="button" className="btn btn-ghost btn-sm" onClick={() => setIsCollapsed((collapsed) => !collapsed)} aria-label={isCollapsed ? 'Expand layers panel' : 'Collapse layers panel'} title={isCollapsed ? 'Expand panel' : 'Collapse panel'}>
            {isCollapsed ? <ChevronDown size={15} /> : <ChevronUp size={15} />}
          </button>
          <button type="button" className="btn btn-ghost btn-sm" onClick={onClose} aria-label="Close layers panel" title="Close panel">
            <X size={15} />
          </button>
        </div>
      </div>

      {!isCollapsed && <div className="gis-panel-body">
        {/* Layer Group 1: Building Footprints */}
        <div className="gis-layer-group">
          <button type="button" className="gis-layer-group-header" onClick={() => toggleGroup('buildings')} aria-expanded={expandedGroups.buildings}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
              {expandedGroups.buildings ? <ChevronDown size={13} /> : <ChevronRight size={13} />}
              <span>Building Footprints</span>
            </div>
            <span style={{ fontSize: '0.6875rem', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}>
              {buildingLayers.filter((l) => l.visible).length}/{buildingLayers.length}
            </span>
          </button>
          {expandedGroups.buildings && buildingLayers.map(renderLayerItem)}
        </div>

        {/* Layer Group 2: Administrative Boundaries */}
        <div className="gis-layer-group">
          <button type="button" className="gis-layer-group-header" onClick={() => toggleGroup('admin')} aria-expanded={expandedGroups.admin}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
              {expandedGroups.admin ? <ChevronDown size={13} /> : <ChevronRight size={13} />}
              <span>Administrative Boundaries</span>
            </div>
            <span style={{ fontSize: '0.6875rem', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}>
              {adminLayers.filter((l) => l.visible).length}/{adminLayers.length}
            </span>
          </button>
          {expandedGroups.admin && adminLayers.map(renderLayerItem)}
        </div>

        {/* Layer Group 3: Environmental & Risk Context */}
        <div className="gis-layer-group">
          <button type="button" className="gis-layer-group-header" onClick={() => toggleGroup('environment')} aria-expanded={expandedGroups.environment}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
              {expandedGroups.environment ? <ChevronDown size={13} /> : <ChevronRight size={13} />}
              <span>Environmental & Risk Context</span>
            </div>
            <span style={{ fontSize: '0.6875rem', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}>
              {envLayers.filter((l) => l.visible).length}/{envLayers.length}
            </span>
          </button>
          {expandedGroups.environment && envLayers.map(renderLayerItem)}
        </div>
      </div>}
    </div>
  );
};
