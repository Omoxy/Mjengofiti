// API client with TypeScript interfaces for Mjengofiti Go backend

const API_BASE = '';

export interface MapConfig {
  tile_url: string;
  center: [number, number];
  zoom: number;
}

export interface Building {
  id: string;
  external_id?: string;
  source_dataset_id?: string;
  source_name: string;
  longitude: number;
  latitude: number;
  area_sq_m?: number;
  height_m?: number;
  height_method?: string;
  match_method?: string;
  match_confidence?: number;
  data_quality: string;
  created_at: string;
  updated_at: string;
}

export interface BuildingDetail extends Building {
  observations: Observation[];
  indicators: Indicator[];
  geojson?: string;
}

export interface Observation {
  id: string;
  type: string;
  value: string;
  unit?: string;
  observation_date?: string;
  source_version: string;
  method?: string;
  confidence?: number;
  limitations?: string;
}

export interface Indicator {
  id: string;
  indicator_type: string;
  version: string;
  value: number;
  unit: string;
  calculation_method: string;
  input_datasets: string[];
  limitations?: string;
}

export interface Dataset {
  id: string;
  name: string;
  description: string;
  publisher: string;
  source_url: string;
  licence: string;
  licence_url?: string;
  data_type: string;
  format: string;
  srid: number;
  geo_coverage: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface DatasetVersion {
  id: string;
  dataset_id: string;
  version: string;
  acquisition_at: string;
  checksum?: string;
  file_size?: number;
  feature_count?: number;
  geo_coverage_pct?: string;
  validation_status: string;
  promotion_status: string;
  promoted_at?: string;
  promoted_by?: string;
  notes?: string;
}

export interface IndicatorDefinition {
  id: string;
  name: string;
  display_name: string;
  description: string;
  purpose: string;
  calculation_method: string;
  input_schema: Record<string, unknown>;
  output_unit: string;
  missing_data_behaviour: string;
  limitations: string;
  version: string;
  is_active: boolean;
}

export interface InspectionCase {
  id: string;
  building_id: string;
  title: string;
  description: string;
  priority: 'low' | 'medium' | 'high' | 'critical';
  status: 'open' | 'triaged' | 'assigned' | 'in_progress' | 'pending_review' | 'closed' | 'reopened';
  referral_reason: string;
  evidence_refs: string[];
  assigned_to?: string;
  assigned_at?: string;
  due_date?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
  closed_at?: string;
}

export interface Standard {
  id: string;
  authority: string;
  document_title: string;
  version: string;
  publication_date: string;
  legal_status: string;
  clause_ref: string;
  clause_title: string;
  requirement: string;
  interpretation: string;
  software_rule: string;
  reviewed_by: string;
  reviewed_at: string;
}

export interface AuditEvent {
  id: string;
  user_email: string;
  user_role: string;
  action: string;
  resource_type: string;
  resource_id: string;
  details?: Record<string, unknown>;
  ip_address: string;
  created_at: string;
}

export interface UserProfile {
  id: string;
  email: string;
  display_name: string;
  role: string;
  permissions: string[];
}

export interface SystemStatus {
  status: string;
  database: {
    status: string;
    type: string;
  };
  app_env: string;
  uptime_seconds: number;
}

// Fetch helper with error handling
async function request<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!res.ok) {
    let errBody: { error?: string; message?: string } = {};
    try {
      errBody = await res.json();
    } catch {
      // ignore
    }
    const message = errBody.message || errBody.error || `HTTP ${res.status} ${res.statusText}`;
    const err = new Error(message) as Error & { status: number; code?: string };
    err.status = res.status;
    err.code = errBody.error;
    throw err;
  }

  return res.json();
}

