package greeting

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

func RunProjector(ctx context.Context, view *MongoView) error {
	reader := NewReader("hola-projector")
	defer reader.Close()
	log.Printf("projector listening topic=%s", KafkaTopic())
	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		var evt Event
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			log.Printf("projector skip: %v", err)
			if err := reader.CommitMessages(ctx, msg); err != nil {
				return err
			}
			continue
		}
		if err := view.Apply(evt); err != nil {
			return err
		}
		if err := reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
		log.Printf("projected %s v%d", evt.Type, evt.Version)
	}
}

func DecodeEvent(msg kafka.Message) (Event, error) {
	var evt Event
	err := json.Unmarshal(msg.Value, &evt)
	return evt, err
}
