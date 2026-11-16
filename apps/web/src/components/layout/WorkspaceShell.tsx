import React, { useState } from 'react';
import { NavLink, useNavigate, useLocation } from 'react-router-dom';
import {
  Compass,
  Building2,
  FileCheck2,
  BarChart3,
  Plus,
  Database,
  Calculator,
  ShieldCheck,
  Settings,
  LogOut,
  ChevronDown,
  Layers,
  Menu,
  X,
} from 'lucide-react';
import type { BuildingDetail } from '../../api/client';
import { useAuth } from '../../hooks';

interface WorkspaceShellProps {
  children: React.ReactNode;
  selectedBuilding?: BuildingDetail | null;
  onClearSelectedBuilding?: () => void;
  activeScreenTitle?: string;
  activeScreenDesc?: string;
}

export const WorkspaceShell: React.FC<WorkspaceShellProps> = ({
  children,
  selectedBuilding,
  activeScreenTitle,
  activeScreenDesc,
}) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user } = useAuth();
  const [showExtendedNav, setShowExtendedNav] = useState(false);
  const [isMobileNavOpen, setIsMobileNavOpen] = useState(false);
  const isMapWorkspace = location.pathname === '/buildings';
  const isExtendedRoute = ['/research', '/analysis', '/governance', '/admin'].includes(location.pathname);
  const extendedNavOpen = showExtendedNav || isExtendedRoute;

  // Determine current screen title and description
  const getScreenInfo = () => {
    if (activeScreenTitle) {
      return { title: activeScreenTitle, desc: activeScreenDesc || '' };
    }
    const path = location.pathname;
    if (path === '/workspace' || path === '/workspace/map' || path === '/buildings') {
      return {
        title: 'Map overview',
        desc: 'Review building locations and available filters.',
      };
    }
    if (path.includes('/buildings/list') || path === '/workspace/buildings') {
      return {
        title: 'Building records',
        desc: 'Search and review available building records.',
      };
    }
    if (path === '/inspections/new' || path === '/workspace/inspections/new') {
      return {
        title: 'Field inspection',
        desc: 'Record observations using the approved inspection protocol.',
      };
    }
    if (path.includes('/inspections') || path === '/workspace/inspections') {
      return {
        title: 'Inspection queue',
        desc: 'Priority follows the rules set in your workspace.',
      };
    }
    if (path.includes('/reports') || path === '/workspace/reports') {
      return {
        title: 'Reports',
        desc: 'Review verified inspection outcomes and reporting summaries.',
      };
    }
    if (path.includes('/research')) {
      return {
        title: 'Research & Data',
        desc: 'Dataset catalogue, versions, and validation records.',
      };
    }
    if (path.includes('/analysis')) {
      return {
        title: 'Analysis & Methodology',
        desc: 'Engineering indicators and automated triage calculations.',
      };
    }
    if (path.includes('/governance')) {
      return {
        title: 'Security & Governance',
        desc: 'National building code standards and audit trail.',
      };
    }
    if (path.includes('/admin')) {
      return {
        title: 'Administration',
        desc: 'System health, PostGIS connection status, and diagnostics.',
      };
    }
    return {
      title: 'Map overview',
      desc: 'Review building locations and available filters.',
    };
  };

  const screenInfo = getScreenInfo();

  return (
    <div className={`workspace-shell${isMapWorkspace ? ' gis-workspace-shell' : ''}`}>
      {/* Left Rail Navigation on Desktop, Reachable on Mobile */}
      <aside
        id="workspace-navigation"
        className={`workspace-sidebar${isMobileNavOpen ? ' mobile-open' : ''}`}
        aria-label="Main application navigation"
        onClick={(event) => {
          if (event.target instanceof Element && event.target.closest('a')) {
            setIsMobileNavOpen(false);
          }
        }}
      >
        <NavLink to="/" className="brand-link" style={{ fontSize: '16px' }} title="Return to public landing">
          <span className="brandmark-box" style={{ width: '32px', height: '32px', borderRadius: '9px' }}>
            <Layers size={16} />
          </span>
          <span>
            mjengo <em>fiti</em>
          </span>
        </NavLink>

        <div className="workspace-sidebar-caption">Inspection workspace</div>

        {/* Primary 5 Core Workspace Items */}
        <nav className="workspace-nav-group" aria-label="Core operational tabs">
          <NavLink
            to="/buildings"
            end
            className={({ isActive }) =>
              `workspace-nav-item ${isActive || location.pathname === '/workspace' || location.pathname === '/workspace/map' ? 'active' : ''}`
            }
          >
            <span className="nav-glyph" aria-hidden="true">
              <Compass size={16} />
            </span>
            <span>Map overview</span>
          </NavLink>

          <NavLink
            to="/workspace/buildings"
            className={({ isActive }) =>
              `workspace-nav-item ${isActive ? 'active' : ''}`
            }
          >
            <span className="nav-glyph" aria-hidden="true">
              <Building2 size={16} />
            </span>
            <span>Buildings</span>
          </NavLink>

          <NavLink
            to="/inspections"
            end
            className={({ isActive }) =>
              `workspace-nav-item ${isActive && !location.pathname.includes('/new') ? 'active' : ''}`
            }
          >
            <span className="nav-glyph" aria-hidden="true">
              <FileCheck2 size={16} />
            </span>
            <span>Inspections</span>
          </NavLink>

          <NavLink
            to="/workspace/reports"
            className={({ isActive }) =>
              `workspace-nav-item ${isActive ? 'active' : ''}`
            }
          >
            <span className="nav-glyph" aria-hidden="true">
              <BarChart3 size={16} />
            </span>
            <span>Reports</span>
          </NavLink>

          <NavLink
            to="/inspections/new"
            className={({ isActive }) =>
              `workspace-nav-item new-inspection-btn ${isActive ? 'active' : ''}`
            }
          >
            <span className="nav-glyph" aria-hidden="true">
              <Plus size={16} />
            </span>
            <span>New inspection</span>
          </NavLink>
        </nav>

        {/* Secondary Extended System Workspaces */}
        <div style={{ marginTop: '10px' }}>
          <button
            type="button"
            className="workspace-nav-item workspace-nav-section-toggle"
            onClick={() => setShowExtendedNav(!extendedNavOpen)}
            aria-expanded={extendedNavOpen}
            aria-controls="extended-workspace-navigation"
          >
            <span style={{ fontWeight: 700, letterSpacing: '0.5px', textTransform: 'uppercase' }}>
              Extended Tools
            </span>
            <ChevronDown
              size={12}
              style={{
                transform: extendedNavOpen ? 'rotate(180deg)' : 'none',
                transition: 'transform 0.2s ease',
              }}
            />
          </button>

          {extendedNavOpen && (
            <div id="extended-workspace-navigation" className="workspace-nav-group workspace-extended-nav">
              <NavLink
                to="/research"
                className={({ isActive }) => `workspace-nav-item ${isActive ? 'active' : ''}`}
                style={{ fontSize: '12px' }}
              >
                <span className="nav-glyph">
                  <Database size={14} />
                </span>
                <span>Research & Data</span>
              </NavLink>

              <NavLink
                to="/analysis"
                className={({ isActive }) => `workspace-nav-item ${isActive ? 'active' : ''}`}
                style={{ fontSize: '12px' }}
              >
                <span className="nav-glyph">
                  <Calculator size={14} />
                </span>
                <span>Analysis & Metrics</span>
              </NavLink>

              <NavLink
                to="/governance"
                className={({ isActive }) => `workspace-nav-item ${isActive ? 'active' : ''}`}
                style={{ fontSize: '12px' }}
              >
                <span className="nav-glyph">
                  <ShieldCheck size={14} />
                </span>
                <span>Governance & Codes</span>
              </NavLink>

              <NavLink
                to="/admin"
                className={({ isActive }) => `workspace-nav-item ${isActive ? 'active' : ''}`}
                style={{ fontSize: '12px' }}
              >
                <span className="nav-glyph">
                  <Settings size={14} />
                </span>
                <span>Administration</span>
              </NavLink>
            </div>
          )}
        </div>

        {/* Sidebar Footer with Signed In Profile */}
        <div className="workspace-sidebar-bottom">
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <span style={{ fontWeight: 700, color: 'var(--ink)' }}>
              {user?.display_name ? user.display_name.split(' (')[0] : 'Eng. J. Kamau'}
            </span>
            <button
              type="button"
              className="mjengo-btn-ghost"
              style={{ padding: '2px 4px', fontSize: '10px' }}
              onClick={() => navigate('/')}
              title="Sign out to landing page"
            >
              <LogOut size={13} />
            </button>
          </div>
          <p>Nairobi County Inspectorate</p>
          {selectedBuilding && (
            <div
              style={{
                marginTop: '10px',
                padding: '6px 8px',
                backgroundColor: 'var(--winep)',
                borderRadius: '6px',
                fontSize: '10px',
                color: 'var(--wine)',
              }}
            >
              <b>Active selection:</b> {selectedBuilding.external_id || selectedBuilding.id.substring(0, 10)}
            </div>
          )}
        </div>
      </aside>

      {isMobileNavOpen && (
        <button
          type="button"
          className="workspace-nav-backdrop"
          aria-label="Close navigation"
          onClick={() => setIsMobileNavOpen(false)}
        />
      )}

      {/* Main Workspace Canvas */}
      <main className="workspace-main" role="main">
        {/* Workspace Subheader with Crumb & Title */}
        <header className={`workspace-header${isMapWorkspace ? ' gis-workspace-header' : ''}`}>
          <button
            type="button"
            className="workspace-mobile-menu"
            aria-label={isMobileNavOpen ? 'Close navigation' : 'Open navigation'}
            aria-expanded={isMobileNavOpen}
            aria-controls="workspace-navigation"
            onClick={() => setIsMobileNavOpen((open) => !open)}
          >
            {isMobileNavOpen ? <X size={18} /> : <Menu size={18} />}
          </button>
          <div className="workspace-heading">
            <div className="workspace-crumb">
              Mjengo Fiti &nbsp;/&nbsp; <span>{screenInfo.title}</span>
            </div>
            <h1 className="workspace-title">{screenInfo.title}</h1>
            <p className="workspace-desc">{screenInfo.desc}</p>
          </div>
        </header>

        {/* Content Outlet with Entrance Motion */}
        <div className={`workspace-content screen-enter-animation${isMapWorkspace ? ' gis-workspace-content' : ''}`}>
          {children}
        </div>
      </main>
    </div>
  );
};
