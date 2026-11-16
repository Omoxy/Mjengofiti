package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mjengofiti/api/internal/audit"
	"github.com/mjengofiti/api/internal/datasets"
	"github.com/mjengofiti/api/internal/indicators"
	"github.com/mjengofiti/api/internal/ingestion"
	"github.com/mjengofiti/api/internal/inspections"
)

func TestSensitiveRoutesRejectAnonymousRequests(t *testing.T) {
	router := chi.NewRouter()
	audit.NewHandler(nil, nil).Routes(router)
	datasets.NewHandler(nil, nil).Routes(router)
	indicators.NewHandler(nil, nil).Routes(router)
	ingestion.NewHandler(nil, nil).Routes(router)
	inspections.NewHandler(nil, nil).Routes(router)

	for _, testCase := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/audit-events"},
		{method: http.MethodGet, path: "/api/v1/inspection-cases"},
		{method: http.MethodPost, path: "/api/v1/inspection-cases"},
				{method: http.MethodPut, path: "/api/v1/inspection-cases/case-1/assignment"},
		{method: http.MethodPatch, path: "/api/v1/inspection-cases/case-1"},
		{method: http.MethodGet, path: "/api/v1/inspection-cases/case-1/history"},
		{method: http.MethodPost, path: "/api/v1/ingestion/sources"},
		{method: http.MethodPost, path: "/api/v1/ingestion/acquire"},
		{method: http.MethodGet, path: "/api/v1/ingestion/artifacts/version-1"},
		{method: http.MethodGet, path: "/api/v1/ingestion/jobs"},
		{method: http.MethodPost, path: "/api/v1/ingestion/jobs"},
		{method: http.MethodPost, path: "/api/v1/ingestion/validate"},
		{method: http.MethodGet, path: "/api/v1/ingestion/reports/version-1"},
		{method: http.MethodPost, path: "/api/v1/ingestion/promote"},
		{method: http.MethodGet, path: "/api/v1/analysis-jobs/job-1"},
		{method: http.MethodGet, path: "/api/v1/datasets/dataset-1/versions"},
		{method: http.MethodGet, path: "/api/v1/jobs/job-1"},
	} {
		t.Run(testCase.method+" "+testCase.path, func(t *testing.T) {
			request := httptest.NewRequest(testCase.method, testCase.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("expected anonymous request to receive 401, got %d", response.Code)
			}
		})
	}
}
