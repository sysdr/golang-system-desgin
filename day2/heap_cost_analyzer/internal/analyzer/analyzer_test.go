package analyzer

import "testing"

func TestProcessRequestPointer(t *testing.T) {
	stats := ProcessRequestPointer(42)
	if stats == nil {
		t.Fatal("ProcessRequestPointer returned nil")
	}
	if stats.ID != 42 {
		t.Errorf("ID = %d, want 42", stats.ID)
	}
	if stats.Status != 200 {
		t.Errorf("Status = %d, want 200", stats.Status)
	}
}

func TestProcessRequestValue(t *testing.T) {
	stats := ProcessRequestValue(99)
	if stats.ID != 99 {
		t.Errorf("ID = %d, want 99", stats.ID)
	}
	if stats.Status != 200 {
		t.Errorf("Status = %d, want 200", stats.Status)
	}
}
