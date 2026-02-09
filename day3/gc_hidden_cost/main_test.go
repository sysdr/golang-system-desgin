package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDashboardHandler(t *testing.T) {
	mux := NewMux()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("dashboard status = %d, want 200", w.Code)
	}
	if w.Body.Len() == 0 {
		t.Error("dashboard body empty")
	}
	if !strings.Contains(w.Body.String(), "GC Hidden Cost Dashboard") {
		t.Error("dashboard body missing title")
	}
	if !strings.Contains(w.Body.String(), "Naive requests") {
		t.Error("dashboard body missing Naive request metrics")
	}
	if !strings.Contains(w.Body.String(), "Pooled requests") {
		t.Error("dashboard body missing Pooled request metrics")
	}
	if !strings.Contains(w.Body.String(), "Run Naive") {
		t.Error("dashboard body missing Run Naive action")
	}
	if !strings.Contains(w.Body.String(), "Run Pooled") {
		t.Error("dashboard body missing Run Pooled action")
	}
}

func TestMemStatsHandler(t *testing.T) {
	mux := NewMux()
	req := httptest.NewRequest("GET", "/debug/mem", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("memstats status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "NumGC") {
		t.Error("memstats body missing NumGC")
	}
}

func TestNaiveHandler(t *testing.T) {
	mux := NewMux()
	req := httptest.NewRequest("GET", "/naive?size=1024", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("naive status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Naive: Processed 1024 bytes") {
		t.Errorf("naive body = %s", w.Body.String())
	}
}

func TestPooledHandler(t *testing.T) {
	mux := NewMux()
	req := httptest.NewRequest("GET", "/pooled?size=1024", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("pooled status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Pooled: Processed 1024 bytes") {
		t.Errorf("pooled body = %s", w.Body.String())
	}
}
