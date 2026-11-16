import { describe, it, expect } from 'vitest';
import * as Components from '../components';
import * as Hooks from '../hooks';
import { NAIROBI_CENTER, NAIROBI_DEFAULT_ZOOM } from '../hooks/useMapViewport';

describe('Shared Components Exports', () => {
  it('exports all standard layout, feedback, form, and data-display components', () => {
    expect(Components.WorkspaceShell).toBeDefined();
    expect(Components.LoadingSpinner).toBeDefined();
    expect(Components.ErrorAlert).toBeDefined();
    expect(Components.EmptyState).toBeDefined();
    expect(Components.Button).toBeDefined();
    expect(Components.SearchInput).toBeDefined();
    expect(Components.Select).toBeDefined();
    expect(Components.Badge).toBeDefined();
    expect(Components.Card).toBeDefined();
    expect(Components.DataTable).toBeDefined();
    expect(Components.Modal).toBeDefined();
  });
});

describe('Shared Hooks Exports & Logic', () => {
  it('exports all domain and operational hooks', () => {
    expect(typeof Hooks.useAuth).toBe('function');
    expect(typeof Hooks.useSystemStatus).toBe('function');
    expect(typeof Hooks.useBuildings).toBe('function');
    expect(typeof Hooks.useDatasets).toBe('function');
    expect(typeof Hooks.useInspections).toBe('function');
    expect(typeof Hooks.useIndicators).toBe('function');
    expect(typeof Hooks.useStandards).toBe('function');
    expect(typeof Hooks.useMapViewport).toBe('function');
  });

  it('defines valid default Nairobi geospatial coordinates for map viewport', () => {
    expect(NAIROBI_CENTER).toHaveLength(2);
    expect(NAIROBI_CENTER[0]).toBeCloseTo(36.82, 1); // Longitude
    expect(NAIROBI_CENTER[1]).toBeCloseTo(-1.29, 1); // Latitude
    expect(NAIROBI_DEFAULT_ZOOM).toBe(12);
  });
});
