package domain

import "github.com/google/uuid"

// User account status constants.
const (
	UserStatusActive = "active"
)

// User role constants.
const (
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// User represents a user in the system.
type User struct {
	ID           uuid.UUID    `json:"id"`
	Name         string       `json:"name"`
	Email        string       `json:"email"`
	PasswordHash string       `json:"-"`
	Status       string       `json:"status"`
	Role         string       `json:"role"`
	Organization Organization `json:"organization"`
}

// BelongsToOrganization checks if the user is a member of the given organization.
func (u User) BelongsToOrganization(orgID uuid.UUID) bool {
	return u.Organization.ID == orgID
}
