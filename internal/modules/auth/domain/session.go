package domain

import "errors"

// ErrInvalidCredentials indicates the provided login details are invalid.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Session contains the token returned after authentication.
type Session struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int
}
