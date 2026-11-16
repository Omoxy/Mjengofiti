import { describe, it, expect } from 'vitest';
import { api } from '../api/client';

describe('API Client Structure', () => {
  it('exposes all required workspace query and mutation methods', () => {
    expect(typeof api.getMapConfig).toBe('function');
    expect(typeof api.getBoundaries).toBe('function');
    expect(typeof api.getCoverage).toBe('function');
    expect(typeof api.getBuildings).toBe('function');
    expect(typeof api.getBuildingDetail).toBe('function');
    expect(typeof api.getDatasets).toBe('function');
    expect(typeof api.getDataset).toBe('function');
    expect(typeof api.getDatasetVersions).toBe('function');
    expect(typeof api.getIndicators).toBe('function');
    expect(typeof api.getIndicator).toBe('function');
    expect(typeof api.triggerAnalysisJob).toBe('function');
    expect(typeof api.getInspectionCases).toBe('function');
    expect(typeof api.createInspectionCase).toBe('function');
    expect(typeof api.updateInspectionCaseStatus).toBe('function');
    expect(typeof api.getStandards).toBe('function');
    expect(typeof api.getAuditEvents).toBe('function');
    expect(typeof api.getSystemStatus).toBe('function');
    expect(typeof api.getUserProfile).toBe('function');
    expect(typeof api.getRoles).toBe('function');

    // Ingestion subsystem API endpoints
    expect(typeof api.getIngestionSources).toBe('function');
    expect(typeof api.getIngestionJobs).toBe('function');
    expect(typeof api.createIngestionJob).toBe('function');
    expect(typeof api.validateGeoJSON).toBe('function');
    expect(typeof api.promoteDataset).toBe('function');
  });
});
