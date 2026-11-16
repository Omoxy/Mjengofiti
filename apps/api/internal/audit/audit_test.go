package audit_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mjengofiti/api/internal/audit"
)

func TestAuditLogAndList(t *testing.T) {
	repo := audit.NewRepository(nil)
	ctx := context.Background()

	ev := audit.Event{
		ID:           "test-aud-1",
		UserEmail:    "test@example.com",
		UserRole:     "inspector",
		Action:       "case.create",
		ResourceType: "case",
		ResourceID:   "c-123",
		IPAddress:    "127.0.0.1",
		CreatedAt:    time.Now(),
	}

	if err := repo.Log(ctx, ev); err != nil {
		t.Fatalf("failed to log audit event: %v", err)
	}

	res, err := repo.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("failed to list audit events: %v", err)
	}

	if res.Total < 1 {
		t.Errorf("expected at least 1 audit event, got %d", res.Total)
	}
}

func TestAuditHandler(t *testing.T) {
	repo := audit.NewRepository(nil)
	handler := audit.NewHandler(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit-events", nil)
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected HTTP 200, got %d", w.Code)
	}
}
