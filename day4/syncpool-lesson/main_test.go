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
	if !strings.Contains(w.Body.String(), "sync.Pool Lesson Dashboard") {
		t.Error("dashboard body missing title")
	}
	if !strings.Contains(w.Body.String(), "Buggy requests") {
		t.Error("dashboard body missing Buggy request metrics")
	}
	if !strings.Contains(w.Body.String(), "Fixed requests") {
		t.Error("dashboard body missing Fixed request metrics")
	}
}

func TestBuggyHandler(t *testing.T) {
	req := httptest.NewRequest("POST", "/buggy", strings.NewReader("test"))
	w := httptest.NewRecorder()
	buggyHandlerWithStats(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("buggy status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Buggy response") {
		t.Errorf("buggy body = %s", w.Body.String())
	}
}

func TestFixedHandler(t *testing.T) {
	req := httptest.NewRequest("POST", "/fixed", strings.NewReader("test"))
	w := httptest.NewRecorder()
	fixedHandlerWithStats(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("fixed status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Fixed response") {
		t.Errorf("fixed body = %s", w.Body.String())
	}
}

func TestMemStatsHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/debug/mem", nil)
	w := httptest.NewRecorder()
	handleMemStats(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("memstats status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "NumGC") {
		t.Error("memstats body missing NumGC")
	}
}
