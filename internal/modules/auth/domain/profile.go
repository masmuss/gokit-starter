package domain

import "github.com/google/uuid"

// Profile represents public user data.
type Profile struct {
	ID           uuid.UUID    `json:"id"`
	Name         string       `json:"name"`
	Email        string       `json:"email"`
	Status       string       `json:"status"`
	Organization Organization `json:"organization"`
}
