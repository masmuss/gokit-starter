// Package domain contains auth domain models.
package domain

// Credentials contains the data required to authenticate a user.
type Credentials struct {
	Email    string
	Password string
}
