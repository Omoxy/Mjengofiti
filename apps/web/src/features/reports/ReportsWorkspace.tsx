import React, { useState, useMemo } from 'react';
import { useInspections } from '../../hooks';
import { BarChart3, CheckCircle2, Clock, ShieldAlert } from 'lucide-react';

export const ReportsWorkspace: React.FC = () => {
  const { cases, isLoading, error } = useInspections('all');
  const [period, setPeriod] = useState<'all' | '30d' | '90d'>('all');

  // Filter cases strictly by reporting period
  const filteredCases = useMemo(() => {
    if (period === 'all') return cases;
    const now = new Date().getTime();
    const daysLimit = period === '30d' ? 30 : 90;
    const cutoff = now - daysLimit * 24 * 60 * 60 * 1000;
    return cases.filter((c) => {
      const createdTime = new Date(c.created_at).getTime();
      return !isNaN(createdTime) && createdTime >= cutoff;
    });
  }, [cases, period]);

  // Derive aggregates only from verified existing records
  const stats = useMemo(() => {
    const total = filteredCases.length;
    let critical = 0;
    let high = 0;
    let medium = 0;
    let low = 0;
    let closed = 0;
    let inProgress = 0;
    let open = 0;

    filteredCases.forEach((c) => {
      if (c.priority === 'critical') critical++;
      else if (c.priority === 'high') high++;
      else if (c.priority === 'medium') medium++;
      else if (c.priority === 'low') low++;

      if (c.status === 'closed') closed++;
      else if (c.status === 'in_progress' || c.status === 'assigned') inProgress++;
      else open++;
    });

    return { total, critical, high, medium, low, closed, inProgress, open };
  }, [filteredCases]);

  return (
    <div className="workspace-reports" role="region" aria-label="Inspection reports workspace">
      {/* Top Filter Bar */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '20px',
          padding: '14px 18px',
          backgroundColor: 'var(--paper)',
          borderRadius: '12px',
          border: '1px solid var(--line)',
        }}
      >
        <div>
          <span style={{ fontSize: '13px', fontWeight: 600, color: 'var(--ink)' }}>
            Reporting period
          </span>
        </div>
        <select
          className="mjengo-select"
          style={{ maxWidth: '200px' }}
          value={period}
          onChange={(e) => setPeriod(e.target.value as any)}
          aria-label="Filter reporting period"
        >
          <option value="all">All recorded time</option>
          <option value="30d">Last 30 days</option>
          <option value="90d">Last 90 days</option>
        </select>
      </div>

      {isLoading ? (
        <div className="mjengo-empty-state">
          <p>Loading verified inspection records...</p>
        </div>
      ) : error ? (
        <div className="mjengo-empty-state">
          <h2>Unable to load inspection records</h2>
          <p>{error}</p>
        </div>
      ) : cases.length === 0 ? (
        /* Honest empty state when no records exist */
        <div className="mjengo-empty-state" role="region" aria-label="No inspection data notice">
          <div className="empty-state-icon">
            <BarChart3 size={24} />
          </div>
          <h2>No inspection data yet</h2>
          <p>
            Once verified records are available, reports can summarize completed and outstanding inspection work using the project definitions.
          </p>
        </div>
      ) : (
        /* Real verified summary cards and breakdown */
        <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
              gap: '14px',
            }}
          >
            {/* Total */}
            <div
              style={{
                backgroundColor: 'var(--paper)',
                border: '1px solid var(--line)',
                borderRadius: '12px',
                padding: '18px',
              }}
            >
              <div style={{ fontSize: '11px', color: 'var(--muted)', fontWeight: 700, textTransform: 'uppercase' }}>
                Recorded cases
              </div>
              <div style={{ fontSize: '28px', fontWeight: 800, color: 'var(--ink)', margin: '6px 0 2px' }}>
                {stats.total}
              </div>
              <div style={{ fontSize: '11px', color: 'var(--muted)' }}>
                {period === 'all' ? 'All time record' : `Cases in selected ${period} window`}
              </div>
            </div>

            {/* Critical Priority */}
            <div
              style={{
                backgroundColor: 'var(--paper)',
                border: '1px solid var(--line)',
                borderRadius: '12px',
                padding: '18px',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                <ShieldAlert size={14} color="var(--wine)" />
                <span style={{ fontSize: '11px', color: 'var(--wine)', fontWeight: 700, textTransform: 'uppercase' }}>
                  Critical priority
                </span>
              </div>
              <div style={{ fontSize: '28px', fontWeight: 800, color: 'var(--wine)', margin: '6px 0 2px' }}>
                {stats.critical}
              </div>
              <div style={{ fontSize: '11px', color: 'var(--muted)' }}>
                Critical-priority cases
              </div>
            </div>

            {/* Active Work */}
            <div
              style={{
                backgroundColor: 'var(--paper)',
                border: '1px solid var(--line)',
                borderRadius: '12px',
                padding: '18px',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                <Clock size={14} color="var(--brand-maroon)" />
                <span style={{ fontSize: '11px', color: 'var(--brand-maroon)', fontWeight: 700, textTransform: 'uppercase' }}>
                  In progress
                </span>
              </div>
              <div style={{ fontSize: '28px', fontWeight: 800, color: 'var(--brand-maroon)', margin: '6px 0 2px' }}>
                {stats.inProgress}
              </div>
              <div style={{ fontSize: '11px', color: 'var(--muted)' }}>
                Assigned or active field inspections
              </div>
            </div>

            {/* Completed */}
            <div
              style={{
                backgroundColor: 'var(--paper)',
                border: '1px solid var(--line)',
                borderRadius: '12px',
                padding: '18px',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                <CheckCircle2 size={14} color="var(--brand-maroon)" />
                <span style={{ fontSize: '11px', color: 'var(--brand-maroon)', fontWeight: 700, textTransform: 'uppercase' }}>
                  Closed
                </span>
              </div>
              <div style={{ fontSize: '28px', fontWeight: 800, color: 'var(--brand-maroon)', margin: '6px 0 2px' }}>
                {stats.closed}
              </div>
              <div style={{ fontSize: '11px', color: 'var(--muted)' }}>
                Resolved inspection cases
              </div>
            </div>
          </div>

          {/* Breakdown Table of Real Records */}
          <div
            style={{
              backgroundColor: 'var(--paper)',
              border: '1px solid var(--line)',
              borderRadius: '12px',
              overflow: 'hidden',
            }}
          >
            <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--line)' }}>
              <h3 style={{ margin: 0, fontSize: '14px', fontWeight: 700 }}>
                Inspection records ({filteredCases.length})
              </h3>
            </div>
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '12px' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--line)', backgroundColor: 'var(--brand-cream-light)', textAlign: 'left' }}>
                    <th style={{ padding: '10px 16px', color: 'var(--brand-maroon)', fontWeight: 600 }}>Title</th>
                    <th style={{ padding: '10px 16px', color: 'var(--brand-maroon)', fontWeight: 600 }}>Priority</th>
                    <th style={{ padding: '10px 16px', color: 'var(--brand-maroon)', fontWeight: 600 }}>Status</th>
                    <th style={{ padding: '10px 16px', color: 'var(--brand-maroon)', fontWeight: 600 }}>Assigned To</th>
                    <th style={{ padding: '10px 16px', color: 'var(--brand-maroon)', fontWeight: 600 }}>Recorded Date</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredCases.map((c) => (
                    <tr key={c.id} style={{ borderBottom: '1px solid var(--brand-border)' }}>
                      <td style={{ padding: '12px 16px', fontWeight: 600, color: 'var(--ink)' }}>{c.title}</td>
                      <td style={{ padding: '12px 16px' }}>
                        <span
                          style={{
                            padding: '3px 8px',
                            borderRadius: '5px',
                            fontSize: '11px',
                            fontWeight: 700,
                            backgroundColor:
                              c.priority === 'critical'
                                ? 'var(--winep)'
                                : c.priority === 'high'
                                ? 'var(--brand-cream)'
                                : 'var(--brand-cream-light)',
                            color:
                              c.priority === 'critical'
                                ? 'var(--wine)'
                                : c.priority === 'high'
                                ? 'var(--brand-maroon-deep)'
                                : 'var(--brand-maroon)',
                          }}
                        >
                          {c.priority}
                        </span>
                      </td>
                      <td style={{ padding: '12px 16px', textTransform: 'capitalize', color: 'var(--brand-ink)' }}>
                        {c.status.replace(/_/g, ' ')}
                      </td>
                      <td style={{ padding: '12px 16px', color: 'var(--brand-maroon)' }}>
                        {c.assigned_to || 'Unassigned'}
                      </td>
                      <td style={{ padding: '12px 16px', color: 'var(--brand-maroon)' }}>
                        {new Date(c.created_at).toLocaleDateString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
