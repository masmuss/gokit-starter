package domain

import "github.com/google/uuid"

// Organization represents the user's organization.
type Organization struct {
	ID     uuid.UUID
	Name   string
	Code   string
	Type   string
	Status string
}
