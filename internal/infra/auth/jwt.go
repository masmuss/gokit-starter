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

// TokenIssuer issues access tokens.
type TokenIssuer interface {
	Issue(ctx context.Context, userID uuid.UUID, orgID uuid.UUID, email string) (string, error)
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
}

type jwtClaims struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	Email          string `json:"email,omitempty"`
	jwt.RegisteredClaims
}

// JWTManager signs and verifies JWT access tokens.
type JWTManager struct {
	issuer string
	secret []byte
	ttl    time.Duration
}

// NewJWTManager creates a new JWT manager.
func NewJWTManager(secret, issuer string, ttl time.Duration) *JWTManager {
	if ttl <= 0 {
		ttl = time.Hour
	}

	return &JWTManager{
		issuer: issuer,
		secret: []byte(secret),
		ttl:    ttl,
	}
}

// NewJWTManagerFromConfig creates a JWTManager from config.
func NewJWTManagerFromConfig(cfg *config.Config) *JWTManager {
	return NewJWTManager(
		cfg.Auth.JWTSecret,
		cfg.Auth.JWTIssuer,
		time.Duration(cfg.Auth.JWTTTL)*time.Minute,
	)
}

// Issue returns a signed JWT access token.
func (m *JWTManager) Issue(_ context.Context, userID uuid.UUID, orgID uuid.UUID, email string) (string, error) {
	now := time.Now().UTC()
	claims := jwtClaims{
		UserID:         userID.String(),
		OrganizationID: orgID.String(),
		Email:          email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

// Verify parses and validates a JWT access token.
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

	return Claims{
		UserID:         userID,
		OrganizationID: orgID,
		Email:          claims.Email,
	}, nil
}
