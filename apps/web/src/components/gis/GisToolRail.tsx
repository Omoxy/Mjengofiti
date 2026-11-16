import React from 'react';
import {
  Layers,
  Map,
  ListTree,
  Bookmark,
  Ruler,
  TableProperties,
  Info,
} from 'lucide-react';

export type GisActiveTool = 'layers' | 'context' | 'basemaps' | 'legend' | 'bookmarks' | 'measure' | null;

interface GisToolRailProps {
  activeTool: GisActiveTool;
  onToggleTool: (tool: GisActiveTool) => void;
  viewMode: 'map' | 'table';
  onToggleViewMode: () => void;
}

export const GisToolRail: React.FC<GisToolRailProps> = ({
  activeTool,
  onToggleTool,
  viewMode,
  onToggleViewMode,
}) => {
  return (
    <aside className="gis-tool-rail" aria-label="Geospatial Analysis & Navigation Tools">
      {/* Layer Contents Tool */}
      <button
        type="button"
        className={`gis-tool-btn ${activeTool === 'layers' ? 'active' : ''}`}
        onClick={() => onToggleTool(activeTool === 'layers' ? null : 'layers')}
        title="Map Contents & Layers"
        aria-label="Map Contents & Layers"
        aria-pressed={activeTool === 'layers'}
      >
        <Layers size={16} />
      </button>

      <button
        type="button"
        className={`gis-tool-btn ${activeTool === 'context' ? 'active' : ''}`}
        onClick={() => onToggleTool(activeTool === 'context' ? null : 'context')}
        title="Spatial overview"
        aria-label="Spatial overview"
        aria-pressed={activeTool === 'context'}
      >
        <Info size={16} />
      </button>

      {/* Basemap Switcher */}
      <button
        type="button"
        className={`gis-tool-btn ${activeTool === 'basemaps' ? 'active' : ''}`}
        onClick={() => onToggleTool(activeTool === 'basemaps' ? null : 'basemaps')}
        title="Basemap Gallery"
        aria-label="Basemap Gallery"
        aria-pressed={activeTool === 'basemaps'}
      >
        <Map size={16} />
      </button>

      {/* Map Legend */}
      <button
        type="button"
        className={`gis-tool-btn ${activeTool === 'legend' ? 'active' : ''}`}
        onClick={() => onToggleTool(activeTool === 'legend' ? null : 'legend')}
        title="Map Legend & Symbology"
        aria-label="Map Legend & Symbology"
        aria-pressed={activeTool === 'legend'}
      >
        <ListTree size={16} />
      </button>

      {/* Nairobi Spatial Bookmarks */}
      <button
        type="button"
        className={`gis-tool-btn ${activeTool === 'bookmarks' ? 'active' : ''}`}
        onClick={() => onToggleTool(activeTool === 'bookmarks' ? null : 'bookmarks')}
        title="Spatial Extent Bookmarks"
        aria-label="Spatial Extent Bookmarks"
        aria-pressed={activeTool === 'bookmarks'}
      >
        <Bookmark size={16} />
      </button>

      {/* Measurement Tool */}
      <button
        type="button"
        className={`gis-tool-btn ${activeTool === 'measure' ? 'active' : ''}`}
        onClick={() => onToggleTool(activeTool === 'measure' ? null : 'measure')}
        title="Geospatial Measurement (Distance / Area)"
        aria-label="Geospatial Measurement"
        aria-pressed={activeTool === 'measure'}
      >
        <Ruler size={16} />
      </button>

      {/* Divider */}
      <div style={{ width: '20px', height: '1px', backgroundColor: 'var(--border-subtle)', margin: '4px 0' }} />

      {/* View Mode Toggle (Map vs Table) */}
      <button
        type="button"
        className={`gis-tool-btn ${viewMode === 'table' ? 'active' : ''}`}
        onClick={onToggleViewMode}
        title={viewMode === 'map' ? 'Switch to Attribute Table View' : 'Switch to Full Map View'}
        aria-label={viewMode === 'map' ? 'Switch to Attribute Table View' : 'Switch to Full Map View'}
      >
        <TableProperties size={16} />
      </button>
    </aside>
  );
};
