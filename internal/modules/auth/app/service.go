// Package app contains auth use cases.
package app

import (
	"context"
	"fmt"

	"github.com/masmuss/gokit-starter/internal/modules/auth/domain"
)

// Authenticator resolves a session from credentials.
type Authenticator interface {
	Authenticate(context.Context, domain.Credentials) (domain.Session, error)
}

// Service implements auth use cases.
type Service struct {
	authenticator Authenticator
}

// New creates a new auth service.
func New(authenticator Authenticator) *Service {
	return &Service{authenticator: authenticator}
}

// Login authenticates a user and returns a session token.
func (s *Service) Login(ctx context.Context, credentials domain.Credentials) (domain.Session, error) {
	session, err := s.authenticator.Authenticate(ctx, credentials)
	if err != nil {
		return domain.Session{}, fmt.Errorf("authenticate credentials: %w", err)
	}

	return session, nil
}
