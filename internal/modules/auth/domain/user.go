package domain

import "github.com/google/uuid"

// User represents a user in the system.
type User struct {
	ID           uuid.UUID    `json:"id"`
	Name         string       `json:"name"`
	Email        string       `json:"email"`
	PasswordHash string       `json:"-"`
	Status       string       `json:"status"`
	Organization Organization `json:"organization"`
}
