package domain

import "github.com/google/uuid"

// Profile represents public user data.
type Profile struct {
	ID           uuid.UUID
	Name         string
	Email        string
	Status       string
	Organization Organization
}
