import React, { useState, useEffect } from 'react';
import type { AuditEvent } from '../../api/client';
import { api } from '../../api/client';
import {
  ShieldCheck,
  BookOpen,
  History,
  Lock,
  Scale,
} from 'lucide-react';
import { useStandards } from '../../hooks';

export const GovernanceWorkspace: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'standards' | 'audit' | 'rbac'>('standards');
  const { standards } = useStandards();
  const [auditEvents, setAuditEvents] = useState<AuditEvent[]>([]);
  const [roles, setRoles] = useState<any[]>([]);

  useEffect(() => {
    async function loadData() {
      try {
        const [audRes, roleRes] = await Promise.all([
          api.getAuditEvents(50).catch(() => ({ events: [], total: 0 })),
          api.getRoles().catch(() => ({ roles: [] })),
        ]);
        setAuditEvents(audRes.events || []);
        setRoles(roleRes.roles || []);
      } catch (err) {
        console.warn('Failed to load governance telemetry:', err);
      }
    }
    loadData();
  }, []);

  return (
    <div className="workspace-page">
      {/* Workspace Header */}
      <div className="workspace-header">
        <div>
          <h2 className="workspace-title">
            <ShieldCheck size={20} color="var(--gis-cyan)" />
            <span>Security & Governance Workspace</span>
          </h2>
          <p className="workspace-subtitle">
            Standards records, audit events, and role permission definitions. Contents depend on verified records and implemented workflows.
          </p>
        </div>
      </div>

      {/* Tabs */}
      <div
        style={{
          display: 'flex',
          gap: '4px',
          borderBottom: '1px solid var(--border-subtle)',
          marginBottom: '14px',
        }}
      >
        <button
          type="button"
          className="btn btn-ghost"
          onClick={() => setActiveTab('standards')}
          style={{
            borderBottom: activeTab === 'standards' ? '2px solid var(--gis-cyan)' : '2px solid transparent',
            color: activeTab === 'standards' ? 'var(--text-primary)' : 'var(--text-muted)',
            fontWeight: activeTab === 'standards' ? 600 : 500,
            borderRadius: 0,
            padding: '6px 14px',
          }}
        >
          <Scale size={13} style={{ marginRight: '6px' }} />
          <span>Standards Register ({standards.length})</span>
        </button>

        <button
          type="button"
          className="btn btn-ghost"
          onClick={() => setActiveTab('audit')}
          style={{
            borderBottom: activeTab === 'audit' ? '2px solid var(--gis-cyan)' : '2px solid transparent',
            color: activeTab === 'audit' ? 'var(--text-primary)' : 'var(--text-muted)',
            fontWeight: activeTab === 'audit' ? 600 : 500,
            borderRadius: 0,
            padding: '6px 14px',
          }}
        >
          <History size={13} style={{ marginRight: '6px' }} />
          <span>Audit events ({auditEvents.length})</span>
        </button>

        <button
          type="button"
          className="btn btn-ghost"
          onClick={() => setActiveTab('rbac')}
          style={{
            borderBottom: activeTab === 'rbac' ? '2px solid var(--gis-cyan)' : '2px solid transparent',
            color: activeTab === 'rbac' ? 'var(--text-primary)' : 'var(--text-muted)',
            fontWeight: activeTab === 'rbac' ? 600 : 500,
            borderRadius: 0,
            padding: '6px 14px',
          }}
        >
          <Lock size={13} style={{ marginRight: '6px' }} />
          <span>RBAC Permissions Matrix ({roles.length})</span>
        </button>
      </div>

      {/* Tab 1: Standards Register */}
      {activeTab === 'standards' && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
          <div
            style={{
              padding: '10px 14px',
              borderRadius: 'var(--radius-sm)',
              backgroundColor: 'var(--gis-cyan-subtle)',
              border: '1px solid var(--gis-cyan-border)',
              color: 'var(--gis-cyan-hover)',
              fontSize: '0.75rem',
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
            }}
          >
            <BookOpen size={15} />
            <span>
              No standards have been independently reviewed and entered in the application register yet. Automated indicators remain provisional screening information, not legal determinations.
            </span>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            {standards.length === 0 && <p role="status">No reviewed standards are available.</p>}
            {standards.map((s) => (
              <div key={s.id} className="card" style={{ margin: 0 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
                  <div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginBottom: '3px' }}>
                      <span className="badge badge-validated">Section {s.clause_ref}</span>
                      <span className="badge badge-info">{s.document_title}</span>
                    </div>
                    <h3 style={{ fontSize: '0.9rem', fontWeight: 600, color: 'var(--text-primary)' }}>
                      {s.clause_title}
                    </h3>
                  </div>
                  <span style={{ fontSize: '0.7rem', color: 'var(--text-muted)' }}>Authority: {s.authority}</span>
                </div>

                <div
                  style={{
                    padding: '8px 12px',
                    backgroundColor: 'var(--bg-app)',
                    borderRadius: 'var(--radius-xs)',
                    border: '1px solid var(--border-subtle)',
                    marginBottom: '10px',
                    fontSize: '0.75rem',
                    color: 'var(--text-primary)',
                    fontStyle: 'italic',
                  }}
                >
                  "{s.requirement}"
                </div>

                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px', fontSize: '0.725rem' }}>
                  <div>
                    <span style={{ color: 'var(--text-muted)', fontWeight: 600 }}>Legal Interpretation:</span>
                    <p style={{ color: 'var(--text-secondary)', marginTop: '2px' }}>{s.interpretation}</p>
                  </div>
                  <div>
                    <span style={{ color: 'var(--gis-cyan)', fontWeight: 600 }}>Software Implementation Rule:</span>
                    <p style={{ color: 'var(--text-primary)', marginTop: '2px' }}>{s.software_rule}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Tab 2: Audit events */}
      {activeTab === 'audit' && (
        <div className="card" style={{ margin: 0 }}>
          <div className="card-header">
            <h3 className="card-title">
              <History size={14} color="var(--gis-cyan)" />
              <span>Audit events</span>
            </h3>
            <span style={{ fontSize: '0.7rem', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}>
              Event persistence is not yet connected to the application workflows.
            </span>
          </div>

          <table className="data-table">
            <thead>
              <tr>
                <th>Timestamp</th>
                <th>Actor</th>
                <th>Role</th>
                <th>Action</th>
                <th>Resource Type</th>
                <th>Resource ID</th>
                <th>IP Address</th>
              </tr>
            </thead>
            <tbody>
              {auditEvents.length === 0 ? (
                <tr><td colSpan={7}>No audit events have been recorded.</td></tr>
              ) : auditEvents.map((ev) => (
                <tr key={ev.id}>
                  <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.7rem' }}>
                    {new Date(ev.created_at).toLocaleString()}
                  </td>
                  <td style={{ fontWeight: 600 }}>{ev.user_email}</td>
                  <td><span className="badge badge-info">{ev.user_role}</span></td>
                  <td><code style={{ color: 'var(--gis-cyan)' }}>{ev.action}</code></td>
                  <td>{ev.resource_type}</td>
                  <td style={{ fontFamily: 'var(--font-mono)' }}>{ev.resource_id}</td>
                  <td style={{ fontFamily: 'var(--font-mono)', color: 'var(--text-muted)' }}>{ev.ip_address}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Tab 3: RBAC Matrix */}
      {activeTab === 'rbac' && (
        <div className="card" style={{ margin: 0 }}>
          <div className="card-header">
            <h3 className="card-title">
              <Lock size={14} color="var(--gis-cyan)" />
              <span>Role-Based Access Control (RBAC) Governance Matrix</span>
            </h3>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            {roles.map((r) => (
              <div
                key={r.role}
                style={{
                  padding: '10px 12px',
                  backgroundColor: 'var(--bg-surface-elevated)',
                  borderRadius: 'var(--radius-xs)',
                  border: '1px solid var(--border-subtle)',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '4px' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <span style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{r.title}</span>
                    <span className="badge badge-validated">{r.role}</span>
                  </div>
                </div>
                <p style={{ fontSize: '0.725rem', color: 'var(--text-secondary)', marginBottom: '8px' }}>
                  {r.description}
                </p>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '4px' }}>
                  {r.permissions.map((p: string) => (
                    <span
                      key={p}
                      style={{
                        padding: '2px 6px',
                        backgroundColor: 'var(--bg-app)',
                        borderRadius: 'var(--radius-xs)',
                        border: '1px solid var(--border-subtle)',
                        fontSize: '0.675rem',
                        fontFamily: 'var(--font-mono)',
                        color: 'var(--gis-cyan)',
                      }}
                    >
                      {p}
                    </span>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
