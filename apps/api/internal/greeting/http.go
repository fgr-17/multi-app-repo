package greeting

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type putBody struct {
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updatedAt"`
	UpdatedBy string    `json:"updatedBy"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func Handler(store Store) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/hello", func(w http.ResponseWriter, r *http.Request) {
		rec, err := store.Get()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, rec)
	})

	mux.HandleFunc("PUT /api/hello", func(w http.ResponseWriter, r *http.Request) {
		var body putBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "json inválido")
			return
		}
		name := strings.TrimSpace(body.Name)
		if name == "" {
			writeError(w, http.StatusBadRequest, "name es requerido")
			return
		}
		if body.UpdatedBy == "" {
			writeError(w, http.StatusBadRequest, "updatedBy es requerido")
			return
		}
		if body.UpdatedAt.IsZero() {
			writeError(w, http.StatusBadRequest, "updatedAt es requerido")
			return
		}
		incoming := Record{
			Name:      name,
			UpdatedAt: body.UpdatedAt.UTC(),
			UpdatedBy: body.UpdatedBy,
		}
		current, accepted, err := store.SaveIfNewer(incoming)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		status := http.StatusOK
		if !accepted {
			status = http.StatusConflict
		}
		writeJSON(w, status, current)
	})

	return mux
}

func Health(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := store.Ping(); err != nil {
			writeError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	}
}
