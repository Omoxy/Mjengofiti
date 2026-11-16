import React, { useState, useMemo } from 'react';
import type { Building } from '../../api/client';
import { useBuildings } from '../../hooks';
import { Search, FileCheck2, MapPin } from 'lucide-react';

interface BuildingRecordsWorkspaceProps {
  onSelectBuilding: (building: Building) => void;
  onStartInspection: (building: Building) => void;
  selectedBuildingId?: string | null;
}

export const BuildingRecordsWorkspace: React.FC<BuildingRecordsWorkspaceProps> = ({
  onSelectBuilding,
  onStartInspection,
  selectedBuildingId,
}) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [sourceFilter, setSourceFilter] = useState('all');

  const { buildings, isLoading, isDegraded } = useBuildings({
    query: searchQuery,
    limit: 100,
  });

  const filteredBuildings = useMemo(() => {
    return buildings.filter((b) => {
      if (sourceFilter !== 'all' && b.source_name !== sourceFilter) {
        return false;
      }
      return true;
    });
  }, [buildings, sourceFilter]);

  return (
    <div className="building-records-workspace" role="region" aria-label="Building records catalogue">
      {/* Search and Filters Strip */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          gap: '12px',
          marginBottom: '16px',
          flexWrap: 'wrap',
        }}
      >
        <div style={{ position: 'relative', width: '320px', maxWidth: '100%' }}>
          <Search
            size={14}
            color="var(--muted)"
            style={{ position: 'absolute', left: '12px', top: '12px' }}
          />
          <input
            type="text"
            className="mjengo-select"
            style={{ paddingLeft: '34px' }}
            placeholder="Search building records..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            aria-label="Search building records"
          />
        </div>

        <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
          <select
            className="mjengo-select"
            style={{ width: '180px' }}
            value={sourceFilter}
            onChange={(e) => setSourceFilter(e.target.value)}
            aria-label="Filter source dataset"
          >
            <option value="all">All datasets</option>
            <option value="ms-footprints">Microsoft Footprints</option>
            <option value="google-openbldg">Google Open Buildings</option>
            <option value="nairobi-county">Nairobi County Registry</option>
          </select>
        </div>
      </div>

      {/* Buildings Table */}
      <div
        style={{
          backgroundColor: 'var(--paper)',
          border: '1px solid var(--line)',
          borderRadius: '14px',
          overflow: 'hidden',
          boxShadow: 'var(--shadow-subtle)',
        }}
      >
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '13px' }}>
            <thead>
              <tr style={{ backgroundColor: 'var(--brand-cream-light)', borderBottom: '1px solid var(--line)', textAlign: 'left' }}>
                <th style={{ padding: '12px 16px', color: 'var(--brand-maroon)', fontWeight: 700, fontSize: '11px', textTransform: 'uppercase' }}>
                  Building
                </th>
                <th style={{ padding: '12px 16px', color: 'var(--brand-maroon)', fontWeight: 700, fontSize: '11px', textTransform: 'uppercase' }}>
                  Location
                </th>
                <th style={{ padding: '12px 16px', color: 'var(--brand-maroon)', fontWeight: 700, fontSize: '11px', textTransform: 'uppercase' }}>
                  Record Status
                </th>
                <th style={{ padding: '12px 16px', color: 'var(--brand-maroon)', fontWeight: 700, fontSize: '11px', textTransform: 'uppercase' }}>
                  Source Dataset
                </th>
                <th style={{ padding: '12px 16px', textAlign: 'right' }}></th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr>
                  <td colSpan={5} style={{ textAlign: 'center', padding: '48px 16px', color: 'var(--muted)' }}>
                    Loading verified building records...
                  </td>
                </tr>
              ) : filteredBuildings.length === 0 ? (
                <tr>
                  <td colSpan={5} style={{ textAlign: 'center', padding: '64px 20px' }}>
                    <b style={{ display: 'block', color: 'var(--ink)', fontSize: '15px', marginBottom: '8px' }}>
                      Building list is empty
                    </b>
                    <span style={{ fontSize: '12px', color: 'var(--muted)', display: 'block', maxWidth: '380px', margin: '0 auto' }}>
                      {isDegraded
                        ? 'PostgreSQL/PostGIS is running in degraded mode. Connect the database to query live building records.'
                        : 'Load approved building records to fill this list.'}
                    </span>
                  </td>
                </tr>
              ) : (
                filteredBuildings.map((b) => {
                  const isSelected = selectedBuildingId === b.id;
                  return (
                    <tr
                      key={b.id}
                      onClick={() => onSelectBuilding(b)}
                      style={{
                        borderBottom: '1px solid var(--brand-border)',
                        cursor: 'pointer',
                        backgroundColor: isSelected ? 'var(--winep)' : undefined,
                      }}
                    >
                      <td style={{ padding: '14px 16px', fontWeight: 700, color: 'var(--ink)' }}>
                        <div>{b.external_id || b.id.substring(0, 12)}</div>
                        <div style={{ fontSize: '11px', color: 'var(--muted)', fontWeight: 400 }}>
                          Area: {b.area_sq_m ? `${Math.round(b.area_sq_m)} m²` : 'Unspecified'}
                        </div>
                      </td>
                      <td style={{ padding: '14px 16px', color: 'var(--brand-ink)' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                          <MapPin size={12} color="var(--wine)" />
                          <span>
                            {b.latitude.toFixed(4)}°, {b.longitude.toFixed(4)}°
                          </span>
                        </div>
                      </td>
                      <td style={{ padding: '14px 16px' }}>
                        <span
                          style={{
                            padding: '3px 8px',
                            borderRadius: '5px',
                            fontSize: '11px',
                            fontWeight: 600,
                            backgroundColor: 'var(--brand-cream)',
                            color: 'var(--brand-maroon)',
                          }}
                        >
                          {b.data_quality || 'Evaluated'}
                        </span>
                      </td>
                      <td style={{ padding: '14px 16px', color: 'var(--brand-maroon)' }}>
                        {b.source_name}
                      </td>
                      <td style={{ padding: '14px 16px', textAlign: 'right' }}>
                        <button
                          type="button"
                          className="mjengo-btn alt mjengo-btn-sm"
                          onClick={(e) => {
                            e.stopPropagation();
                            onStartInspection(b);
                          }}
                          title="Start field inspection for this building"
                        >
                          <FileCheck2 size={12} />
                          <span>Inspect</span>
                        </button>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
