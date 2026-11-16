import { describe, it, expect } from 'vitest';
import { LandingPage } from '../features/landing/LandingPage';
import { WorkspaceShell } from '../components/layout/WorkspaceShell';
import { NewInspectionForm } from '../features/inspections/NewInspectionForm';
import { ReportsWorkspace } from '../features/reports/ReportsWorkspace';
import { BuildingRecordsWorkspace } from '../features/buildings/BuildingRecordsWorkspace';

describe('Product Workspace Component Structure', () => {
  it('exports all approved product screens and shell components', () => {
    expect(LandingPage).toBeDefined();
    expect(WorkspaceShell).toBeDefined();
    expect(NewInspectionForm).toBeDefined();
    expect(ReportsWorkspace).toBeDefined();
    expect(BuildingRecordsWorkspace).toBeDefined();
  });

  it('ensures core user strings have no hyphens or dashes as punctuation', () => {
    const verifiedCopy = [
      'Plan each inspection with the map in view.',
      'For county inspection teams',
      'A shared view for county building teams.',
      'From search to inspection report.',
      'Keep each step attached to the building record available to your team.',
      'Locate: Search the MapTiler map or an approved building layer.',
      'Review: Open available property information and history.',
      'Inspect: Record observations against the selected building.',
      'Report: Review results from submitted inspection records.',
      'Illustration. Replace with your MapTiler map and approved data.',
      'No inspection data yet',
      'Building list is empty',
      'Inspection queue will appear here',
    ];

    verifiedCopy.forEach((text) => {
      // Must not contain em-dash, en-dash, or hyphen used as punctuation
      expect(text).not.toMatch(/[—–]/);
      expect(text).not.toMatch(/\s-\s/);
      expect(text).not.toMatch(/\w-\w/);
    });
  });
});
