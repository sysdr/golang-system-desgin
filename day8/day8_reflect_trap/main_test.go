package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPopulateUserReflect(t *testing.T) {
	user, err := PopulateUserReflect(testMap)
	if err != nil {
		t.Fatalf("PopulateUserReflect failed: %v", err)
	}
	if user.ID != 123 || user.Name != "Alice Wonderland" || user.Email != "alice@example.com" {
		t.Errorf("PopulateUserReflect produced incorrect user: %+v", user)
	}
}

func TestPopulateUserDirect(t *testing.T) {
	user, err := PopulateUserDirect(testMap)
	if err != nil {
		t.Fatalf("PopulateUserDirect failed: %v", err)
	}
	if user.ID != 123 || user.Name != "Alice Wonderland" || user.Email != "alice@example.com" {
		t.Errorf("PopulateUserDirect produced incorrect user: %+v", user)
	}
}

func BenchmarkPopulateUserReflect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = PopulateUserReflect(testMap)
	}
}

func BenchmarkPopulateUserDirect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = PopulateUserDirect(testMap)
	}
}

func TestDashboardHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	w := httptest.NewRecorder()
	dashboardHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("dashboard status = %d, want 200", w.Code)
	}
	if w.Body.Len() == 0 {
		t.Error("dashboard body empty")
	}
	if !strings.Contains(w.Body.String(), "Reflect vs Direct") {
		t.Error("dashboard body missing title")
	}
	if !strings.Contains(w.Body.String(), "Reflect path") {
		t.Error("dashboard body missing Reflect metric")
	}
	if !strings.Contains(w.Body.String(), "Direct path") {
		t.Error("dashboard body missing Direct metric")
	}
	if !strings.Contains(w.Body.String(), "Application workflow") {
		t.Error("dashboard body missing application workflow")
	}
	if !strings.Contains(w.Body.String(), "What is this application") {
		t.Error("dashboard body missing project explanation")
	}
}
