// Package infra provides auth infrastructure implementations.
package infra

import (
	"context"
	"strings"

	"github.com/masmuss/gokit-starter/internal/modules/auth/domain"
)

const (
	demoEmail    = "admin@gokitstarter.local"
	demoPassword = "password123"
	demoToken    = "demo-access-token"
)

// DemoAuthenticator accepts a fixed demo credential pair.
type DemoAuthenticator struct{}

// NewDemoAuthenticator creates a demo authenticator for starter projects.
func NewDemoAuthenticator() *DemoAuthenticator {
	return &DemoAuthenticator{}
}

// Authenticate validates the provided credentials against the demo account.
func (a *DemoAuthenticator) Authenticate(_ context.Context, credentials domain.Credentials) (domain.Session, error) {
	if !strings.EqualFold(credentials.Email, demoEmail) || credentials.Password != demoPassword {
		return domain.Session{}, domain.ErrInvalidCredentials
	}

	return domain.Session{
		AccessToken: demoToken,
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}, nil
}
