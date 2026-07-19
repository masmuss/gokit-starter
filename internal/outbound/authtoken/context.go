package authtoken

import (
	"context"

	"github.com/google/uuid"
)

type claimsKey struct{}

// WithClaims stores auth claims in the context.
func WithClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimsKey{}, claims)
}

// ClaimsFromContext returns auth claims from the context.
func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(claimsKey{}).(Claims)
	return claims, ok
}

// CurrentUserID returns the authenticated user ID from context.
func CurrentUserID(ctx context.Context) (uuid.UUID, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return uuid.UUID{}, false
	}

	return claims.UserID, true
}

// CurrentOrganizationID returns the authenticated organization ID from context.
func CurrentOrganizationID(ctx context.Context) (uuid.UUID, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return uuid.UUID{}, false
	}

	return claims.OrganizationID, true
}
