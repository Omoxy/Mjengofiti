import React, { useState } from 'react';
import type { Building, BuildingDetail } from '../../api/client';
import { useInspections } from '../../hooks';
import { CheckCircle2, AlertCircle, RefreshCw } from 'lucide-react';

interface NewInspectionFormProps {
  selectedBuilding?: Building | BuildingDetail | null;
  onSuccess?: (caseId: string) => void;
  onSelectBuildingPrompt?: () => void;
}

export const NewInspectionForm: React.FC<NewInspectionFormProps> = ({
  selectedBuilding,
  onSuccess,
  onSelectBuildingPrompt,
}) => {
  const { createCase, isSubmitting } = useInspections('all');

  const [buildingId, setBuildingId] = useState(selectedBuilding?.id || '');
  const [title, setTitle] = useState(
    selectedBuilding
      ? `Field Inspection for ${selectedBuilding.external_id || selectedBuilding.id.substring(0, 8)}`
      : ''
  );
  const [priority, setPriority] = useState<'low' | 'medium' | 'high' | 'critical'>('high');
  const [inspector, setInspector] = useState('lead.inspector@nairobi.go.ke');
  const [dueDate, setDueDate] = useState('');
  const [referralReason, setReferralReason] = useState(
    'Field structural screening and verification per County Physical Planning regulations.'
  );
  const [observations, setObservations] = useState('');

  const [validationError, setValidationError] = useState<string | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [createdCaseId, setCreatedCaseId] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setValidationError(null);
    setSubmitError(null);

    const targetBuildingId = selectedBuilding?.id || buildingId.trim();
    if (!targetBuildingId) {
      setValidationError('Please select or specify a verified building record before submitting.');
      return;
    }

    if (!title.trim()) {
      setValidationError('Inspection title is required.');
      return;
    }

    if (!referralReason.trim()) {
      setValidationError('Referral reason and justification are required.');
      return;
    }

    try {
      const created = await createCase({
        building_id: targetBuildingId,
        title: title.trim(),
        description: observations.trim() || referralReason.trim(),
        priority,
        referral_reason: referralReason.trim(),
        assigned_to: inspector.trim() || undefined,
        due_date: dueDate ? new Date(dueDate).toISOString() : undefined,
        created_by: 'inspector.workspace@nairobi.go.ke',
      });

      setCreatedCaseId(created.id);
      if (onSuccess) {
        onSuccess(created.id);
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Unable to save inspection record.';
      setSubmitError(msg);
    }
  };

  const handleReset = () => {
    setTitle('');
    setBuildingId('');
    setPriority('high');
    setObservations('');
    setReferralReason('Field structural screening and verification per County Physical Planning regulations.');
    setValidationError(null);
    setSubmitError(null);
    setCreatedCaseId(null);
  };

  if (createdCaseId) {
    return (
      <div className="form-card" role="region" aria-label="Inspection submission success">
        <div style={{ textAlign: 'center', padding: '24px 12px' }}>
          <div className="empty-state-icon" style={{ backgroundColor: 'var(--brand-cream)', color: 'var(--brand-maroon)' }}>
            <CheckCircle2 size={28} />
          </div>
          <h2>Inspection record saved successfully</h2>
          <p>
            Case reference <b>{createdCaseId}</b> has been recorded in the inspection queue.
          </p>
          <div style={{ display: 'flex', gap: '10px', justifyContent: 'center', marginTop: '16px' }}>
            <button
              type="button"
              className="mjengo-btn"
              onClick={handleReset}
            >
              Record another inspection
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="form-card" role="region" aria-label="New inspection entry form">
      <h2>Inspection details</h2>
      <p>
        Capture observations for a verified building record. Required fields and criteria follow the approved county protocol.
      </p>

      {/* Guidance when no building is selected */}
      {!selectedBuilding && (
        <div className="form-status-banner">
          <b>Guidance:</b> No building is currently highlighted. Select a building on the map or input a known building identifier below.
          {onSelectBuildingPrompt && (
            <div style={{ marginTop: '8px' }}>
              <button
                type="button"
                className="mjengo-btn alt mjengo-btn-sm"
                onClick={onSelectBuildingPrompt}
              >
                Choose building from map
              </button>
            </div>
          )}
        </div>
      )}

      {validationError && (
        <div
          className="form-status-banner"
          style={{ backgroundColor: 'var(--brand-maroon-soft)', borderColor: 'var(--gis-crimson-border)', color: 'var(--brand-maroon-deep)' }}
          role="alert"
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <AlertCircle size={16} />
            <span>{validationError}</span>
          </div>
        </div>
      )}

      {submitError && (
        <div
          className="form-status-banner"
          style={{ backgroundColor: 'var(--brand-maroon-soft)', borderColor: 'var(--gis-crimson-border)', color: 'var(--brand-maroon-deep)' }}
          role="alert"
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <AlertCircle size={16} />
            <span>Failed to save inspection: {submitError}</span>
          </div>
          <div style={{ marginTop: '8px' }}>
            <button
              type="button"
              className="mjengo-btn alt mjengo-btn-sm"
              onClick={handleSubmit}
              disabled={isSubmitting}
            >
              <RefreshCw size={12} />
              <span>Retry saving</span>
            </button>
          </div>
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div className="form-grid-2col">
          {/* Building Record Field */}
          <div className="form-field full-span">
            <label htmlFor="ins-building-id">Building record</label>
            {selectedBuilding ? (
              <input
                id="ins-building-id"
                type="text"
                readOnly
                value={`${selectedBuilding.external_id || selectedBuilding.id} (${selectedBuilding.source_name || 'verified source'})`}
                style={{ backgroundColor: 'var(--brand-cream-light)' }}
              />
            ) : (
              <input
                id="ins-building-id"
                type="text"
                placeholder="Enter building UUID or external ID"
                value={buildingId}
                onChange={(e) => setBuildingId(e.target.value)}
              />
            )}
          </div>

          {/* Title */}
          <div className="form-field full-span">
            <label htmlFor="ins-title">Inspection title</label>
            <input
              id="ins-title"
              type="text"
              placeholder="Inspection summary or site description"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              required
            />
          </div>

          {/* Priority */}
          <div className="form-field">
            <label htmlFor="ins-priority">Priority</label>
            <select
              id="ins-priority"
              value={priority}
              onChange={(e) => setPriority(e.target.value as any)}
            >
              <option value="critical">Critical</option>
              <option value="high">High</option>
              <option value="medium">Medium</option>
              <option value="low">Low</option>
            </select>
          </div>

          {/* Due Date */}
          <div className="form-field">
            <label htmlFor="ins-due-date">Scheduled completion date</label>
            <input
              id="ins-due-date"
              type="date"
              value={dueDate}
              onChange={(e) => setDueDate(e.target.value)}
            />
          </div>

          {/* Inspector */}
          <div className="form-field full-span">
            <label htmlFor="ins-inspector">Assigned inspector</label>
            <input
              id="ins-inspector"
              type="text"
              value={inspector}
              onChange={(e) => setInspector(e.target.value)}
              placeholder="Inspector email or personnel identifier"
            />
          </div>

          {/* Referral Reason */}
          <div className="form-field full-span">
            <label htmlFor="ins-reason">Referral reason</label>
            <textarea
              id="ins-reason"
              value={referralReason}
              onChange={(e) => setReferralReason(e.target.value)}
              placeholder="State the statutory justification or observation trigger"
              required
            />
          </div>

          {/* Observations */}
          <div className="form-field full-span">
            <label htmlFor="ins-notes">Observations and field notes</label>
            <textarea
              id="ins-notes"
              value={observations}
              onChange={(e) => setObservations(e.target.value)}
              placeholder="Document visible structural conditions or notes"
            />
          </div>
        </div>

        <div className="form-submit-row">
          <button
            type="button"
            className="mjengo-btn alt"
            onClick={handleReset}
            disabled={isSubmitting}
          >
            Clear
          </button>
          <button
            type="submit"
            className="mjengo-btn"
            disabled={isSubmitting}
          >
            {isSubmitting ? 'Saving inspection...' : 'Save inspection'}
          </button>
        </div>
      </form>
    </div>
  );
};
