import React from 'react';
import {
  Map,
  ListTree,
  Bookmark,
  Check,
  Navigation,
} from 'lucide-react';

export interface BasemapOption {
  id: string;
  name: string;
  type: string;
  previewColor: string;
}

export const BASEMAPS: BasemapOption[] = [
  { id: 'streets', name: 'MapTiler Streets', type: 'Streets v2', previewColor: 'var(--brand-cream)' },
  { id: 'satellite', name: 'MapTiler Satellite', type: 'Satellite style', previewColor: 'var(--brand-cream)' },
];

export const NAIROBI_BOOKMARKS = [
  { name: 'Nairobi County Overview', center: [36.8822, -1.3021] as [number, number], zoom: 9.5, desc: 'geoBoundaries ADM1 source extent' },
  { name: 'Central Business District (CBD)', center: [36.8219, -1.2833] as [number, number], zoom: 15, desc: 'High-density commercial core' },
  { name: 'Westlands Sub-County', center: [36.8044, -1.2639] as [number, number], zoom: 14, desc: 'Commercial & high-rise mixed development' },
  { name: 'Kilimani / Kileleshwa', center: [36.7878, -1.2917] as [number, number], zoom: 14, desc: 'Densifying residential multi-storey zone' },
  { name: 'Eastleigh (Kamukunji)', center: [36.8486, -1.2758] as [number, number], zoom: 15, desc: 'Extreme urban density commercial hub' },
  { name: 'Embakasi East / Pipeline', center: [36.8967, -1.3167] as [number, number], zoom: 15, desc: 'High-density residential tenement corridor' },
  { name: 'Kasarani Sub-County', center: [36.8972, -1.2222] as [number, number], zoom: 14, desc: 'Rapidly growing residential peri-urban zone' },
  { name: 'Kibra / Ngong River Corridor', center: [36.7833, -1.3167] as [number, number], zoom: 15, desc: 'Informal settlement & riparian drainage basin' },
];

interface BasemapGalleryProps {
  currentBasemap: string;
  onSelectBasemap: (id: string) => void;
  onClose: () => void;
}

