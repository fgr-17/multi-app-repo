package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hola/api/internal/greeting"
	"github.com/hola/api/internal/httputil"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	store, err := waitForPostgres(ctx, greeting.DatabaseURL())
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer store.Close()

	addr := os.Getenv("PORT")
	if addr == "" {
		addr = "8080"
	}

	mux := http.NewServeMux()
	mux.Handle("GET /health", greeting.Health(store))
	mux.Handle("/", greeting.Handler(store))

	log.Printf("api listening on :%s (postgres)", addr)
	if err := http.ListenAndServe(":"+addr, httputil.CORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func waitForPostgres(ctx context.Context, url string) (*greeting.Postgres, error) {
	var last error
	for {
		store, err := greeting.OpenPostgres(ctx, url)
		if err == nil {
			return store, nil
		}
		last = err
		select {
		case <-ctx.Done():
			return nil, last
		case <-time.After(500 * time.Millisecond):
		}
	}
}
