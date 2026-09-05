package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/hola/api/internal/httputil"
)

type helloResponse struct {
	Name string `json:"name"`
}

func main() {
	name := os.Getenv("GREETING_NAME")
	if name == "" {
		name = "Mundo"
	}

	addr := os.Getenv("PORT")
	if addr == "" {
		addr = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /api/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(helloResponse{Name: name})
	})

	log.Printf("api listening on :%s (name=%q)", addr, name)
	if err := http.ListenAndServe(":"+addr, httputil.CORS(mux)); err != nil {
		log.Fatal(err)
	}
}