export const BasemapGallery: React.FC<BasemapGalleryProps> = ({
  currentBasemap,
  onSelectBasemap,
  onClose,
}) => {
  return (
    <div className="gis-side-panel" style={{ width: '280px' }}>
      <div className="gis-panel-header">
        <div className="gis-panel-title">
          <Map size={14} color="var(--brand-maroon)" />
          <span>Basemap Gallery</span>
        </div>
        <button type="button" className="btn btn-ghost btn-sm" onClick={onClose}>✕</button>
      </div>
      <div className="gis-panel-body">
        <div style={{ display: 'grid', gridTemplateColumns: '1fr', gap: '8px' }}>
          {BASEMAPS.map((bm) => {
            const isSelected = currentBasemap === bm.id;
            return (
              <div
                key={bm.id}
                onClick={() => onSelectBasemap(bm.id)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '10px',
                  padding: '8px 10px',
                  borderRadius: 'var(--radius-sm)',
                  backgroundColor: isSelected ? 'var(--gis-cyan-subtle)' : 'var(--bg-surface-elevated)',
                  border: isSelected ? '1px solid var(--gis-cyan)' : '1px solid var(--border-subtle)',
                  cursor: 'pointer',
                  transition: 'all 0.15s ease',
                }}
              >
                <div
                  style={{
                    width: '32px',
                    height: '32px',
                    borderRadius: 'var(--radius-xs)',
                    backgroundColor: bm.previewColor,
                    border: '1px solid var(--border-medium)',
                    flexShrink: 0,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  {isSelected && <Check size={16} color="var(--brand-maroon)" />}
                </div>
                <div style={{ display: 'flex', flexDirection: 'column' }}>
                  <span style={{ fontSize: '0.775rem', fontWeight: 600, color: 'var(--text-primary)' }}>{bm.name}</span>
                  <span style={{ fontSize: '0.675rem', color: 'var(--text-muted)' }}>{bm.type}</span>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

interface MapLegendProps {
  onClose: () => void;
}

export const MapLegend: React.FC<MapLegendProps> = ({ onClose }) => {
  return (
    <div className="gis-side-panel gis-legend-panel">
      <div className="gis-panel-header">
        <div className="gis-panel-title">
          <ListTree size={14} color="var(--brand-maroon)" />
          <span>Map legend</span>
        </div>
        <button type="button" className="btn btn-ghost btn-sm" onClick={onClose}>✕</button>
      </div>
      <div className="gis-panel-body gis-legend-body">
        <div className="gis-legend-row">
          <span className="gis-legend-symbol building" aria-hidden="true" />
          <span>PostGIS building records</span>
        </div>
        <div className="gis-legend-row">
          <span className="gis-legend-symbol boundary" aria-hidden="true" />
          <span>Nairobi County ADM1 boundary (verified alternative)</span>
        </div>
        <p>Other registered datasets are catalog entries, not rendered overlays.</p>
      </div>
    </div>
  );
};

interface BookmarksPanelProps {
  onSelectBookmark: (center: [number, number], zoom: number) => void;
  onClose: () => void;
}

export const BookmarksPanel: React.FC<BookmarksPanelProps> = ({
  onSelectBookmark,
  onClose,
}) => {
  return (
    <div className="gis-side-panel" style={{ width: '300px' }}>
      <div className="gis-panel-header">
        <div className="gis-panel-title">
          <Bookmark size={14} color="var(--brand-maroon)" />
          <span>Spatial Bookmarks (Nairobi)</span>
        </div>
        <button type="button" className="btn btn-ghost btn-sm" onClick={onClose}>✕</button>
      </div>
      <div className="gis-panel-body">
        <p style={{ fontSize: '0.725rem', color: 'var(--text-muted)', marginBottom: '8px' }}>
          Select an administrative sub-county to zoom the viewport:
        </p>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
          {NAIROBI_BOOKMARKS.map((bm) => (
            <div
              key={bm.name}
              onClick={() => onSelectBookmark(bm.center, bm.zoom)}
              style={{
                padding: '8px 10px',
                borderRadius: 'var(--radius-sm)',
                backgroundColor: 'var(--bg-surface-elevated)',
                border: '1px solid var(--border-subtle)',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                transition: 'all 0.15s ease',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.borderColor = 'var(--brand-maroon)';
                e.currentTarget.style.backgroundColor = 'var(--brand-maroon-soft)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.borderColor = 'var(--border-subtle)';
                e.currentTarget.style.backgroundColor = 'var(--bg-surface-elevated)';
              }}
            >
              <div style={{ display: 'flex', flexDirection: 'column' }}>
                <span style={{ fontSize: '0.775rem', fontWeight: 600, color: 'var(--text-primary)' }}>{bm.name}</span>
                <span style={{ fontSize: '0.675rem', color: 'var(--text-muted)' }}>{bm.desc}</span>
              </div>
              <Navigation size={12} color="var(--brand-maroon)" />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export interface SpatialFilterCriteria {
  subCounty: string;
  minHeight: number;
  minArea: number;
  riparianOnly: boolean;
  sourceDataset: string;
}

interface SpatialFilterPanelProps {
  criteria: SpatialFilterCriteria;
  onApplyCriteria: (criteria: SpatialFilterCriteria) => void;
  onResetCriteria: () => void;
  onClose: () => void;
}

export const SpatialFilterPanel: React.FC<SpatialFilterPanelProps> = ({
  criteria,
  onApplyCriteria,
  onResetCriteria,
  onClose,
}) => {
  const [localCriteria, setLocalCriteria] = React.useState<SpatialFilterCriteria>(criteria);

  return (
    <div className="gis-side-panel" style={{ width: '310px' }}>
      <div className="gis-panel-header">
        <div className="gis-panel-title">
          <span>Spatial & Attribute Filters</span>
        </div>
        <button type="button" className="btn btn-ghost btn-sm" onClick={onClose}>✕</button>
      </div>

      <div className="gis-panel-body" style={{ gap: '14px' }}>
        <div>
          <label style={{ fontSize: '0.725rem', fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>
            Administrative Sub-County
          </label>
          <select
            className="select"
            style={{ width: '100%' }}
            value={localCriteria.subCounty}
            onChange={(e) => setLocalCriteria({ ...localCriteria, subCounty: e.target.value })}
          >
            <option value="all">All Nairobi Sub-Counties (17)</option>
            <option value="starehe">Starehe / Central Business District</option>
            <option value="westlands">Westlands</option>
            <option value="dagoretti">Dagoretti North / South</option>
            <option value="kilimani">Kilimani / Kileleshwa</option>
            <option value="kamukunji">Kamukunji / Eastleigh</option>
            <option value="embakasi">Embakasi East / Pipeline</option>
            <option value="kasarani">Kasarani / Roysambu</option>
            <option value="kibra">Kibra / Lang'ata</option>
          </select>
        </div>

        <div>
          <label style={{ fontSize: '0.725rem', fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>
            Data Source Register
          </label>
          <select
            className="select"
            style={{ width: '100%' }}
            value={localCriteria.sourceDataset}
            onChange={(e) => setLocalCriteria({ ...localCriteria, sourceDataset: e.target.value })}
          >
            <option value="all">All Evaluated Sources</option>
            <option value="microsoft">Microsoft Kenya Footprints (842k)</option>
            <option value="google">Google Open Buildings v3 (895k)</option>
            <option value="osm">OpenStreetMap Kenya</option>
          </select>
        </div>

        <div>
          <label style={{ fontSize: '0.725rem', fontWeight: 600, color: 'var(--text-secondary)', display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
            <span>Min Building Height:</span>
            <span style={{ fontFamily: 'var(--font-mono)', color: 'var(--gis-cyan)' }}>
              {localCriteria.minHeight > 0 ? `${localCriteria.minHeight}m (~${Math.round(localCriteria.minHeight / 3)} storeys)` : 'Any height'}
            </span>
          </label>
          <input
            type="range"
            min="0"
            max="45"
            step="3"
            value={localCriteria.minHeight}
            onChange={(e) => setLocalCriteria({ ...localCriteria, minHeight: parseInt(e.target.value, 10) })}
            style={{ width: '100%', accentColor: 'var(--gis-cyan)', cursor: 'pointer' }}
          />
        </div>

        <div>
          <label style={{ fontSize: '0.725rem', fontWeight: 600, color: 'var(--text-secondary)', display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
            <span>Min Footprint Area:</span>
            <span style={{ fontFamily: 'var(--font-mono)', color: 'var(--gis-cyan)' }}>
              {localCriteria.minArea > 0 ? `${localCriteria.minArea} m²` : 'Any area'}
            </span>
          </label>
          <input
            type="range"
            min="0"
            max="1000"
            step="50"
            value={localCriteria.minArea}
            onChange={(e) => setLocalCriteria({ ...localCriteria, minArea: parseInt(e.target.value, 10) })}
            style={{ width: '100%', accentColor: 'var(--gis-cyan)', cursor: 'pointer' }}
          />
        </div>

        <div style={{ padding: '8px', backgroundColor: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-subtle)' }}>
          <label style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer', fontSize: '0.75rem' }}>
            <input
              type="checkbox"
              checked={localCriteria.riparianOnly}
              onChange={(e) => setLocalCriteria({ ...localCriteria, riparianOnly: e.target.checked })}
              style={{ accentColor: 'var(--gis-cyan)' }}
            />
            <span style={{ fontWeight: 500, color: 'var(--text-primary)' }}>
              Highlight 30m Riparian Buffer Encroachment (PLUPA Sec 69)
            </span>
          </label>
        </div>

        <div style={{ display: 'flex', gap: '8px', marginTop: '6px' }}>
          <button
            type="button"
            className="btn btn-secondary"
            style={{ flex: 1 }}
            onClick={() => {
              onResetCriteria();
              setLocalCriteria({
                subCounty: 'all',
                minHeight: 0,
                minArea: 0,
                riparianOnly: false,
                sourceDataset: 'all',
              });
            }}
          >
            Reset
          </button>
          <button
            type="button"
            className="btn btn-primary"
            style={{ flex: 1 }}
            onClick={() => {
              onApplyCriteria(localCriteria);
              onClose();
            }}
          >
            Apply Query
          </button>
        </div>
      </div>
    </div>
  );
};
