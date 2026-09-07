package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type createBody struct {
	Email       string `json:"email"`
	GivenName   string `json:"givenName"`
	FamilyName  string `json:"familyName"`
	DisplayName string `json:"displayName"`
	Phone       string `json:"phone"`
	City        string `json:"city"`
	Country     string `json:"country"`
	Locale      string `json:"locale"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func Handler(store Store) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/users", func(w http.ResponseWriter, r *http.Request) {
		users, err := store.List()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, users)
	})

	mux.HandleFunc("GET /api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id inválido"})
			return
		}
		u, err := store.Get(id)
		if errors.Is(err, ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "no existe"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, u)
	})

	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		var body createBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "json inválido"})
			return
		}
		email := strings.TrimSpace(strings.ToLower(body.Email))
		given := strings.TrimSpace(body.GivenName)
		family := strings.TrimSpace(body.FamilyName)
		if email == "" || given == "" || family == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email, givenName y familyName son requeridos"})
			return
		}
		u, err := store.Create(User{
			Email:       email,
			GivenName:   given,
			FamilyName:  family,
			DisplayName: strings.TrimSpace(body.DisplayName),
			Phone:       strings.TrimSpace(body.Phone),
			City:        strings.TrimSpace(body.City),
			Country:     strings.TrimSpace(body.Country),
			Locale:      strings.TrimSpace(body.Locale),
		})
		if err != nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, u)
	})

	return mux
}
