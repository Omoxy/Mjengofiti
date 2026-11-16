import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  FileCheck2,
  Plus,
  Building,
  CheckCircle,
  CheckCircle2,
  ArrowRight,
  X,
  Send,
} from 'lucide-react';
import { useInspections } from '../../hooks';

export const InspectionsWorkspace: React.FC = () => {
  const navigate = useNavigate();
  const {
    cases,
    statusFilter,
    setStatusFilter,
    isSubmitting,
    createCase,
    updateCaseStatus,
  } = useInspections('all');

  const [selectedCaseId, setSelectedCaseId] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);

  // Form State
  const [newTitle, setNewTitle] = useState('');
  const [newBuildingId, setNewBuildingId] = useState('');
  const [newPriority, setNewPriority] = useState<'low' | 'medium' | 'high' | 'critical'>('high');
  const [newReason, setNewReason] = useState('');
  const [transitionMsg, setTransitionMsg] = useState<string | null>(null);

  const selectedCase =
    cases.find((c) => c.id === selectedCaseId) || cases[0] || null;

  React.useEffect(() => {
    if (!selectedCaseId && cases.length > 0) {
      setSelectedCaseId(cases[0].id);
    }
  }, [selectedCaseId, cases]);

  const handleCreateCase = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTitle || !newBuildingId || !newReason) {
      alert('Please fill out all required fields.');
      return;
    }

    try {
      const created = await createCase({
        title: newTitle,
        building_id: newBuildingId,
        priority: newPriority,
        referral_reason: newReason,
        created_by: 'lead.supervisor@nairobi.go.ke',
      });
      setShowCreateModal(false);
      setNewTitle('');
      setNewBuildingId('');
      setNewReason('');
      setSelectedCaseId(created.id);
    } catch (err: unknown) {
      alert(`Case creation failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
    }
  };

  const handleStatusTransition = async (nextStatus: string) => {
    if (!selectedCase) return;
    setTransitionMsg(null);
    try {
      await updateCaseStatus(selectedCase.id, nextStatus);
      setTransitionMsg(`Case status successfully advanced to "${nextStatus}".`);
      setTimeout(() => setTransitionMsg(null), 4000);
    } catch (err: unknown) {
      alert(`Status transition rejected: ${err instanceof Error ? err.message : 'Invalid state transition'}`);
    }
  };

  const getPriorityBadge = (p: string) => {
    switch (p) {
      case 'critical':
        return <span className="badge badge-critical">Critical</span>;
      case 'high':
        return <span className="badge badge-warning">High</span>;
      case 'medium':
        return <span className="badge badge-info">Medium</span>;
      default:
        return <span className="badge badge-success">Low</span>;
    }
  };

  const getStatusBadge = (s: string) => {
    switch (s) {
      case 'closed':
        return <span className="badge badge-success">Closed</span>;
      case 'in_progress':
        return <span className="badge badge-validated">In Progress</span>;
      case 'assigned':
        return <span className="badge badge-info">Assigned</span>;
      case 'triaged':
        return <span className="badge badge-warning">Triaged</span>;
      default:
        return <span className="badge badge-draft">Open</span>;
    }
  };

  const filterTabs = [
    { id: 'all', label: 'All Cases' },
    { id: 'open', label: 'Open' },
    { id: 'triaged', label: 'Triaged' },
    { id: 'assigned', label: 'Assigned' },
    { id: 'in_progress', label: 'In Progress' },
    { id: 'pending_review', label: 'Pending Review' },
    { id: 'closed', label: 'Closed' },
  ];

  return (
    <div className="workspace-page">
      {/* Workspace Header */}
      <div className="workspace-header">
        <div>
          <h2 className="workspace-title">
            <FileCheck2 size={20} color="var(--gis-cyan)" />
            <span>Building Inspection Case Management</span>
            <span className="badge badge-warning" style={{ marginLeft: '6px' }}>Statutory Audit Queue</span>
          </h2>
          <p className="workspace-subtitle">
            Authorised inspection referral queue, triage records, certified site assignments, and professional engineering audit findings. Enforces strict state machine transitions with an immutable audit log.
          </p>
        </div>

        <button
          type="button"
          className="btn btn-primary"
          onClick={() => setShowCreateModal(true)}
        >
          <Plus size={13} />
          <span>New Inspection Referral</span>
        </button>
      </div>

      {transitionMsg && (
        <div
          style={{
            padding: '8px 12px',
            borderRadius: 'var(--radius-sm)',
            backgroundColor: 'var(--gis-emerald-subtle)',
            border: '1px solid var(--gis-emerald-border)',
            color: 'var(--brand-maroon)',
            fontSize: '0.75rem',
            marginBottom: '12px',
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
          }}
        >
          <CheckCircle2 size={14} />
          <span>{transitionMsg}</span>
        </div>
      )}

      {/* Filter Tabs */}
      <div
        style={{
          display: 'flex',
          gap: '4px',
          borderBottom: '1px solid var(--border-subtle)',
          marginBottom: '14px',
          overflowX: 'auto',
        }}
      >
        {filterTabs.map((t) => (
          <button
            key={t.id}
            type="button"
            className="btn btn-ghost"
            onClick={() => setStatusFilter(t.id)}
            style={{
              borderBottom: statusFilter === t.id ? '2px solid var(--gis-cyan)' : '2px solid transparent',
              color: statusFilter === t.id ? 'var(--text-primary)' : 'var(--text-muted)',
              fontWeight: statusFilter === t.id ? 600 : 500,
              borderRadius: 0,
              padding: '6px 12px',
              whiteSpace: 'nowrap',
            }}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* Split Queue & Detail Layout */}
      <div style={{ display: 'grid', gridTemplateColumns: '400px 1fr', gap: '14px', flex: 1, minHeight: 0 }}>
        {/* Left Case Queue List */}
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
            Inspection Queue ({cases.length})
          </div>

          <div style={{ flex: 1, overflowY: 'auto', padding: '6px' }}>
            {cases.length === 0 ? (
              <div className="mjengo-empty-state" style={{ margin: '24px auto', padding: '24px 16px' }}>
                <div className="empty-state-icon">✓</div>
                <h2>Inspection queue will appear here</h2>
                <p>
                  Connect building records to show eligible work. This preview uses no sample records or outcomes.
                </p>
                <button
                  type="button"
                  className="mjengo-btn"
                  onClick={() => navigate('/inspections/new')}
                >
                  Open field inspection form
                </button>
              </div>
            ) : (
              cases.map((c) => {
                const isSelected = selectedCase?.id === c.id;
                return (
                  <div
                    key={c.id}
                    onClick={() => setSelectedCaseId(c.id)}
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
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '4px' }}>
                      <span style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-primary)' }}>{c.title}</span>
                      {getPriorityBadge(c.priority)}
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '0.7rem', color: 'var(--text-secondary)', marginBottom: '4px' }}>
                      <Building size={11} />
                      <span style={{ fontFamily: 'var(--font-mono)' }}>Bldg: {c.building_id}</span>
                    </div>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.675rem' }}>
                      {getStatusBadge(c.status)}
                      <span style={{ color: 'var(--text-muted)' }}>{new Date(c.created_at).toLocaleDateString()}</span>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </div>

        {/* Right Selected Case Technical Dossier */}
        {selectedCase ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px', overflowY: 'auto' }}>
            <div className="card" style={{ margin: 0 }}>
              <div className="card-header">
                <div>
                  <h3 className="card-title">{selectedCase.title}</h3>
                  <div style={{ fontSize: '0.725rem', color: 'var(--text-secondary)', fontFamily: 'var(--font-mono)', marginTop: '2px' }}>
                    Case ID: {selectedCase.id} • Building: {selectedCase.building_id}
                  </div>
                </div>
                <div style={{ display: 'flex', gap: '6px' }}>
                  {getPriorityBadge(selectedCase.priority)}
                  {getStatusBadge(selectedCase.status)}
                </div>
              </div>

              {/* State Machine Transition Actions */}
              <div
                style={{
                  padding: '10px 12px',
                  backgroundColor: 'var(--bg-surface-elevated)',
                  borderRadius: 'var(--radius-xs)',
                  border: '1px solid var(--border-subtle)',
                  marginBottom: '14px',
                }}
              >
                <div style={{ fontSize: '0.725rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: '8px' }}>
                  Statutory State Machine Actions
                </div>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '6px' }}>
                  {selectedCase.status === 'open' && (
                    <button type="button" className="btn btn-secondary btn-sm" onClick={() => handleStatusTransition('triaged')}>
                      <ArrowRight size={11} /> Advance to Triaged
                    </button>
                  )}
                  {selectedCase.status === 'triaged' && (
                    <button type="button" className="btn btn-secondary btn-sm" onClick={() => handleStatusTransition('assigned')}>
                      <ArrowRight size={11} /> Assign Field Inspector
                    </button>
                  )}
                  {selectedCase.status === 'assigned' && (
                    <button type="button" className="btn btn-secondary btn-sm" onClick={() => handleStatusTransition('in_progress')}>
                      <ArrowRight size={11} /> Mark Inspection in Progress
                    </button>
                  )}
                  {selectedCase.status === 'in_progress' && (
                    <button type="button" className="btn btn-secondary btn-sm" onClick={() => handleStatusTransition('pending_review')}>
                      <ArrowRight size={11} /> Submit Findings for Review
                    </button>
                  )}
                  {selectedCase.status === 'pending_review' && (
                    <button type="button" className="btn btn-primary btn-sm" onClick={() => handleStatusTransition('closed')}>
                      <CheckCircle size={11} /> Close & Seal Case Record
                    </button>
                  )}
                </div>
              </div>

              {/* Referral Reason & Notes */}
              <div style={{ marginBottom: '14px' }}>
                <div style={{ fontSize: '0.725rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: '4px' }}>
                  Referral Rationale & Findings Notes
                </div>
                <p style={{ fontSize: '0.775rem', color: 'var(--text-primary)', lineHeight: 1.5, padding: '8px 10px', backgroundColor: 'var(--bg-app)', borderRadius: 'var(--radius-xs)', border: '1px solid var(--border-subtle)' }}>
                  {selectedCase.referral_reason}
                </p>
              </div>

              {/* Inspector Assignment & Timestamps */}
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '10px' }}>
                <div style={{ padding: '8px 10px', backgroundColor: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-xs)' }}>
                  <div style={{ fontSize: '0.675rem', color: 'var(--text-muted)' }}>Assigned Officer</div>
                  <div style={{ fontSize: '0.775rem', fontWeight: 600, color: 'var(--text-primary)', marginTop: '2px' }}>
                    {selectedCase.assigned_to || 'Unassigned (Awaiting Triage)'}
                  </div>
                </div>

                <div style={{ padding: '8px 10px', backgroundColor: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-xs)' }}>
                  <div style={{ fontSize: '0.675rem', color: 'var(--text-muted)' }}>Created By</div>
                  <div style={{ fontSize: '0.775rem', fontWeight: 600, color: 'var(--text-primary)', marginTop: '2px' }}>
                    {selectedCase.created_by}
                  </div>
                </div>

                <div style={{ padding: '8px 10px', backgroundColor: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-xs)' }}>
                  <div style={{ fontSize: '0.675rem', color: 'var(--text-muted)' }}>Due Date</div>
                  <div style={{ fontSize: '0.775rem', fontWeight: 600, color: 'var(--gis-cyan)', marginTop: '2px' }}>
                    {selectedCase.due_date ? new Date(selectedCase.due_date).toLocaleDateString() : '7 Days from Triage'}
                  </div>
                </div>
              </div>
            </div>
          </div>
        ) : (
          <div style={{ padding: '24px', textAlign: 'center', color: 'var(--text-muted)' }}>Select an inspection case to review details.</div>
        )}
      </div>

      {/* New Referral Modal */}
      {showCreateModal && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            backgroundColor: 'rgba(84, 23, 43, 0.28)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 1000,
            padding: '20px',
          }}
        >
          <div
            style={{
              width: '100%',
              maxWidth: '540px',
              backgroundColor: 'var(--bg-surface)',
              border: '1px solid var(--border-subtle)',
              borderRadius: 'var(--radius-md)',
              boxShadow: 'var(--shadow-lg)',
              display: 'flex',
              flexDirection: 'column',
            }}
          >
            <div className="gis-panel-header" style={{ height: '42px' }}>
              <div className="gis-panel-title">
                <Plus size={14} color="var(--gis-cyan)" />
                <span>New Statutory Inspection Referral</span>
              </div>
              <button type="button" className="btn btn-ghost btn-sm" onClick={() => setShowCreateModal(false)}>
                <X size={14} />
              </button>
            </div>

            <form onSubmit={handleCreateCase} style={{ padding: '16px', display: 'flex', flexDirection: 'column', gap: '12px' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.725rem', color: 'var(--text-secondary)', marginBottom: '4px' }}>
                  Case Title:
                </label>
                <input
                  type="text"
                  className="input"
                  style={{ width: '100%' }}
                  placeholder="e.g. Slenderness & Overcrowding Audit - Pipeline Plot 42"
                  value={newTitle}
                  onChange={(e) => setNewTitle(e.target.value)}
                  required
                />
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.725rem', color: 'var(--text-secondary)', marginBottom: '4px' }}>
                  Target Building ID / Footprint Ref:
                </label>
                <input
                  type="text"
                  className="input"
                  style={{ width: '100%' }}
                  placeholder="e.g. bldg-nbo-pipeline-00842"
                  value={newBuildingId}
                  onChange={(e) => setNewBuildingId(e.target.value)}
                  required
                />
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.725rem', color: 'var(--text-secondary)', marginBottom: '4px' }}>
                  Priority Level:
                </label>
                <select
                  className="select"
                  style={{ width: '100%' }}
                  value={newPriority}
                  onChange={(e) => setNewPriority(e.target.value as any)}
                >
                  <option value="critical">Critical (Imminent Risk - 24h)</option>
                  <option value="high">High (Priority Audit - 7d)</option>
                  <option value="medium">Medium (Routine - 30d)</option>
                  <option value="low">Low (Standard Review)</option>
                </select>
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.725rem', color: 'var(--text-secondary)', marginBottom: '4px' }}>
                  Statutory Rationale & Engineering Notes:
                </label>
                <textarea
                  className="input"
                  rows={4}
                  style={{ width: '100%', resize: 'vertical' }}
                  placeholder="Explain why this building is being referred under the National Construction Authority Act..."
                  value={newReason}
                  onChange={(e) => setNewReason(e.target.value)}
                  required
                />
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '6px' }}>
                <button type="button" className="btn btn-secondary" onClick={() => setShowCreateModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary" disabled={isSubmitting}>
                  <Send size={12} />
                  <span>Submit Referral</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
