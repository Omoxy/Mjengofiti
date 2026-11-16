package indicators_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mjengofiti/api/internal/indicators"
)

func TestIndicatorDefinitions(t *testing.T) {
	repo := indicators.NewRepository(nil)
	defs, err := repo.GetDefinitions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(defs) == 0 {
		t.Fatal("expected at least 4 core indicators")
	}

	// Verify all indicators have non-empty calculation methods and explicit limitations
	for _, d := range defs {
		if d.CalculationMethod == "" {
			t.Errorf("indicator %s is missing calculation method", d.Name)
		}
		if d.Limitations == "" {
			t.Errorf("indicator %s is missing explicit limitations statement", d.Name)
		}
		if d.Purpose == "" {
			t.Errorf("indicator %s is missing explicit purpose", d.Name)
		}
	}
}

func TestAnalysisJobIsExplicitlyUnavailable(t *testing.T) {
	handler := indicators.NewHandler(indicators.NewRepository(nil), nil)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/analysis-jobs", nil)
	response := httptest.NewRecorder()

	handler.HandleCreateJob(response, request)

	if response.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501 until a calculation worker exists, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "analysis_unavailable") {
		t.Fatalf("expected an explicit unavailable error, got %q", response.Body.String())
	}
}

func TestIndicatorsHandler(t *testing.T) {
	repo := indicators.NewRepository(nil)
	handler := indicators.NewHandler(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/indicators", nil)
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected HTTP 200, got %d", w.Code)
	}
}
