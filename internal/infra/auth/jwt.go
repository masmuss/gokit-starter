package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/masmuss/gokit-starter/internal/config"
)

// ErrInvalidToken indicates the token is invalid.
var ErrInvalidToken = errors.New("invalid token")

// Token type constants.
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// TokenIssuer issues access tokens.
type TokenIssuer interface {
	Issue(ctx context.Context, userID uuid.UUID, orgID uuid.UUID, email, role string) (string, error)
}

// RefreshTokenIssuer issues refresh tokens.
type RefreshTokenIssuer interface {
	IssueRefresh(ctx context.Context, userID uuid.UUID, orgID uuid.UUID, email, role string) (string, error)
}

// TokenVerifier verifies access tokens.
type TokenVerifier interface {
	Verify(token string) (Claims, error)
}

// Claims contains the authenticated identity.
type Claims struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	Email          string
	Role           string
	TokenType      string
	jti            string
	expiresAt      time.Time
}

// TokenID extracts the JWT ID from claims for token revocation.
func (c Claims) TokenID() string {
	return c.jti
}

// ExpiresAt returns the token expiration time.
func (c Claims) ExpiresAt() time.Time {
	return c.expiresAt
}

// IsAccess returns true for access tokens.
func (c Claims) IsAccess() bool {
	return c.TokenType == TokenTypeAccess
}

// IsRefresh returns true for refresh tokens.
func (c Claims) IsRefresh() bool {
	return c.TokenType == TokenTypeRefresh
}

type jwtClaims struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	Email          string `json:"email,omitempty"`
	Role           string `json:"role"`
	TokenType      string `json:"token_type"`
	jwt.RegisteredClaims
}

// JWTManager signs and verifies JWT tokens.
type JWTManager struct {
	issuer     string
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewJWTManager creates a new JWT manager.
func NewJWTManager(secret, issuer string, accessTTL, refreshTTL time.Duration) *JWTManager {
	if accessTTL <= 0 {
		accessTTL = time.Hour
	}
	if refreshTTL <= 0 {
		refreshTTL = 24 * time.Hour * 7
	}

	return &JWTManager{
		issuer:     issuer,
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// NewJWTManagerFromConfig creates a JWTManager from config.
func NewJWTManagerFromConfig(cfg *config.Config) *JWTManager {
	return NewJWTManager(
		cfg.Auth.JWTSecret,
		cfg.Auth.JWTIssuer,
		time.Duration(cfg.Auth.JWTTTL)*time.Minute,
		time.Duration(cfg.Auth.JWTRefreshTTL)*time.Minute,
	)
}

// Issue returns a signed JWT access token.
func (m *JWTManager) Issue(
	_ context.Context,
	userID uuid.UUID,
	orgID uuid.UUID,
	email, role string,
) (string, error) {
	return m.issue(userID, orgID, email, role, TokenTypeAccess, m.accessTTL)
}

// IssueRefresh returns a signed JWT refresh token.
func (m *JWTManager) IssueRefresh(
	_ context.Context,
	userID uuid.UUID,
	orgID uuid.UUID,
	email, role string,
) (string, error) {
	return m.issue(userID, orgID, email, role, TokenTypeRefresh, m.refreshTTL)
}

func (m *JWTManager) issue(
	userID uuid.UUID,
	orgID uuid.UUID,
	email, role string,
	tokenType string,
	ttl time.Duration,
) (string, error) {
	now := time.Now().UTC()
	claims := jwtClaims{
		UserID:         userID.String(),
		OrganizationID: orgID.String(),
		Email:          email,
		Role:           role,
		TokenType:      tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID.String(),
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

// Verify parses and validates a JWT token.
func (m *JWTManager) Verify(token string) (Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}

		return m.secret, nil
	})
	if err != nil {
		return Claims{}, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	claims, ok := parsed.Claims.(*jwtClaims)
	if !ok || !parsed.Valid {
		return Claims{}, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return Claims{}, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	orgID, err := uuid.Parse(claims.OrganizationID)
	if err != nil {
		return Claims{}, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	var expiresAt time.Time
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	}

	return Claims{
		UserID:         userID,
		OrganizationID: orgID,
		Email:          claims.Email,
		Role:           claims.Role,
		TokenType:      claims.TokenType,
		jti:            claims.ID,
		expiresAt:      expiresAt,
	}, nil
}
