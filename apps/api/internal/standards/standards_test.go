package standards_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mjengofiti/api/internal/standards"
)

func TestStandardsRepositoryReturnsEmptyWithoutReviewedRecords(t *testing.T) {
	repository := standards.NewRepository(nil)
	items, err := repository.GetAuthoritativeStandards(context.Background())
	if err != nil {
		t.Fatalf("unexpected error fetching standards: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no fabricated standards when no database is available, got %d", len(items))
	}
}

func TestStandardsHandlerReturnsEmptyRegister(t *testing.T) {
	handler := standards.NewHandler(standards.NewRepository(nil), nil)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/standards", nil)
	response := httptest.NewRecorder()

	handler.HandleList(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON content type, got %q", response.Header().Get("Content-Type"))
	}
	var result struct {
		Standards []standards.Standard `json:"standards"`
		Total     int                  `json:"total"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatalf("decode standards response: %v", err)
	}
	if result.Total != 0 || len(result.Standards) != 0 {
		t.Fatalf("expected empty standards response, got %#v", result)
	}
}
