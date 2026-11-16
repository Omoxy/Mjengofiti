package datasets_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mjengofiti/api/internal/datasets"
)

func TestDatasetsListAndFind(t *testing.T) {
	repo := datasets.NewRepository(nil)
	ctx := context.Background()

	dsList, total, err := repo.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error listing datasets: %v", err)
	}

	if total < 5 {
		t.Errorf("expected at least 5 registered candidate datasets, got %d", total)
	}

	// Verify Microsoft and Google datasets
	foundMS := false
	foundGoogle := false
	for _, d := range dsList {
		if d.ID == "ds-ms-kenya" {
			foundMS = true
			if d.Status != "registered" {
				t.Errorf("expected Microsoft source to remain registered, got %q", d.Status)
			}
		}
		if d.ID == "ds-google-openbldg" {
			foundGoogle = true
			if d.Status != "registered" {
				t.Errorf("expected Google source to remain registered, got %q", d.Status)
			}
		}
	}

	if !foundMS {
		t.Error("expected Microsoft Kenya Footprints in dataset catalogue")
	}
	if !foundGoogle {
		t.Error("expected Google Open Buildings in dataset catalogue")
	}

	// Find by ID
	msDataset, err := repo.FindByID(ctx, "ds-ms-kenya")
	if err != nil {
		t.Fatalf("failed finding dataset ds-ms-kenya: %v", err)
	}
	if msDataset.Publisher != "Microsoft Maps / Bing Imagery Team" {
		t.Errorf("unexpected publisher: %s", msDataset.Publisher)
	}

	// List versions
	versions, err := repo.ListVersions(ctx, "ds-ms-kenya")
	if err != nil {
		t.Fatalf("failed listing versions: %v", err)
	}
	if len(versions) == 0 {
		return
	}
	t.Fatalf("expected no synthetic versions for an unacquired source, got %d", len(versions))
}

func TestDatasetsHandler(t *testing.T) {
	repo := datasets.NewRepository(nil)
	handler := datasets.NewHandler(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/datasets", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected HTTP 200, got %d", w.Code)
	}
}
