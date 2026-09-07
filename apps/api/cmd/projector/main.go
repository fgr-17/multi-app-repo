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

	var view *greeting.MongoView
	var err error
	for i := 0; i < 60; i++ {
		dial, cancel := context.WithTimeout(ctx, 3*time.Second)
		view, err = greeting.OpenMongoView(dial, greeting.MongoURL(), greeting.MongoDBName())
		cancel()
		if err == nil {
			break
		}
		log.Printf("projector waiting mongo: %v", err)
		time.Sleep(time.Second)
	}
	if err != nil {
		log.Fatal(err)
	}

	log.Print("projector kafka → mongo")
	if err := greeting.RunProjector(ctx, view); err != nil && ctx.Err() == nil {
		log.Fatal(err)
	}
}
