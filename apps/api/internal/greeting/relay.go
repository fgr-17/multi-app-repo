package greeting

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

func KafkaBrokers() []string {
	raw := os.Getenv("KAFKA_BROKERS")
	if raw == "" {
		raw = "localhost:9092"
	}
	return strings.Split(raw, ",")
}

func KafkaTopic() string {
	if v := os.Getenv("KAFKA_TOPIC"); v != "" {
		return v
	}
	return "greeting.events"
}

func NewWriter() *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(KafkaBrokers()...),
		Topic:    KafkaTopic(),
		Balancer: &kafka.Hash{},
	}
}

func NewReader(group string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  KafkaBrokers(),
		Topic:    KafkaTopic(),
		GroupID:  group,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
}

// PublishOutbox copies unpublished events into Kafka and marks them done.
func PublishOutbox(ctx context.Context, store *EventStore, writer *kafka.Writer) error {
	events, err := store.Unpublished(ctx, 50)
	if err != nil {
		return err
	}
	for _, evt := range events {
		body, err := json.Marshal(evt)
		if err != nil {
			return err
		}
		err = writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(evt.StreamID),
			Value: body,
		})
		if err != nil {
			return err
		}
		if err := store.MarkPublished(ctx, evt.StreamID, evt.Version); err != nil {
			return err
		}
		log.Printf("relay %s v%d %s", evt.Type, evt.Version, evt.StreamID)
	}
	return nil
}

func RunRelay(ctx context.Context, store *EventStore) error {
	writer := NewWriter()
	defer writer.Close()
	tick := time.NewTicker(500 * time.Millisecond)
	defer tick.Stop()
	for {
		if err := PublishOutbox(ctx, store, writer); err != nil {
			log.Printf("relay: %v", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
		}
	}
}
