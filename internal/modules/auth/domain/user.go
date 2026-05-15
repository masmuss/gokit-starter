package domain

import "github.com/google/uuid"

// User represents the authenticated account with a password hash.
type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	Status       string
	Organization Organization
}
