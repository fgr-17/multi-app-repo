package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hola/api/internal/greeting"
	"github.com/hola/api/internal/httputil"
)

func TestHealthAndCORS(t *testing.T) {
	store := greeting.NewMemory(greeting.Record{Name: "Mundo"})
	mux := http.NewServeMux()
	mux.Handle("GET /health", greeting.Health(store))

	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	rec := httptest.NewRecorder()
	httputil.CORS(mux).ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS status %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("missing CORS header")
	}
}
