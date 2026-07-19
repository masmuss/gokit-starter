// Package model provides GORM database models.
package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the database.
type User struct {
	ID             uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrganizationID uuid.UUID    `gorm:"type:uuid;not null;index"`
	OutletID       *uuid.UUID   `gorm:"type:uuid"`
	Name           string       `gorm:"size:128;not null"`
	Email          string       `gorm:"size:128;not null;uniqueIndex"`
	PasswordHash   string       `gorm:"size:255;not null"`
	Status         string       `gorm:"size:32;not null;default:active"`
	Role           string       `gorm:"size:32;not null;default:member"`
	CreatedAt      time.Time    `gorm:"autoCreateTime"`
	UpdatedAt      time.Time    `gorm:"autoUpdateTime"`
	Organization   Organization `gorm:"foreignKey:OrganizationID"`
}

// TableName overrides the default table name.
func (User) TableName() string {
	return "users"
}

// Organization represents an organization in the database.
type Organization struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ParentID  *uuid.UUID `gorm:"type:uuid"`
	Name      string     `gorm:"size:128;not null"`
	Code      string     `gorm:"size:16;not null;uniqueIndex"`
	Type      string     `gorm:"size:32;not null"`
	Status    string     `gorm:"size:32;not null"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
	Users     []User     `gorm:"foreignKey:OrganizationID"`
}

// TableName overrides the default table name.
func (Organization) TableName() string {
	return "organizations"
}
