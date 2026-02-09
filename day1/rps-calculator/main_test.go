package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSimpleHandler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(simpleHandler))
	defer server.Close()

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"no delay", "/", http.StatusOK, "OK"},
		{"with delay param", "/?delay_us=0", http.StatusOK, "OK"},
		{"invalid delay", "/?delay_us=invalid", http.StatusOK, "OK"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(server.URL + tt.path)
			if err != nil {
				t.Fatalf("Get: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d; want %d", resp.StatusCode, tt.wantStatus)
			}
			body := make([]byte, 4)
			n, _ := resp.Body.Read(body)
			if n > 0 && string(body[:n]) != tt.wantBody {
				t.Errorf("body = %q; want %q", string(body[:n]), tt.wantBody)
			}
		})
	}
}
