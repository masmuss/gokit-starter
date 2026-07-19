// Package middleware contains HTTP middleware for the delivery layer.
package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/masmuss/gokit-starter/internal/delivery/response"
	"github.com/masmuss/gokit-starter/internal/infra/authtoken"
	"github.com/masmuss/gokit-starter/internal/pkg/audit"
)

// AuthMiddleware validates bearer tokens and stores claims in the request context.
type AuthMiddleware struct {
	log       *slog.Logger
	audit     *audit.Logger
	verifier  authtoken.TokenVerifier
	blacklist *authtoken.TokenBlacklist
}

// NewAuthMiddleware creates a new auth middleware.
func NewAuthMiddleware(
	verifier authtoken.TokenVerifier,
	blacklist *authtoken.TokenBlacklist,
	log *slog.Logger,
	audit *audit.Logger,
) *AuthMiddleware {
	return &AuthMiddleware{
		log:       log,
		audit:     audit,
		verifier:  verifier,
		blacklist: blacklist,
	}
}

// Require ensures the request has a valid, non-revoked bearer token.
func (m *AuthMiddleware) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			m.audit.Warn("auth.token_missing", "failed", audit.IP(r))
			_ = response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing bearer token", nil)
			return
		}

		claims, err := m.verifier.Verify(token)
		if err != nil {
			m.audit.Warn("auth.token_invalid", "failed", audit.IP(r))
			_ = response.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid bearer token", nil)
			return
		}

		if m.blacklist != nil {
			blacklisted, checkErr := m.blacklist.IsBlacklisted(r.Context(), claims.TokenID())
			if checkErr != nil && m.log != nil {
				m.log.WarnContext(r.Context(), "blacklist check failed, blocking", "error", checkErr)
			}

			if blacklisted || checkErr != nil {
				m.audit.Warn("auth.token_revoked", "failed",
					audit.UserID(claims.UserID), audit.OrgID(claims.OrganizationID), audit.IP(r))
				_ = response.WriteError(w, http.StatusUnauthorized, "unauthorized", "token revoked", nil)
				return
			}
		}

		next.ServeHTTP(w, r.WithContext(authtoken.WithClaims(r.Context(), claims)))
	})
}

func bearerToken(header string) (string, bool) {
	if header == "" {
		return "", false
	}

	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	return parts[1], true
}
