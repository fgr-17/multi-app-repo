package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/hola/api/internal/greeting"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var store *greeting.EventStore
	var err error
	for i := 0; i < 60; i++ {
		dial, cancel := context.WithTimeout(ctx, 3*time.Second)
		store, err = greeting.OpenEventStore(dial, greeting.DatabaseURL())
		cancel()
		if err == nil {
			break
		}
		log.Printf("relay waiting postgres: %v", err)
		time.Sleep(time.Second)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	log.Print("relay outbox → kafka")
	if err := greeting.RunRelay(ctx, store); err != nil && ctx.Err() == nil {
		log.Fatal(err)
	}
}
