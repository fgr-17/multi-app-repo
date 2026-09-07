package greeting

import "testing"

func TestReadThroughFallsBackToWrites(t *testing.T) {
	mem := NewMemory(Record{Name: "Mundo", UpdatedBy: "seed", Version: 1})
	rt := ReadThrough{Writes: mem}
	got, err := rt.Get()
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Mundo" {
		t.Fatalf("got %q", got.Name)
	}
}
