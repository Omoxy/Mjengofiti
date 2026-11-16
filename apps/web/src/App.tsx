import { useState, type ReactNode } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import { LandingPage } from './features/landing/LandingPage';
import { WorkspaceShell } from './components/layout/WorkspaceShell';
import { BuildingIntelligence } from './features/buildings/BuildingIntelligence';
import { BuildingRecordsWorkspace } from './features/buildings/BuildingRecordsWorkspace';
import { InspectionsWorkspace } from './features/inspections/InspectionsWorkspace';
import { NewInspectionForm } from './features/inspections/NewInspectionForm';
import { ReportsWorkspace } from './features/reports/ReportsWorkspace';
import { ResearchWorkspace } from './features/research/ResearchWorkspace';
import { AnalysisWorkspace } from './features/analysis/AnalysisWorkspace';
import { GovernanceWorkspace } from './features/governance/GovernanceWorkspace';
import { AdminWorkspace } from './features/administration/AdminWorkspace';
import type { Building, BuildingDetail } from './api/client';
import { useAuth } from './hooks/useAuth';

function PermissionRoute({ permission, children }: { permission: string; children: ReactNode }) {
  const { user, isLoading, hasPermission } = useAuth();

  if (isLoading) {
    return <main role="status">Checking access...</main>;
  }

  if (!user || !hasPermission(permission)) {
    return (
      <main role="alert">
        <h1>Access unavailable</h1>
        <p>This area requires an authenticated account with the appropriate permission.</p>
      </main>
    );
  }

  return <>{children}</>;
}

function WorkspaceRouter() {
  const navigate = useNavigate();
  const [selectedBuilding, setSelectedBuilding] = useState<Building | BuildingDetail | null>(null);
  const [searchTarget, setSearchTarget] = useState<string>('');

  const handleStartInspection = (building: Building | BuildingDetail) => {
    setSelectedBuilding(building);
    navigate('/inspections/new');
  };

  return (
    <Routes>
      {/* Public Landing Page */}
      <Route
        path="/"
        element={<LandingPage onEnterWorkspace={() => navigate('/buildings')} />}
      />

      {/* Map Overview Workspace (Main View) */}
      <Route
        path="/buildings"
        element={
          <WorkspaceShell
            selectedBuilding={selectedBuilding as BuildingDetail | null}
            activeScreenTitle="Map overview"
            activeScreenDesc="Review building locations and available filters."
          >
            <BuildingIntelligence
              searchTarget={searchTarget}
              onResetSearchTarget={() => setSearchTarget('')}
              onBuildingSelectNotify={(b) => setSelectedBuilding(b)}
              onStartInspection={handleStartInspection}
            />
          </WorkspaceShell>
        }
      />
      <Route path="/workspace" element={<Navigate to="/buildings" replace />} />
      <Route path="/workspace/map" element={<Navigate to="/buildings" replace />} />

      {/* Building Records Tab */}
      <Route
        path="/workspace/buildings"
        element={
          <WorkspaceShell
            selectedBuilding={selectedBuilding as BuildingDetail | null}
            activeScreenTitle="Building records"
            activeScreenDesc="Search and review available building records."
          >
            <BuildingRecordsWorkspace
              onSelectBuilding={(b) => setSelectedBuilding(b)}
              onStartInspection={handleStartInspection}
              selectedBuildingId={selectedBuilding?.id}
            />
          </WorkspaceShell>
        }
      />
      <Route path="/buildings/list" element={<Navigate to="/workspace/buildings" replace />} />

      {/* Inspections Queue Tab */}
      <Route
        path="/inspections"
        element={
          <PermissionRoute permission="inspections:read">
          <WorkspaceShell
            selectedBuilding={selectedBuilding as BuildingDetail | null}
            activeScreenTitle="Inspection queue"
            activeScreenDesc="Priority follows the rules set in your workspace."
          >
            <InspectionsWorkspace />
          </WorkspaceShell>
          </PermissionRoute>
        }
      />
      <Route path="/workspace/inspections" element={<Navigate to="/inspections" replace />} />

      {/* New Inspection Entry Form */}
      <Route
        path="/inspections/new"
        element={
          <PermissionRoute permission="inspections:triage">
          <WorkspaceShell
            selectedBuilding={selectedBuilding as BuildingDetail | null}
            activeScreenTitle="Field inspection"
            activeScreenDesc="Record observations using the approved inspection protocol."
          >
            <NewInspectionForm
              selectedBuilding={selectedBuilding}
              onSuccess={() => navigate('/inspections')}
              onSelectBuildingPrompt={() => navigate('/buildings')}
            />
          </WorkspaceShell>
          </PermissionRoute>
        }
      />
      <Route path="/workspace/inspections/new" element={<Navigate to="/inspections/new" replace />} />
      <Route path="/workspace/capture" element={<Navigate to="/inspections/new" replace />} />

      {/* Reports Tab */}
      <Route
        path="/workspace/reports"
        element={
          <PermissionRoute permission="inspections:read">
          <WorkspaceShell
            selectedBuilding={selectedBuilding as BuildingDetail | null}
            activeScreenTitle="Reports"
            activeScreenDesc="Review verified inspection outcomes and reporting summaries."
          >
            <ReportsWorkspace />
          </WorkspaceShell>
          </PermissionRoute>
        }
      />
      <Route path="/reports" element={<Navigate to="/workspace/reports" replace />} />

      {/* Extended Preserved Workspaces */}
      <Route
        path="/research"
        element={
          <WorkspaceShell
            selectedBuilding={selectedBuilding as BuildingDetail | null}
            activeScreenTitle="Research & Data"
            activeScreenDesc="Geospatial dataset registry and versions."
          >
            <ResearchWorkspace />
          </WorkspaceShell>
        }
      />
      <Route
        path="/analysis"
        element={
          <WorkspaceShell
            selectedBuilding={selectedBuilding as BuildingDetail | null}
            activeScreenTitle="Analysis & Methodology"
            activeScreenDesc="Codified indicators and automated triage calculations."
          >
            <AnalysisWorkspace />
          </WorkspaceShell>
        }
      />
      <Route
        path="/governance"
        element={
          <PermissionRoute permission="audit:read">
          <WorkspaceShell
            selectedBuilding={selectedBuilding as BuildingDetail | null}
            activeScreenTitle="Security & Governance"
            activeScreenDesc="National building code standards and immutable audit trail."
          >
            <GovernanceWorkspace />
          </WorkspaceShell>
          </PermissionRoute>
        }
      />
      <Route
        path="/admin"
        element={
          <PermissionRoute permission="system:read">
          <WorkspaceShell
            selectedBuilding={selectedBuilding as BuildingDetail | null}
            activeScreenTitle="Administration"
            activeScreenDesc="System telemetry, PostGIS connection status, and ingestion diagnostics."
          >
            <AdminWorkspace />
          </WorkspaceShell>
          </PermissionRoute>
        }
      />

      {/* Catch-all */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

export function App() {
  return (
    <BrowserRouter>
      <WorkspaceRouter />
    </BrowserRouter>
  );
}

export default App;
