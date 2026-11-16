package buildings

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseBBoxValid(t *testing.T) {
	input := "36.8123,-1.2921,36.8345,-1.2804"
	coords, err := parseBBox(input)
	if err != nil {
		t.Fatalf("expected valid bbox parse, got error: %v", err)
	}

	expected := [4]float64{36.8123, -1.2921, 36.8345, -1.2804}
	for i := 0; i < 4; i++ {
		if coords[i] != expected[i] {
			t.Errorf("coordinate %d: expected %f, got %f", i, expected[i], coords[i])
		}
	}
}

func TestParseBBoxInvalid(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{name: "empty string", input: ""},
		{name: "too few coordinates", input: "36.8,-1.2,36.9"},
		{name: "too many coordinates", input: "36.8,-1.2,36.9,-1.1,1.0"},
		{name: "non-numeric coordinate", input: "36.8,invalid,36.9,-1.1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseBBox(tc.input)
			if err == nil {
				t.Fatalf("expected error for input %q, got nil", tc.input)
			}
		})
	}
}

func TestSplitComma(t *testing.T) {
	res := splitComma("a,b,c")
	if len(res) != 3 || res[0] != "a" || res[1] != "b" || res[2] != "c" {
		t.Errorf("unexpected split result: %v", res)
	}

	resEmpty := splitComma("single")
	if len(resEmpty) != 1 || resEmpty[0] != "single" {
		t.Errorf("unexpected single element result: %v", resEmpty)
	}
}

func TestParseIntParam(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test?limit=50&negative=-5&invalid=abc", nil)

	if val := parseIntParam(req, "limit", 20); val != 50 {
		t.Errorf("expected 50, got %d", val)
	}
	if val := parseIntParam(req, "missing", 20); val != 20 {
		t.Errorf("expected default 20 for missing key, got %d", val)
	}
	if val := parseIntParam(req, "negative", 20); val != 20 {
		t.Errorf("expected default 20 for negative value, got %d", val)
	}
	if val := parseIntParam(req, "invalid", 20); val != 20 {
		t.Errorf("expected default 20 for invalid string, got %d", val)
	}
}

func TestListByBBoxInvalidValidation(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := NewHandler(nil, logger)

	// Test invalid bbox format
	req := httptest.NewRequest(http.MethodGet, "/api/v1/buildings?bbox=invalid,coords", nil)
	w := httptest.NewRecorder()
	handler.List(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid bbox, got %d", w.Code)
	}

	var errResp apiError
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed decoding error response: %v", err)
	}
	if errResp.Error != "invalid_bbox" {
		t.Errorf("expected error code invalid_bbox, got %s", errResp.Error)
	}
}

func TestListByBBoxAreaTooLarge(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := NewHandler(nil, logger)

	// Span > 1.0 degree
	req := httptest.NewRequest(http.MethodGet, "/api/v1/buildings?bbox=34.0,-2.0,36.5,0.0", nil)
	w := httptest.NewRecorder()
	handler.List(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for oversized bbox, got %d", w.Code)
	}

	var errResp apiError
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed decoding error response: %v", err)
	}
	if errResp.Error != "bbox_too_large" {
		t.Errorf("expected error code bbox_too_large, got %s", errResp.Error)
	}
}

func TestWriteJSONAndError(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, http.StatusForbidden, "access_denied", "You do not have access")

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}

	wJSON := httptest.NewRecorder()
	writeJSON(wJSON, http.StatusOK, map[string]string{"result": "ok"})
	if wJSON.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", wJSON.Code)
	}
}
