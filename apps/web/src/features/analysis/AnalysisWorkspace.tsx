import React, { useState } from 'react';
import {
  Compass,
  ShieldAlert,
} from 'lucide-react';
import { useIndicators } from '../../hooks';

export const AnalysisWorkspace: React.FC = () => {
  const {
    indicators,
    disclaimer,
  } = useIndicators();

  const [selectedId, setSelectedId] = useState<string | null>(null);

  const selectedIndicator =
    indicators.find((i) => i.id === selectedId) || indicators[0] || null;

  React.useEffect(() => {
    if (!selectedId && indicators.length > 0) {
      setSelectedId(indicators[0].id);
    }
  }, [selectedId, indicators]);


  return (
    <div className="workspace-page">
      {/* Workspace Header */}
      <div className="workspace-header">
        <div>
          <h2 className="workspace-title">
            <Compass size={20} color="var(--gis-cyan)" />
            <span>Analysis & Methodology Workspace</span>
          </h2>
          <p className="workspace-subtitle">
            Candidate screening method definitions and stated data limitations. No calculation worker or calibrated model is currently implemented.
          </p>
        </div>
      </div>

      {/* Non-Structural Determination Policy Banner */}
      <div
        style={{
          padding: '10px 14px',
          borderRadius: 'var(--radius-sm)',
          backgroundColor: 'var(--gis-amber-subtle)',
          border: '1px solid var(--gis-amber-border)',
          color: 'var(--brand-maroon)',
          fontSize: '0.75rem',
          display: 'flex',
          alignItems: 'flex-start',
          gap: '10px',
          marginBottom: '14px',
        }}
      >
        <ShieldAlert size={16} style={{ flexShrink: 0, marginTop: '1px' }} />
        <div>
          <span style={{ fontWeight: 700 }}>Non-Structural Determination Policy: </span>
          <span>
            {disclaimer ||
              'No indicator values are currently calculated. These method descriptions are not engineering assessments or safety determinations.'}
          </span>
        </div>
      </div>


      {/* Split Content Area */}
      <div style={{ display: 'grid', gridTemplateColumns: '360px 1fr', gap: '14px', flex: 1, minHeight: 0 }}>
        {/* Left Indicator Selection List */}
        <div
          style={{
            backgroundColor: 'var(--bg-surface)',
            border: '1px solid var(--border-subtle)',
            borderRadius: 'var(--radius-sm)',
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden',
          }}
        >
          <div
            style={{
              padding: '8px 12px',
              backgroundColor: 'var(--bg-surface-muted)',
              borderBottom: '1px solid var(--border-subtle)',
              fontSize: '0.725rem',
              fontWeight: 700,
              color: 'var(--text-secondary)',
              textTransform: 'uppercase',
            }}
          >
            Codified Screening Indicators ({indicators.length})
          </div>

          <div style={{ flex: 1, overflowY: 'auto', padding: '6px' }}>
            {indicators.map((ind) => {
              const isSelected = selectedIndicator?.id === ind.id;
              return (
                <div
                  key={ind.id}
                  onClick={() => setSelectedId(ind.id)}
                  style={{
                    padding: '10px 12px',
                    borderRadius: 'var(--radius-xs)',
                    backgroundColor: isSelected ? 'var(--gis-cyan-subtle)' : 'var(--bg-surface-elevated)',
                    border: isSelected ? '1px solid var(--gis-cyan)' : '1px solid var(--border-subtle)',
                    marginBottom: '6px',
                    cursor: 'pointer',
                    transition: 'all 0.15s ease',
                  }}
                >
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '3px' }}>
                    <span style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-primary)' }}>
                      {ind.display_name}
                    </span>
                    <span className="badge badge-info">V{ind.version}</span>
                  </div>
                  <p style={{ fontSize: '0.7rem', color: 'var(--text-secondary)', lineHeight: 1.4, margin: '3px 0' }}>
                    {ind.description}
                  </p>
                  <div style={{ fontSize: '0.65rem', color: 'var(--gis-cyan)', fontFamily: 'var(--font-mono)' }}>
                    Output: {ind.output_unit}
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Right Indicator Technical Dossier */}
        {selectedIndicator ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px', overflowY: 'auto' }}>
            <div className="card" style={{ margin: 0 }}>
              <div className="card-header">
                <div>
                  <h3 className="card-title">{selectedIndicator.display_name}</h3>
                  <div style={{ fontSize: '0.725rem', color: 'var(--text-secondary)', fontFamily: 'var(--font-mono)', marginTop: '2px' }}>
                    Definition: {selectedIndicator.name} • Version: {selectedIndicator.version}
                  </div>
                </div>
                <span className="badge badge-draft">Execution unavailable</span>
              </div>

              {/* Operational Purpose */}
              <div style={{ marginBottom: '14px' }}>
                <div style={{ fontSize: '0.725rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: '4px' }}>
                  Intended Use
                </div>
                <p style={{ fontSize: '0.775rem', color: 'var(--text-primary)', lineHeight: 1.5 }}>
                  {selectedIndicator.purpose}
                </p>
              </div>

              {/* Codified Mathematical Formula */}
              <div style={{ marginBottom: '14px' }}>
                <div style={{ fontSize: '0.725rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: '4px' }}>
                  Proposed Calculation Method
                </div>
                <div
                  style={{
                    padding: '10px 12px',
                    backgroundColor: 'var(--bg-app)',
                    borderRadius: 'var(--radius-xs)',
                    border: '1px solid var(--border-subtle)',
                    fontFamily: 'var(--font-mono)',
                    fontSize: '0.75rem',
                    color: 'var(--gis-cyan)',
                  }}
                >
                  {selectedIndicator.calculation_method}
                </div>
              </div>


              {/* Limitations & Missing-Data Behaviour */}
              <div
                style={{
                  padding: '10px 12px',
                  backgroundColor: 'var(--bg-surface-elevated)',
                  borderRadius: 'var(--radius-xs)',
                  border: '1px solid var(--border-subtle)',
                  fontSize: '0.725rem',
                  display: 'flex',
                  flexDirection: 'column',
                  gap: '6px',
                }}
              >
                <div style={{ fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase' }}>
                  Limitations & Missing-Data Rules
                </div>
                <div>
                  <span style={{ color: 'var(--text-muted)' }}>Missing Input Behaviour: </span>
                  <span style={{ color: 'var(--text-primary)' }}>{selectedIndicator.missing_data_behaviour}</span>
                </div>
                <div>
                  <span style={{ color: 'var(--text-muted)' }}>Known Epistemic Limitations: </span>
                  <span style={{ color: 'var(--text-secondary)' }}>{selectedIndicator.limitations}</span>
                </div>
              </div>
            </div>
          </div>
        ) : (
          <div style={{ padding: '24px', textAlign: 'center', color: 'var(--text-muted)' }}>Select an indicator to view method specification.</div>
        )}
      </div>
    </div>
  );
};
