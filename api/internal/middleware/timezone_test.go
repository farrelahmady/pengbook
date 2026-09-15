package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func offsetOf(loc *time.Location) int {
	_, offset := time.Now().In(loc).Zone()
	return offset
}

func TestLocationFromContext_DefaultUTC(t *testing.T) {
	loc := LocationFromContext(context.Background())
	if loc == nil {
		t.Fatal("expected non-nil location")
	}
	if offset := offsetOf(loc); offset != 0 {
		t.Fatalf("expected UTC fallback, got offset %d", offset)
	}
}

func TestTimezoneMiddleware(t *testing.T) {
	handler := Timezone(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loc := LocationFromContext(r.Context())
		if loc == nil {
			t.Error("expected non-nil location in handler")
		}
		if _, err := w.Write([]byte(loc.String())); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))

	// Valid zone is honored.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(TimezoneHeader, "Asia/Jakarta")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if got := w.Body.String(); got != "Asia/Jakarta" {
		t.Fatalf("expected Asia/Jakarta, got %q", got)
	}

	// Missing header falls back to UTC.
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if got := w.Body.String(); got != "UTC" {
		t.Fatalf("expected UTC fallback, got %q", got)
	}

	// Unknown zone falls back to UTC instead of failing.
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(TimezoneHeader, "Bukan/Zona")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if got := w.Body.String(); got != "UTC" {
		t.Fatalf("expected UTC fallback for unknown zone, got %q", got)
	}
}
