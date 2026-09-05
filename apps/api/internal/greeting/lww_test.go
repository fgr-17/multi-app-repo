package greeting

import (
	"testing"
	"time"
)

func TestWins(t *testing.T) {
	older := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := older.Add(time.Minute)

	base := Record{Name: "A", UpdatedAt: older, UpdatedBy: "aaa", Version: 1}

	if !Wins(Record{Name: "B", UpdatedAt: newer, UpdatedBy: "bbb", Version: 1}, base) {
		t.Fatal("newer timestamp should win")
	}
	if Wins(Record{Name: "B", UpdatedAt: older.Add(-time.Second), UpdatedBy: "zzz", Version: 99}, base) {
		t.Fatal("older timestamp should lose")
	}
	if !Wins(Record{Name: "B", UpdatedAt: older, UpdatedBy: "zzz", Version: 1}, base) {
		t.Fatal("same time, higher updatedBy should win")
	}
	if !Wins(Record{Name: "B", UpdatedAt: older, UpdatedBy: "aaa", Version: 2}, base) {
		t.Fatal("same time and device, higher version should win")
	}
}
