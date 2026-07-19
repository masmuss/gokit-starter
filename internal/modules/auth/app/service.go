// Package app contains auth use cases.
package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	authtoken "github.com/masmuss/gokit-starter/internal/infra/authtoken"
	"github.com/masmuss/gokit-starter/internal/modules/auth/domain"
)

// Repository persists and reads auth data.
type Repository interface {
	CreateAccount(ctx context.Context, input domain.RegisterInput, passwordHash string) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
}

// Service implements auth use cases.
type Service struct {
	repository     Repository
	hasher         authtoken.PasswordHasher
	tokens         authtoken.TokenIssuer
	refreshTokens  authtoken.RefreshTokenIssuer
	tokenVerifier  authtoken.TokenVerifier
	blacklist      *authtoken.TokenBlacklist
	expiresIn      int
	refreshExpires int
}

// New creates a new auth service.
func New(
	repository Repository,
	hasher authtoken.PasswordHasher,
	tokens authtoken.TokenIssuer,
	refreshTokens authtoken.RefreshTokenIssuer,
	tokenVerifier authtoken.TokenVerifier,
	blacklist *authtoken.TokenBlacklist,
	expiresIn int,
	refreshExpires int,
) *Service {
	return &Service{
		repository:     repository,
		hasher:         hasher,
		tokens:         tokens,
		refreshTokens:  refreshTokens,
		tokenVerifier:  tokenVerifier,
		blacklist:      blacklist,
		expiresIn:      expiresIn,
		refreshExpires: refreshExpires,
	}
}

// Register creates a new account and returns its session and profile.
func (s *Service) Register(ctx context.Context, input domain.RegisterInput) (domain.Session, domain.Profile, error) {
	createdUser, err := s.createAccount(ctx, input)
	if err != nil {
		return domain.Session{}, domain.Profile{}, err
	}

	return s.issueSession(ctx, createdUser)
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

	if user.Status != domain.UserStatusActive {
		return domain.Session{}, domain.Profile{}, domain.ErrAccountInactive
	}

	return s.issueSession(ctx, user)
}

// Profile returns the current user's public profile.
func (s *Service) Profile(ctx context.Context, userID, orgID uuid.UUID) (domain.Profile, error) {
	user, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		return domain.Profile{}, err
	}

	if user.Organization.ID != orgID {
		return domain.Profile{}, domain.ErrUserNotFound
	}

	return s.profileFromUser(user), nil
}

// Logout revokes both access and refresh tokens.
func (s *Service) Logout(ctx context.Context, accessClaims, refreshClaims authtoken.Claims) error {
	if s.blacklist != nil {
		if err := s.blacklist.Blacklist(ctx, accessClaims.TokenID(), accessClaims.ExpiresAt()); err != nil {
			return fmt.Errorf("blacklist access token: %w", err)
		}

		if err := s.blacklist.Blacklist(ctx, refreshClaims.TokenID(), refreshClaims.ExpiresAt()); err != nil {
			return fmt.Errorf("blacklist refresh token: %w", err)
		}
	}

	return nil
}

// ChangePassword updates the password for a user after verifying the current one.
func (s *Service) ChangePassword(
	ctx context.Context,
	userID uuid.UUID,
	oldPassword, newPassword string,
) error {
	user, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if err = s.hasher.Compare(user.PasswordHash, oldPassword); err != nil {
		return domain.ErrInvalidCredentials
	}

	passwordHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	if err = s.repository.UpdatePassword(ctx, userID, passwordHash); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}

// RefreshAccessToken validates a refresh token and issues a new access token.
func (s *Service) RefreshAccessToken(ctx context.Context, token string) (domain.Session, error) {
	claims, err := s.tokenVerifier.Verify(token)
	if err != nil {
		return domain.Session{}, fmt.Errorf("%w: %w", domain.ErrInvalidCredentials, err)
	}

	if !claims.IsRefresh() {
		return domain.Session{}, domain.ErrInvalidCredentials
	}

	if s.blacklist != nil {
		blacklisted, checkErr := s.blacklist.IsBlacklisted(ctx, claims.TokenID())
		if checkErr != nil {
			return domain.Session{}, fmt.Errorf("blacklist check: %w", checkErr)
		}

		if blacklisted {
			return domain.Session{}, domain.ErrInvalidCredentials
		}
	}

	accessToken, err := s.tokens.Issue(ctx, claims.UserID, claims.OrganizationID, claims.Email, claims.Role)
	if err != nil {
		return domain.Session{}, fmt.Errorf("issue access token: %w", err)
	}

	return s.session(accessToken), nil
}

func (s *Service) issueSession(ctx context.Context, user domain.User) (domain.Session, domain.Profile, error) {
	accessToken, err := s.tokens.Issue(ctx, user.ID, user.Organization.ID, user.Email, user.Role)
	if err != nil {
		return domain.Session{}, domain.Profile{}, fmt.Errorf("issue access token: %w", err)
	}

	refreshToken, err := s.refreshTokens.IssueRefresh(ctx, user.ID, user.Organization.ID, user.Email, user.Role)
	if err != nil {
		return domain.Session{}, domain.Profile{}, fmt.Errorf("issue refresh token: %w", err)
	}

	return domain.Session{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        s.expiresIn,
		RefreshExpiresIn: s.refreshExpires,
	}, s.profileFromUser(user), nil
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
		Role:   user.Role,
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
