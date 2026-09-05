package greeting

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetAndPutLWW(t *testing.T) {
	t0 := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	store := NewMemory(Record{Name: "Mundo", UpdatedAt: t0, UpdatedBy: "seed", Version: 1})
	h := Handler(store)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/hello", nil))
	if rec.Code != 200 {
		t.Fatalf("GET status %d", rec.Code)
	}
	var got Record
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Name != "Mundo" {
		t.Fatalf("got %q", got.Name)
	}

	newer, _ := json.Marshal(putBody{
		Name:      "Ada",
		UpdatedAt: t0.Add(time.Minute),
		UpdatedBy: "web",
	})
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/hello", bytes.NewReader(newer)))
	if rec.Code != 200 {
		t.Fatalf("PUT newer status %d body %s", rec.Code, rec.Body.Bytes())
	}
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Name != "Ada" || got.Version != 2 {
		t.Fatalf("after newer: %+v", got)
	}

	older, _ := json.Marshal(putBody{
		Name:      "Viejo",
		UpdatedAt: t0,
		UpdatedBy: "mobile",
	})
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/hello", bytes.NewReader(older)))
	if rec.Code != http.StatusConflict {
		t.Fatalf("PUT older status %d, want 409", rec.Code)
	}
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Name != "Ada" {
		t.Fatalf("older write should not replace: %+v", got)
	}
}
