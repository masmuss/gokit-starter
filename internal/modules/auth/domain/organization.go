package domain

import "github.com/google/uuid"

// Organization represents a user's organization.
type Organization struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Code   string    `json:"code"`
	Type   string    `json:"type"`
	Status string    `json:"status"`
}
