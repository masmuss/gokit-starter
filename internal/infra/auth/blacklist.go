package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/masmuss/gokit-starter/internal/infra/cache"
)

const blacklistKeyPrefix = "token_bl"

// TokenBlacklist manages revoked tokens.
type TokenBlacklist struct {
	cache cache.Cache
}

// NewTokenBlacklist creates a new token blacklist.
func NewTokenBlacklist(c cache.Cache) *TokenBlacklist {
	return &TokenBlacklist{cache: c}
}

// Blacklist revokes a token until it expires.
func (b *TokenBlacklist) Blacklist(ctx context.Context, jti string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}

	return b.cache.Set(ctx, blacklistKey(jti), "1", ttl)
}

// IsBlacklisted checks whether a token has been revoked.
func (b *TokenBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	var val string
	if err := b.cache.Get(ctx, blacklistKey(jti), &val); err != nil {
		return false, nil
	}

	return val != "", nil
}

func blacklistKey(jti string) string {
	return fmt.Sprintf("%s:%s", blacklistKeyPrefix, jti)
}
