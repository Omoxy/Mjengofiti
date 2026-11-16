-- Migration: 001_initial_schema
-- Purpose: Create core tables for Mjengofiti building intelligence platform.
-- Requires: PostgreSQL 15+ with PostGIS extension.

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";

-- ============================================================
-- USERS AND ROLES
-- ============================================================

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    role_id UUID REFERENCES roles(id),
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- AUDIT EVENTS
-- ============================================================

CREATE TABLE audit_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(255),
    details JSONB,
    ip_address INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_events_user ON audit_events(user_id);
CREATE INDEX idx_audit_events_resource ON audit_events(resource_type, resource_id);
CREATE INDEX idx_audit_events_created ON audit_events(created_at DESC);

-- ============================================================
-- DATASETS CATALOGUE
-- ============================================================

CREATE TABLE datasets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    publisher VARCHAR(255) NOT NULL,
    source_url TEXT NOT NULL DEFAULT '',
    licence VARCHAR(255) NOT NULL DEFAULT 'unknown',
    licence_url TEXT,
    data_type VARCHAR(50) NOT NULL DEFAULT 'vector',
    format VARCHAR(50) NOT NULL DEFAULT 'geojson',
    srid INTEGER NOT NULL DEFAULT 4326,
    geo_coverage VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE dataset_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dataset_id UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,
    acquisition_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    checksum VARCHAR(128),
    file_size BIGINT,
    feature_count INTEGER,
    geo_coverage_pct VARCHAR(50),
    validation_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    promotion_status VARCHAR(50) NOT NULL DEFAULT 'none',
    promoted_at TIMESTAMPTZ,
    promoted_by VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(dataset_id, version)
);

CREATE TABLE ingestion_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dataset_id UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    dataset_version VARCHAR(50),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    job_type VARCHAR(50) NOT NULL DEFAULT 'full',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    records_total INTEGER,
    records_ingested INTEGER,
    records_failed INTEGER,
    error_log TEXT,
    created_by VARCHAR(255) NOT NULL DEFAULT 'system',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ingestion_jobs_dataset ON ingestion_jobs(dataset_id);
CREATE INDEX idx_ingestion_jobs_status ON ingestion_jobs(status);

CREATE TABLE validation_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dataset_version_id UUID NOT NULL REFERENCES dataset_versions(id) ON DELETE CASCADE,
    valid_geometries INTEGER NOT NULL DEFAULT 0,
    invalid_geometries INTEGER NOT NULL DEFAULT 0,
    duplicate_count INTEGER NOT NULL DEFAULT 0,
    missing_values INTEGER NOT NULL DEFAULT 0,
    crs_valid BOOLEAN NOT NULL DEFAULT false,
    bounds_valid BOOLEAN NOT NULL DEFAULT false,
    overall_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    details JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- BUILDINGS
-- ============================================================

CREATE TABLE buildings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id VARCHAR(255),
    source_dataset_id UUID REFERENCES datasets(id),
    source_name VARCHAR(255) NOT NULL,
    geom GEOMETRY(GEOMETRY, 4326) NOT NULL,
    area_sq_m DOUBLE PRECISION,
    height_m DOUBLE PRECISION,
    height_method VARCHAR(100),
    match_method VARCHAR(100),
    match_confidence DOUBLE PRECISION,
    data_quality VARCHAR(50) NOT NULL DEFAULT 'unverified',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Spatial index for bounding box queries
CREATE INDEX idx_buildings_geom ON buildings USING GIST(geom);
CREATE INDEX idx_buildings_source ON buildings(source_name);
CREATE INDEX idx_buildings_external ON buildings(external_id);
CREATE INDEX idx_buildings_dataset ON buildings(source_dataset_id);

-- ============================================================
-- BUILDING OBSERVATIONS
-- ============================================================

CREATE TABLE building_observations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    building_id UUID NOT NULL REFERENCES buildings(id) ON DELETE CASCADE,
    type VARCHAR(100) NOT NULL,
    value TEXT NOT NULL,
    unit VARCHAR(50),
    observation_date VARCHAR(50),
    source_version VARCHAR(100) NOT NULL,
    method VARCHAR(255),
    confidence DOUBLE PRECISION,
    limitations TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_observations_building ON building_observations(building_id);

-- ============================================================
-- BUILDING INDICATORS
-- ============================================================

