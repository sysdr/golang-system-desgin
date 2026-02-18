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
	body := w.Body.String()
	if !strings.Contains(body, "Memory Alignment Dashboard") {
		t.Error("dashboard body missing title")
	}
	if !strings.Contains(body, "Run Analysis") {
		t.Error("dashboard body missing Run Analysis")
	}
}

func TestRunAnalysisHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/run-analysis", nil)
	w := httptest.NewRecorder()
	runAnalysisHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("run-analysis status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "ok") {
		t.Error("run-analysis body missing ok")
	}
}

func TestMemStatsHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/debug/mem", nil)
	w := httptest.NewRecorder()
	memStatsHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("mem status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "NumGC") {
		t.Error("mem body missing NumGC")
	}
}
