package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDashboardHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	dashboardHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("dashboard status = %d, want 200", w.Code)
	}
	if w.Body.Len() == 0 {
		t.Error("dashboard body empty")
	}
	if !strings.Contains(w.Body.String(), "Heap Cost Analyzer Dashboard") {
		t.Error("dashboard body missing title")
	}
}

func TestMetricsJSONHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	metricsJSONHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("metrics status = %d, want 200", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %s", w.Header().Get("Content-Type"))
	}
}

