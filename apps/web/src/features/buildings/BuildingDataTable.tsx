import React, { useState, useMemo } from 'react';
import type { Building } from '../../api/client';
import {
  Search,
  Download,
  ChevronLeft,
  ChevronRight,
  ArrowUpDown,
  AlertCircle,
} from 'lucide-react';

interface BuildingDataTableProps {
  buildings: Building[];
  total?: number;
  selectedBuildingId?: string | null;
  onSelectBuilding: (building: Building) => void;
  isLoading?: boolean;
  isDegraded: boolean;
  onRefresh?: () => void;
}

type SortField = 'external_id' | 'source_name' | 'area_sq_m' | 'height_m' | 'data_quality';
type SortOrder = 'asc' | 'desc';

export const BuildingDataTable: React.FC<BuildingDataTableProps> = ({
  buildings,
  selectedBuildingId,
  onSelectBuilding,
  isDegraded,
}) => {
  const [filterQuery, setFilterQuery] = useState('');
  const [selectedSource, setSelectedSource] = useState('all');
  const [sortField, setSortField] = useState<SortField>('area_sq_m');
  const [sortOrder, setSortOrder] = useState<SortOrder>('desc');
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 20;

  // Filter & sort
  const filtered = useMemo(() => {
    return buildings
      .filter((b) => {
        if (selectedSource !== 'all' && !b.source_name.toLowerCase().includes(selectedSource.toLowerCase())) {
          return false;
        }
        if (filterQuery.trim()) {
          const q = filterQuery.toLowerCase();
          const matchId = (b.external_id || b.id).toLowerCase().includes(q);
          const matchSrc = b.source_name.toLowerCase().includes(q);
          return matchId || matchSrc;
        }
        return true;
      })
      .sort((a, b) => {
        let valA = a[sortField] ?? 0;
        let valB = b[sortField] ?? 0;
        if (typeof valA === 'string') {
          return sortOrder === 'asc'
            ? String(valA).localeCompare(String(valB))
            : String(valB).localeCompare(String(valA));
        }
        return sortOrder === 'asc' ? Number(valA) - Number(valB) : Number(valB) - Number(valA);
      });
  }, [buildings, filterQuery, selectedSource, sortField, sortOrder]);

  const totalPages = Math.max(1, Math.ceil(filtered.length / pageSize));
  const paginated = filtered.slice((currentPage - 1) * pageSize, currentPage * pageSize);

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortOrder('desc');
    }
  };

  const handleExportCSV = () => {
    if (filtered.length === 0) return;
    const headers = ['id', 'external_id', 'source_name', 'latitude', 'longitude', 'area_sq_m', 'height_m', 'data_quality'];
    const rows = filtered.map((b) =>
      [b.id, b.external_id || '', b.source_name, b.latitude, b.longitude, b.area_sq_m || '', b.height_m || '', b.data_quality].join(',')
    );
    const csvContent = 'data:text/csv;charset=utf-8,' + [headers.join(','), ...rows].join('\n');
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement('a');
    link.setAttribute('href', encodedUri);
    link.setAttribute('download', `mjengofiti_buildings_export_${Date.now()}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  return (
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column', height: '100%', backgroundColor: 'var(--bg-app)', overflow: 'hidden' }}>
      {/* Table Toolbar */}
      <div
        style={{
          padding: '8px 14px',
          backgroundColor: 'var(--bg-chrome)',
          borderBottom: '1px solid var(--border-subtle)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: '10px',
          flexShrink: 0,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flex: 1, maxWidth: '600px' }}>
          <div style={{ position: 'relative', flex: 1 }}>
            <Search size={12} color="var(--text-muted)" style={{ position: 'absolute', left: '8px', top: '7px' }} />
            <input
              type="text"
              className="input"
              style={{ width: '100%', paddingLeft: '26px', height: '26px', fontSize: '0.75rem' }}
              placeholder="Filter by building ID, external reference, or source..."
              value={filterQuery}
              onChange={(e) => {
                setFilterQuery(e.target.value);
                setCurrentPage(1);
              }}
            />
          </div>

          <select
            className="select"
            style={{ height: '26px', fontSize: '0.725rem' }}
            value={selectedSource}
            onChange={(e) => {
              setSelectedSource(e.target.value);
              setCurrentPage(1);
            }}
          >
            <option value="all">All Sourced Footprints</option>
            <option value="microsoft">Microsoft Footprints (842k)</option>
            <option value="google">Google Open Buildings (895k)</option>
            <option value="openstreetmap">OpenStreetMap Kenya</option>
          </select>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <button
            type="button"
            className="btn btn-secondary btn-sm"
            onClick={handleExportCSV}
            disabled={filtered.length === 0}
            title="Export filtered records as CSV"
          >
            <Download size={11} />
            <span>Export CSV</span>
          </button>
        </div>
      </div>

      {/* Main Table Area */}
      <div style={{ flex: 1, overflowY: 'auto' }}>
        {filtered.length === 0 ? (
          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              height: '100%',
              padding: '40px 20px',
              textAlign: 'center',
              color: 'var(--text-muted)',
            }}
          >
            <AlertCircle size={32} color="var(--gis-amber)" style={{ marginBottom: '12px' }} />
            <div style={{ fontSize: '0.875rem', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '4px' }}>
              {isDegraded ? 'Spatial Database Running in Degraded Mode' : 'No Building Records in Current Filter'}
            </div>
            <p style={{ fontSize: '0.75rem', maxWidth: '440px', lineHeight: 1.5 }}>
              {isDegraded
                ? 'PostgreSQL/PostGIS is currently unreachable. Live spatial bounding box polygon queries are paused. Candidate dataset catalogues (842,100 Microsoft footprints, 895,000 Google Open Buildings) remain fully evaluated in the Research & Data catalogue.'
                : 'No buildings matched the current filter criteria or active bounding box. Pan the map or clear filters to view more.'}
            </p>
          </div>
        ) : (
          <table className="data-table">
            <thead>
              <tr>
                <th onClick={() => handleSort('external_id')} style={{ cursor: 'pointer' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                    <span>Identifier</span>
                    <ArrowUpDown size={10} />
                  </div>
                </th>
                <th onClick={() => handleSort('source_name')} style={{ cursor: 'pointer' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                    <span>Source</span>
                    <ArrowUpDown size={10} />
                  </div>
                </th>
                <th>Coordinates (EPSG:4326)</th>
                <th onClick={() => handleSort('area_sq_m')} style={{ cursor: 'pointer' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                    <span>Footprint Area (m²)</span>
                    <ArrowUpDown size={10} />
                  </div>
                </th>
                <th onClick={() => handleSort('height_m')} style={{ cursor: 'pointer' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                    <span>Height (m)</span>
                    <ArrowUpDown size={10} />
                  </div>
                </th>
                <th onClick={() => handleSort('data_quality')} style={{ cursor: 'pointer' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                    <span>Quality</span>
                    <ArrowUpDown size={10} />
                  </div>
                </th>
              </tr>
            </thead>
            <tbody>
              {paginated.map((b) => {
                const isSelected = selectedBuildingId === b.id;
                return (
                  <tr
                    key={b.id}
                    onClick={() => onSelectBuilding(b)}
                    style={{
                      cursor: 'pointer',
                      backgroundColor: isSelected ? 'var(--gis-cyan-subtle)' : undefined,
                      borderLeft: isSelected ? '3px solid var(--gis-cyan)' : '3px solid transparent',
                    }}
                  >
                    <td style={{ fontFamily: 'var(--font-mono)', fontWeight: 600, color: 'var(--gis-cyan)' }}>
                      {b.external_id || b.id.substring(0, 10)}
                    </td>
                    <td>{b.source_name}</td>
                    <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.725rem' }}>
                      {b.latitude.toFixed(5)}°, {b.longitude.toFixed(5)}°
                    </td>
                    <td style={{ fontFamily: 'var(--font-mono)' }}>
                      {b.area_sq_m ? `${b.area_sq_m.toFixed(1)} m²` : '—'}
                    </td>
                    <td style={{ fontFamily: 'var(--font-mono)' }}>
                      {b.height_m ? `${b.height_m.toFixed(1)} m` : 'Unmeasured'}
                    </td>
                    <td>
                      <span className={`badge ${b.data_quality === 'verified' ? 'badge-success' : 'badge-warning'}`}>
                        {b.data_quality}
                      </span>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>

      {/* Pagination Footer */}
      {filtered.length > 0 && (
        <div
          style={{
            height: '32px',
            padding: '0 12px',
            backgroundColor: 'var(--bg-chrome)',
            borderTop: '1px solid var(--border-subtle)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            fontSize: '0.725rem',
            color: 'var(--text-secondary)',
            fontFamily: 'var(--font-mono)',
            flexShrink: 0,
          }}
        >
          <div>
            Showing {(currentPage - 1) * pageSize + 1}–{Math.min(currentPage * pageSize, filtered.length)} of {filtered.length} records
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
            <button
              type="button"
              className="btn btn-ghost btn-sm"
              disabled={currentPage === 1}
              onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
              aria-label="Previous page"
            >
              <ChevronLeft size={12} />
            </button>
            <span>
              Page {currentPage} of {totalPages}
            </span>
            <button
              type="button"
              className="btn btn-ghost btn-sm"
              disabled={currentPage === totalPages}
              onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
              aria-label="Next page"
            >
              <ChevronRight size={12} />
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
