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
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	writes, err := waitEventStore(ctx)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer writes.Close()

	var reads *greeting.MongoView
	mongoCtx, mongoCancel := context.WithTimeout(context.Background(), 10*time.Second)
	reads, err = greeting.OpenMongoView(mongoCtx, greeting.MongoURL(), greeting.MongoDBName())
	mongoCancel()
	if err != nil {
		log.Printf("mongo read model unavailable, GET will replay events: %v", err)
		reads = nil
	}

	store := greeting.ReadThrough{Writes: writes, Reads: reads}

	addr := os.Getenv("PORT")
	if addr == "" {
		addr = "8080"
	}

	mux := http.NewServeMux()
	mux.Handle("GET /health", greeting.Health(store))
	mux.Handle("GET /api/events", greeting.EventsHandler(writes.List))
	mux.Handle("/", greeting.Handler(store))

	log.Printf("api listening on :%s (event store + cqrs)", addr)
	if err := http.ListenAndServe(":"+addr, httputil.CORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func waitEventStore(ctx context.Context) (*greeting.EventStore, error) {
	var last error
	for {
		store, err := greeting.OpenEventStore(ctx, greeting.DatabaseURL())
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