// API Functions
export const api = {
  getMapConfig: () => request<MapConfig>('/api/v1/config/map'),
  getBoundaries: () => request<any>('/api/v1/boundaries'),
  getCoverage: () => request<{ database_records?: number; coverage: string; status?: string; note?: string }>('/api/v1/coverage'),
  getSystemStatus: () => request<SystemStatus>('/api/v1/system/status'),
  getUserProfile: () => request<UserProfile>('/api/v1/auth/me'),
  getRoles: () => request<{ roles: Array<{ role: string; title: string; description: string; permissions: string[] }> }>('/api/v1/auth/roles'),

  // Buildings
  getBuildings: (params: { bbox?: string; q?: string; limit?: number; offset?: number }) => {
    const qs = new URLSearchParams();
    if (params.bbox) qs.set('bbox', params.bbox);
    if (params.q) qs.set('q', params.q);
    if (params.limit) qs.set('limit', String(params.limit));
    if (params.offset) qs.set('offset', String(params.offset));
    return request<{ buildings: Building[]; total: number }>('/api/v1/buildings?' + qs.toString());
  },
  getBuildingDetail: (id: string) => request<BuildingDetail>(`/api/v1/buildings/${id}`),

  // Datasets
  getDatasets: () => request<{ datasets: Dataset[]; total: number }>('/api/v1/datasets'),
  getDataset: (id: string) => request<Dataset>(`/api/v1/datasets/${id}`),
  getDatasetVersions: (id: string) => request<{ dataset_id: string; versions: DatasetVersion[] }>(`/api/v1/datasets/${id}/versions`),

  // Indicators
  getIndicators: () => request<{ definitions: IndicatorDefinition[]; total: number; disclaimer: string }>('/api/v1/indicators'),
  getIndicator: (id: string) => request<IndicatorDefinition>(`/api/v1/indicators/${id}`),
  triggerAnalysisJob: (indicatorName: string) =>
    request('/api/v1/analysis-jobs', {
      method: 'POST',
      body: JSON.stringify({ indicator_name: indicatorName }),
    }),

  // Inspections
  getInspectionCases: (status?: string) => {
    const qs = status ? `?status=${encodeURIComponent(status)}` : '';
    return request<{ cases: InspectionCase[]; total: number }>(`/api/v1/inspection-cases${qs}`);
  },
  createInspectionCase: (data: Partial<InspectionCase>) =>
    request<InspectionCase>('/api/v1/inspection-cases', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  updateInspectionCaseStatus: (id: string, status: string) =>
    request<InspectionCase>(`/api/v1/inspection-cases/${id}`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    }),

  // Standards
  getStandards: () => request<{ standards: Standard[]; total: number; message: string }>('/api/v1/standards'),

  // Audit
  getAuditEvents: (limit = 50) => request<{ events: AuditEvent[]; total: number }>(`/api/v1/audit-events?limit=${limit}`),

  // Ingestion
  getIngestionSources: () => request<{ sources: IngestionSource[]; total: number }>('/api/v1/ingestion/sources'),
  getIngestionJobs: () => request<{ jobs: IngestionJob[]; total: number }>('/api/v1/ingestion/jobs'),
  createIngestionJob: (data: { dataset_id: string; version?: string; created_by?: string }) =>
    request<IngestionJob>('/api/v1/ingestion/jobs', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  validateGeoJSON: (data: unknown, versionId?: string) =>
    request<ValidationReport>(`/api/v1/ingestion/validate${versionId ? `?version_id=${encodeURIComponent(versionId)}` : ''}`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  promoteDataset: (data: { dataset_id: string; version: string; promoted_by?: string; notes?: string; force?: boolean }) =>
    request<PromotionResult>('/api/v1/ingestion/promote', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
};

export interface IngestionSource {
  id: string;
  name: string;
  publisher: string;
  source_url: string;
  format: string;
  expected_crs: string;
  licence: string;
  licence_url?: string;
  description: string;
  created_at: string;
}

export interface IngestionJob {
  id: string;
  dataset_id: string;
  dataset_version: string;
  status: string;
  job_type: string;
  started_at?: string;
  completed_at?: string;
  records_total: number;
  records_ingested: number;
  records_failed: number;
  error_log?: string;
  created_by: string;
  created_at: string;
}

export interface ValidationReport {
  id: string;
  dataset_version_id: string;
  valid_geometries: number;
  invalid_geometries: number;
  duplicate_count: number;
  missing_values: number;
  crs_valid: boolean;
  bounds_valid: boolean;
  overall_status: 'passed' | 'passed_with_warnings' | 'failed';
  issues?: string[];
  validated_at: string;
}

export interface PromotionResult {
  dataset_id: string;
  version: string;
  status: string;
  promoted_at: string;
  promoted_by: string;
  message: string;
}