CREATE TABLE building_indicators (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    building_id UUID NOT NULL REFERENCES buildings(id) ON DELETE CASCADE,
    indicator_type VARCHAR(100) NOT NULL,
    version VARCHAR(50) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    unit VARCHAR(50) NOT NULL,
    calculation_method TEXT NOT NULL,
    input_datasets TEXT[] NOT NULL DEFAULT '{}',
    missing_inputs TEXT[],
    uncertainty TEXT,
    limitations TEXT,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_indicators_building ON building_indicators(building_id);
CREATE INDEX idx_indicators_type ON building_indicators(indicator_type);

-- ============================================================
-- INDICATOR DEFINITIONS
-- ============================================================

CREATE TABLE indicator_definitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    purpose TEXT NOT NULL,
    calculation_method TEXT NOT NULL,
    input_schema JSONB NOT NULL DEFAULT '{}',
    output_unit VARCHAR(50) NOT NULL,
    missing_data_behaviour TEXT NOT NULL DEFAULT 'returns null',
    limitations TEXT,
    version VARCHAR(50) NOT NULL DEFAULT '1.0.0',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- INSPECTION CASES
-- ============================================================

CREATE TABLE inspection_cases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    building_id UUID NOT NULL REFERENCES buildings(id),
    title VARCHAR(500) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    referral_reason TEXT NOT NULL DEFAULT '',
    evidence_refs TEXT[] NOT NULL DEFAULT '{}',
    assigned_to VARCHAR(255),
    assigned_at TIMESTAMPTZ,
    due_date TIMESTAMPTZ,
    created_by VARCHAR(255) NOT NULL DEFAULT 'system',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    CONSTRAINT chk_priority CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT chk_status CHECK (status IN ('open', 'triaged', 'assigned', 'in_progress', 'pending_review', 'closed', 'reopened'))
);

CREATE INDEX idx_cases_building ON inspection_cases(building_id);
CREATE INDEX idx_cases_status ON inspection_cases(status);
CREATE INDEX idx_cases_assigned ON inspection_cases(assigned_to);

CREATE TABLE case_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    case_id UUID NOT NULL REFERENCES inspection_cases(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    from_status VARCHAR(50),
    to_status VARCHAR(50),
    changed_by VARCHAR(255) NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    details TEXT
);

CREATE INDEX idx_case_history_case ON case_history(case_id);

CREATE TABLE case_findings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    case_id UUID NOT NULL REFERENCES inspection_cases(id) ON DELETE CASCADE,
    finding_type VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    severity VARCHAR(50) NOT NULL DEFAULT 'informational',
    recorded_by VARCHAR(255) NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    attachments TEXT[] DEFAULT '{}'
);

CREATE INDEX idx_findings_case ON case_findings(case_id);

-- ============================================================
-- STANDARDS REGISTER
-- ============================================================

CREATE TABLE standards (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    authority VARCHAR(255) NOT NULL DEFAULT '',
    document_title VARCHAR(500) NOT NULL,
    official_source TEXT NOT NULL,
    version VARCHAR(100),
    publication_date VARCHAR(50),
    legal_status VARCHAR(100) NOT NULL DEFAULT 'unverified',
    clause_ref VARCHAR(255),
    clause_title VARCHAR(255),
    requirement TEXT,
    interpretation TEXT,
    software_rule TEXT,
    reviewed_by VARCHAR(255),
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- SPATIAL REFERENCE DATA
-- ============================================================

CREATE TABLE boundaries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    boundary_type VARCHAR(100) NOT NULL,
    geom GEOMETRY(MULTIPOLYGON, 4326) NOT NULL,
    source VARCHAR(255) NOT NULL,
    source_url TEXT,
    acquired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    properties JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_boundaries_geom ON boundaries USING GIST(geom);
CREATE INDEX idx_boundaries_type ON boundaries(boundary_type);

-- ============================================================
-- DEFAULT ROLES
-- ============================================================

INSERT INTO roles (name, description, permissions) VALUES
    ('viewer', 'Read-only access to public building data and maps', '["buildings:read","datasets:read","standards:read","coverage:read"]'),
    ('researcher', 'Access to dataset catalogue and research tools', '["buildings:read","datasets:read","datasets:manage","ingestion:create","standards:read","coverage:read"]'),
    ('steward', 'Data steward with promotion and governance authority', '["buildings:read","datasets:read","datasets:manage","datasets:promote","ingestion:create","standards:read","standards:manage","coverage:read"]'),
    ('inspector', 'Field inspection and case management', '["buildings:read","cases:read","cases:create","cases:update","findings:create","datasets:read","standards:read","coverage:read"]'),
    ('supervisor', 'Inspection supervision and case review', '["buildings:read","cases:read","cases:create","cases:update","cases:assign","cases:close","findings:create","findings:review","datasets:read","standards:read","coverage:read"]'),
    ('admin', 'Full system administration', '["*"]');
