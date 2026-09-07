package user

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
)

type memory struct {
	mu   sync.Mutex
	byID map[uuid.UUID]User
}

func newMemory(seed []User) *memory {
	m := &memory{byID: map[uuid.UUID]User{}}
	for _, u := range seed {
		if u.ID == uuid.Nil {
			u.ID = uuid.New()
		}
		m.byID[u.ID] = u
	}
	return m
}

func (m *memory) List() ([]User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]User, 0, len(m.byID))
	for _, u := range m.byID {
		out = append(out, u)
	}
	return out, nil
}

func (m *memory) Get(id uuid.UUID) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byID[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}

func (m *memory) Create(u User) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u.ID = uuid.New()
	if u.DisplayName == "" {
		u.DisplayName = u.GivenName + " " + u.FamilyName
	}
	m.byID[u.ID] = u
	return u, nil
}

func (m *memory) Ping() error { return nil }

func TestUsersHTTP(t *testing.T) {
	fed := User{
		ID:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Email: "rouxfederico@gmail.com", GivenName: "Federico", FamilyName: "Roux",
		DisplayName: "Federico Roux", City: "Buenos Aires", Country: "AR",
	}
	h := Handler(newMemory([]User{fed}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/users", nil))
	if rec.Code != 200 {
		t.Fatalf("list %d", rec.Code)
	}
	var listed []User
	if err := json.NewDecoder(rec.Body).Decode(&listed); err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].Email != "rouxfederico@gmail.com" {
		t.Fatalf("listed %+v", listed)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/users/"+fed.ID.String(), nil))
	if rec.Code != 200 {
		t.Fatalf("get %d", rec.Code)
	}

	body := `{"email":"ada@example.com","givenName":"Ada","familyName":"Lovelace"}`
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body)))
	if rec.Code != 201 {
		t.Fatalf("create %d %s", rec.Code, rec.Body.Bytes())
	}
}
