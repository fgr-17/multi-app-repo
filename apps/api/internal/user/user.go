package user

import (
	"time"

	"github.com/google/uuid"
)

// User is the relational identity record. Source of truth: Postgres.
type User struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	GivenName   string    `json:"givenName"`
	FamilyName  string    `json:"familyName"`
	DisplayName string    `json:"displayName"`
	Phone       string    `json:"phone"`
	City        string    `json:"city"`
	Country     string    `json:"country"`
	Locale      string    `json:"locale"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Store interface {
	List() ([]User, error)
	Get(id uuid.UUID) (User, error)
	Create(u User) (User, error)
	Ping() error
}
