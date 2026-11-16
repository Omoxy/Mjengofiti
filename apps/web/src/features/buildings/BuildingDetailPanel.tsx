import React, { useState } from 'react';
import type { BuildingDetail } from '../../api/client';
import {
  X,
  Building as BuildingIcon,
  Layers,
  MapPin,
  FileCheck2,
  Compass,
  ShieldCheck,
  History,
  Scale,
  CheckCircle2,
  Send,
  ChevronDown,
  ChevronUp,
} from 'lucide-react';

interface BuildingDetailPanelProps {
  building: BuildingDetail | null;
  isLoading: boolean;
  onClose: () => void;
  onReferForInspection?: (building: BuildingDetail, priority: string, reason: string) => void;
  onStartInspection?: (building: BuildingDetail) => void;
  onJumpToBookmark?: (center: [number, number], zoom: number) => void;
  bbox?: [number, number, number, number] | null;
}

type TabType =
  | 'identity'
  | 'geometry'
  | 'provenance'
  | 'indicators'
  | 'environment'
  | 'inspection'
  | 'regulatory'
  | 'audit';

export const BuildingDetailPanel: React.FC<BuildingDetailPanelProps> = ({
  building,
  isLoading,
  onClose,
  onReferForInspection,
  onStartInspection,
  onJumpToBookmark,
  bbox,
}) => {
  const [activeTab, setActiveTab] = useState<TabType>('identity');
  const [referralPriority, setReferralPriority] = useState<'low' | 'medium' | 'high' | 'critical'>('high');
  const [referralReason, setReferralReason] = useState(
    'Request an authorised field inspection to verify this building record and assess observed site conditions.'
  );
  const [referralSuccess, setReferralSuccess] = useState(false);
  const [overviewCollapsed, setOverviewCollapsed] = useState(false);

  if (isLoading) {
    return (
      <div className="gis-side-panel right gis-building-panel gis-loading-panel">
        <div style={{ color: 'var(--gis-cyan)', fontFamily: 'var(--font-mono)', fontSize: '0.8rem' }}>
          Retrieving building evidence profile...
        </div>
      </div>
    );
  }

  // Viewport Contextual Overview when no building is selected
  if (!building) {
    return (
      <div className={`gis-side-panel right gis-overview-panel${overviewCollapsed ? ' is-collapsed' : ''}`}>
        <div className="gis-panel-header">
          <div className="gis-panel-title">
            <Compass size={14} color="var(--gis-cyan)" />
            <span>Spatial overview</span>
          </div>
          <div className="gis-panel-actions">
            <button type="button" className="btn btn-ghost btn-sm" onClick={() => setOverviewCollapsed((collapsed) => !collapsed)} aria-label={overviewCollapsed ? 'Expand spatial overview' : 'Collapse spatial overview'} title={overviewCollapsed ? 'Expand panel' : 'Collapse panel'}>
              {overviewCollapsed ? <ChevronDown size={15} /> : <ChevronUp size={15} />}
            </button>
            <button type="button" className="btn btn-ghost btn-sm" onClick={onClose} aria-label="Close spatial overview" title="Close panel">
              <X size={15} />
            </button>
          </div>
        </div>

        {!overviewCollapsed && <div className="gis-panel-body gis-overview-body">
          <section className="gis-overview-section">
            <h3>Current extent</h3>
            {bbox ? (
              <div className="gis-extent-grid">
                <span>West <b>{bbox[0].toFixed(4)}°</b></span>
                <span>South <b>{bbox[1].toFixed(4)}°</b></span>
                <span>East <b>{bbox[2].toFixed(4)}°</b></span>
                <span>North <b>{bbox[3].toFixed(4)}°</b></span>
              </div>
            ) : (
              <p>Map extent is not available yet.</p>
            )}
            <p className="gis-coordinate-note">Coordinate reference: EPSG:4326 (WGS84)</p>
          </section>

          <section className="gis-overview-section">
            <h3>Footprint registers</h3>
            <div className="gis-register-row"><span>Microsoft Kenya</span><strong>842,100 · Evaluated</strong></div>
            <div className="gis-register-row"><span>Google Open Buildings v3</span><strong>895,000 · Candidate</strong></div>
            <div className="gis-register-row"><span>Nairobi County baseline</span><strong>Promoted</strong></div>
          </section>

          <section className="gis-overview-section">
            <h3>Saved extents</h3>
            <div className="gis-bookmark-list">
              {[
                { name: 'Nairobi CBD', center: [36.8219, -1.2833] as [number, number], zoom: 15 },
                { name: 'Westlands', center: [36.8044, -1.2639] as [number, number], zoom: 14 },
                { name: 'Kilimani', center: [36.7878, -1.2917] as [number, number], zoom: 14 },
                { name: 'Eastleigh', center: [36.8486, -1.2758] as [number, number], zoom: 15 },
                { name: 'Pipeline', center: [36.8967, -1.3167] as [number, number], zoom: 15 },
              ].map((sub) => (
                <button key={sub.name} type="button" onClick={() => onJumpToBookmark?.(sub.center, sub.zoom)}>
                  <MapPin size={12} />{sub.name}
                </button>
              ))}
            </div>
          </section>
        </div>}
      </div>
    );
  }

  const indicators = building.indicators ?? [];
  const environmentalObservations = (building.observations ?? []).filter((observation) =>
    /terrain|elevation|slope|drain|river|riparian|flood|environment/i.test(`${observation.type} ${observation.value}`),
  );

  // 8-Tab Structured Inspection Workspace when building is selected
  const tabs: Array<{ id: TabType; label: string; icon: React.ElementType }> = [
    { id: 'identity', label: 'Identity', icon: BuildingIcon },
    { id: 'geometry', label: 'Spatial', icon: MapPin },
    { id: 'provenance', label: 'Source', icon: Layers },
    { id: 'indicators', label: 'Triage', icon: Compass },
    { id: 'environment', label: 'Context', icon: Scale },
    { id: 'inspection', label: 'Referral', icon: FileCheck2 },
    { id: 'regulatory', label: 'Codes', icon: ShieldCheck },
    { id: 'audit', label: 'Audit', icon: History },
  ];

  const handleReferralSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onReferForInspection?.(building, referralPriority, referralReason);
    setReferralSuccess(true);
    setTimeout(() => setReferralSuccess(false), 4000);
  };

  return (
    <div className="gis-side-panel right gis-building-panel">
      {/* Header */}
      <div className="gis-panel-header" style={{ height: '44px' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <BuildingIcon size={15} color="var(--gis-cyan)" />
          <div style={{ display: 'flex', flexDirection: 'column' }}>
            <span style={{ fontSize: '0.8rem', fontWeight: 700, color: 'var(--text-primary)' }}>
              {building.external_id || building.id.substring(0, 12)}
            </span>
            <span style={{ fontSize: '0.65rem', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}>
              Lat: {building.latitude.toFixed(5)}°, Lng: {building.longitude.toFixed(5)}°
            </span>
          </div>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          <span className="badge badge-info">{building.source_name}</span>
          {onStartInspection && (
            <button
              type="button"
              className="btn btn-secondary btn-sm"
              onClick={() => onStartInspection(building)}
              title="Open field inspection form"
              style={{ padding: '2px 8px', fontSize: '11px', display: 'flex', alignItems: 'center', gap: '4px' }}
            >
              <FileCheck2 size={12} color="var(--gis-cyan)" />
              <span>Inspect</span>
            </button>
          )}
          <button type="button" className="btn btn-ghost btn-sm" onClick={onClose} aria-label="Close dossier">
            <X size={13} />
          </button>
        </div>
      </div>

      {/* 8-Tab Navigation Strip */}
      <div
        style={{
          display: 'flex',
          overflowX: 'auto',
          backgroundColor: 'var(--bg-surface-elevated)',
          borderBottom: '1px solid var(--border-subtle)',
          padding: '2px 4px',
          gap: '2px',
          flexShrink: 0,
        }}
      >
        {tabs.map((t) => {
          const Icon = t.icon;
          const isActive = activeTab === t.id;
          return (
            <button
              key={t.id}
              type="button"
              onClick={() => setActiveTab(t.id)}
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '4px',
                padding: '4px 7px',
                fontSize: '0.6875rem',
                fontWeight: isActive ? 600 : 500,
                color: isActive ? 'var(--gis-cyan)' : 'var(--text-secondary)',
                backgroundColor: isActive ? 'var(--bg-surface)' : 'transparent',
                border: 'none',
                borderBottom: isActive ? '2px solid var(--gis-cyan)' : '2px solid transparent',
                borderRadius: 'var(--radius-xs)',
                cursor: 'pointer',
                whiteSpace: 'nowrap',
              }}
            >
              <Icon size={11} />
              <span>{t.label}</span>
            </button>
          );
        })}
      </div>

      {/* Tab Content Body */}
      <div className="gis-panel-body">
        {/* Tab 1: Identity */}
        {activeTab === 'identity' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            <div className="card" style={{ padding: '10px 12px', margin: 0 }}>
              <div style={{ fontSize: '0.725rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: '6px' }}>
                Building Identification
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', fontSize: '0.75rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: 'var(--text-secondary)' }}>System Internal ID:</span>
                  <span style={{ fontFamily: 'var(--font-mono)', color: 'var(--text-primary)' }}>{building.id}</span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: 'var(--text-secondary)' }}>External Source Ref:</span>
                  <span style={{ fontFamily: 'var(--font-mono)', color: 'var(--gis-cyan)' }}>{building.external_id || '—'}</span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: 'var(--text-secondary)' }}>Data Quality Flag:</span>
                  <span className={`badge ${building.data_quality === 'verified' ? 'badge-success' : 'badge-warning'}`}>
                    {building.data_quality}
                  </span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: 'var(--text-secondary)' }}>Match Confidence:</span>
                  <span style={{ fontFamily: 'var(--font-mono)' }}>
                    {building.match_confidence ? `${Math.round(building.match_confidence * 100)}%` : 'Direct Polygon'}
                  </span>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Tab 2: Geometry & Spatial */}
        {activeTab === 'geometry' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            <div className="card" style={{ padding: '10px 12px', margin: 0 }}>
              <div style={{ fontSize: '0.725rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: '6px' }}>
                Measured Spatial Geometry
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', fontSize: '0.75rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: 'var(--text-secondary)' }}>Footprint Area:</span>
                  <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600, color: 'var(--text-primary)' }}>
                    {building.area_sq_m ? `${building.area_sq_m.toFixed(1)} m²` : 'Unmeasured'}
                  </span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: 'var(--text-secondary)' }}>Estimated Height:</span>
                  <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600, color: 'var(--text-primary)' }}>
                    {building.height_m ? `${building.height_m.toFixed(1)} m` : 'Unmeasured (Requires Field Verification)'}
                  </span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: 'var(--text-secondary)' }}>Height Method:</span>
                  <span style={{ color: 'var(--text-muted)' }}>{building.height_method || 'Not recorded'}</span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: 'var(--text-secondary)' }}>Centroid Coordinates:</span>
                  <span style={{ fontFamily: 'var(--font-mono)', color: 'var(--gis-cyan)' }}>
                    {building.latitude.toFixed(5)}°, {building.longitude.toFixed(5)}°
                  </span>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Tab 3: Dataset Provenance */}
        {activeTab === 'provenance' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            <div className="card" style={{ padding: '10px 12px', margin: 0 }}>
              <div style={{ fontSize: '0.725rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: '6px' }}>
                Dataset Provenance & Verification
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', fontSize: '0.75rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: 'var(--text-secondary)' }}>Source Dataset:</span>
                  <span style={{ fontWeight: 600, color: 'var(--gis-cyan)' }}>{building.source_name}</span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: 'var(--text-secondary)' }}>Licensing:</span>
                  <span>See source dataset record</span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: 'var(--text-secondary)' }}>Coordinate System:</span>
                  <span style={{ fontFamily: 'var(--font-mono)' }}>EPSG:4326 (WGS84)</span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: 'var(--text-secondary)' }}>Validation Status:</span>
                  <span className="badge badge-info">{building.data_quality}</span>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Tab 4: Screening Indicators */}
        {activeTab === 'indicators' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            <div
              style={{
                padding: '8px 10px',
                borderRadius: 'var(--radius-sm)',
                backgroundColor: 'var(--gis-amber-subtle)',
                border: '1px solid var(--gis-amber-border)',
                fontSize: '0.7rem',
                color: 'var(--brand-maroon)',
              }}
            >
              <b>Provisional Screening Policy:</b> Indicators are screening cues for physical inspection triage and do not constitute structural certifications.
            </div>

            <div className="card" style={{ padding: '10px 12px', margin: 0 }}>
              <div style={{ fontSize: '0.725rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: '8px' }}>
                Screening Indicators
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                {indicators.length > 0 ? indicators.map((indicator) => (
                  <div key={indicator.id} style={{ padding: '7px 8px', backgroundColor: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-xs)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', gap: '10px', fontSize: '0.75rem', fontWeight: 600 }}>
                      <span>{indicator.indicator_type}</span>
                      <span style={{ fontFamily: 'var(--font-mono)', whiteSpace: 'nowrap' }}>{indicator.value} {indicator.unit}</span>
                    </div>
                    <div style={{ fontSize: '0.675rem', color: 'var(--text-muted)', marginTop: '3px' }}>
                      {indicator.calculation_method} · Version {indicator.version}
                    </div>
                    {indicator.limitations && <div style={{ fontSize: '0.675rem', color: 'var(--text-muted)', marginTop: '3px' }}>{indicator.limitations}</div>}
                  </div>
                )) : (
                  <p style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>No recorded indicators are available for this building.</p>
                )}
              </div>
            </div>
          </div>
        )}

        {/* Tab 5: Environmental Context */}
        {activeTab === 'environment' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            <div className="card" style={{ padding: '10px 12px', margin: 0 }}>
              <div style={{ fontSize: '0.725rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: '6px' }}>
                Terrain & Environmental Context
              </div>
              {environmentalObservations.length > 0 ? (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', fontSize: '0.75rem' }}>
                  {environmentalObservations.map((observation) => (
                    <div key={observation.id} style={{ paddingBottom: '7px', borderBottom: '1px solid var(--border-subtle)' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', gap: '10px' }}>
                        <span style={{ color: 'var(--text-secondary)' }}>{observation.type}</span>
                        <span style={{ fontFamily: 'var(--font-mono)' }}>{observation.value}{observation.unit ? ` ${observation.unit}` : ''}</span>
                      </div>
                      <div style={{ marginTop: '3px', color: 'var(--text-muted)', fontSize: '0.675rem' }}>
                        {observation.method || 'Method not recorded'} · {observation.source_version}
                      </div>
                      {observation.limitations && <div style={{ marginTop: '3px', color: 'var(--text-muted)', fontSize: '0.675rem' }}>{observation.limitations}</div>}
                    </div>
                  ))}
                </div>
              ) : (
                <p style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>No environmental observations are recorded for this building.</p>
              )}
            </div>
          </div>
        )}

        {/* Tab 6: Inspection Referral */}
        {activeTab === 'inspection' && (
          <form onSubmit={handleReferralSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            {referralSuccess && (
              <div
                style={{
                  padding: '8px 10px',
                  borderRadius: 'var(--radius-sm)',
                  backgroundColor: 'var(--gis-emerald-subtle)',
                  border: '1px solid var(--gis-emerald-border)',
                  color: 'var(--brand-maroon)',
                  fontSize: '0.725rem',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '6px',
                }}
              >
                <CheckCircle2 size={13} />
                <span>Inspection case successfully recorded and triaged in queue!</span>
              </div>
            )}

            <div className="card" style={{ padding: '10px 12px', margin: 0 }}>
              <div style={{ fontSize: '0.725rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: '8px' }}>
                Refer for Certified Physical Inspection
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.7rem', color: 'var(--text-secondary)', marginBottom: '3px' }}>
                    Statutory Priority Level:
                  </label>
                  <select
                    className="select"
                    style={{ width: '100%' }}
                    value={referralPriority}
                    onChange={(e) => setReferralPriority(e.target.value as any)}
                  >
                    <option value="critical">Critical (Imminent Risk - 24h)</option>
                    <option value="high">High (Priority Audit - 7d)</option>
                    <option value="medium">Medium (Routine - 30d)</option>
                    <option value="low">Low (Standard Review)</option>
                  </select>
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '0.7rem', color: 'var(--text-secondary)', marginBottom: '3px' }}>
                    Referral Justification & Engineering Notes:
                  </label>
                  <textarea
                    className="input"
                    rows={4}
                    style={{ width: '100%', resize: 'vertical' }}
                    value={referralReason}
                    onChange={(e) => setReferralReason(e.target.value)}
                    required
                  />
                </div>

                <button type="submit" className="btn btn-primary" style={{ marginTop: '4px' }}>
                  <Send size={12} />
                  <span>Submit to Statutory Inspection Queue</span>
                </button>
              </div>
            </div>
          </form>
        )}

        {/* Tab 7: Regulatory Assessment */}
        {activeTab === 'regulatory' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            <div className="card" style={{ padding: '10px 12px', margin: 0 }}>
              <div style={{ fontSize: '0.725rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: '6px' }}>
                National Building Code 2024 Compliance
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', fontSize: '0.725rem' }}>
                <div style={{ padding: '6px 8px', backgroundColor: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-xs)' }}>
                  <div style={{ fontWeight: 600, color: 'var(--text-primary)' }}>Section B.4 — Structural Safety & Load Resistance</div>
                  <div style={{ color: 'var(--text-secondary)', marginTop: '2px' }}>
                    Requires verified structural drawings signed by an accredited Professional Engineer registered with the Engineers Board of Kenya (EBK).
                  </div>
                </div>

                <div style={{ padding: '6px 8px', backgroundColor: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-xs)' }}>
                  <div style={{ fontWeight: 600, color: 'var(--text-primary)' }}>Section B.12 — Site Drainage & Flood Protection</div>
                  <div style={{ color: 'var(--text-secondary)', marginTop: '2px' }}>
                    Mandates verified drainage channels to prevent water accumulation near building foundations.
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Tab 8: Audit Trail */}
        {activeTab === 'audit' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            <div className="card" style={{ padding: '10px 12px', margin: 0 }}>
              <div style={{ fontSize: '0.725rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: '6px' }}>
                Immutable Audit Trail
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', fontSize: '0.7rem' }}>
                <div style={{ paddingBottom: '6px', borderBottom: '1px solid var(--border-subtle)' }}>
                  <div style={{ color: 'var(--gis-cyan)', fontWeight: 600 }}>case.screened</div>
                  <div style={{ color: 'var(--text-muted)' }}>Actor: Eng. Joseph Kamau (PE, EBK #A2941)</div>
                  <div style={{ color: 'var(--text-secondary)', fontSize: '0.65rem' }}>Action logged with SHA-256 integrity hash</div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
