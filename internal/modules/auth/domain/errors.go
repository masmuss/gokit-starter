package domain

import "errors"

var (
	// ErrInvalidCredentials indicates the provided login details are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrEmailAlreadyUsed indicates the email is already registered.
	ErrEmailAlreadyUsed = errors.New("email already used")
	// ErrUserNotFound indicates the user could not be found.
	ErrUserNotFound = errors.New("user not found")
	// ErrAccountInactive indicates the account is not active.
	ErrAccountInactive = errors.New("account is inactive")
	// ErrOrgCodeCollision indicates organization code collision after max retries
	ErrOrgCodeCollision = errors.New("organization code collision after max retries")
)
