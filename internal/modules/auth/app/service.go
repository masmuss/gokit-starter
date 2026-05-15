// Package app contains auth use cases.
package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/masmuss/gokit-starter/internal/modules/auth/domain"
)

// Repository persists and reads auth data.
type Repository interface {
	CreateAccount(ctx context.Context, input domain.RegisterInput, passwordHash string) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (domain.User, error)
}

// PasswordHasher hashes and compares passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

// TokenIssuer issues access tokens.
type TokenIssuer interface {
	Issue(context.Context, uuid.UUID, string) (string, error)
}

// Service implements auth use cases.
type Service struct {
	repository Repository
	hasher     PasswordHasher
	tokens     TokenIssuer
	expiresIn  int
}

// New creates a new auth service.
func New(repository Repository, hasher PasswordHasher, tokens TokenIssuer, expiresIn int) *Service {
	return &Service{
		repository: repository,
		hasher:     hasher,
		tokens:     tokens,
		expiresIn:  expiresIn,
	}
}

// Register creates a new account and returns its session and profile.
func (s *Service) Register(ctx context.Context, input domain.RegisterInput) (domain.Session, domain.Profile, error) {
	createdUser, err := s.createAccount(ctx, input)
	if err != nil {
		return domain.Session{}, domain.Profile{}, err
	}

	token, err := s.tokens.Issue(ctx, createdUser.ID, createdUser.Email)
	if err != nil {
		return domain.Session{}, domain.Profile{}, fmt.Errorf("issue token: %w", err)
	}

	return s.session(token), s.profileFromUser(createdUser), nil
}

// Login authenticates a user and returns a session token.
func (s *Service) Login(ctx context.Context, credentials domain.Credentials) (domain.Session, domain.Profile, error) {
	user, err := s.repository.FindByEmail(ctx, normalizeEmail(credentials.Email))
	if err != nil {
		return domain.Session{}, domain.Profile{}, err
	}

	if err = s.hasher.Compare(user.PasswordHash, credentials.Password); err != nil {
		return domain.Session{}, domain.Profile{}, domain.ErrInvalidCredentials
	}

	token, err := s.tokens.Issue(ctx, user.ID, user.Email)
	if err != nil {
		return domain.Session{}, domain.Profile{}, fmt.Errorf("issue token: %w", err)
	}

	return s.session(token), s.profileFromUser(user), nil
}

// Profile returns the current user's public profile.
func (s *Service) Profile(ctx context.Context, userID uuid.UUID) (domain.Profile, error) {
	user, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		return domain.Profile{}, err
	}

	return s.profileFromUser(user), nil
}

func (s *Service) createAccount(ctx context.Context, input domain.RegisterInput) (domain.User, error) {
	passwordHash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repository.CreateAccount(ctx, domain.RegisterInput{
		Name:             strings.TrimSpace(input.Name),
		Email:            normalizeEmail(input.Email),
		Password:         input.Password,
		OrganizationName: strings.TrimSpace(input.OrganizationName),
	}, passwordHash)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (s *Service) session(token string) domain.Session {
	return domain.Session{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   s.expiresIn,
	}
}

func (s *Service) profileFromUser(user domain.User) domain.Profile {
	return domain.Profile{
		ID:     user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Status: user.Status,
		Organization: domain.Organization{
			ID:     user.Organization.ID,
			Name:   user.Organization.Name,
			Code:   user.Organization.Code,
			Type:   user.Organization.Type,
			Status: user.Organization.Status,
		},
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
