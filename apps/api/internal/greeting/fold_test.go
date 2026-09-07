package greeting

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFoldReplaysRenameHistory(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	seed, err := json.Marshal(GreetingPayload{Name: "Mundo", UpdatedAt: t0, UpdatedBy: "seed"})
	if err != nil {
		t.Fatal(err)
	}
	rename, err := json.Marshal(GreetingPayload{Name: "Ada", UpdatedAt: t0.Add(time.Minute), UpdatedBy: "web"})
	if err != nil {
		t.Fatal(err)
	}

	rec, err := Fold([]Event{
		{StreamID: StreamID, Version: 1, Type: TypeGreetingSeeded, Payload: seed},
		{StreamID: StreamID, Version: 2, Type: TypeGreetingRenamed, Payload: rename},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rec.Name != "Ada" || rec.Version != 2 || rec.UpdatedBy != "web" {
		t.Fatalf("folded %+v", rec)
	}
}
