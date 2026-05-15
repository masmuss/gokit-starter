// Package middleware contains HTTP middleware for the delivery layer.
package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/masmuss/gokit-starter/internal/delivery/response"
	platformauth "github.com/masmuss/gokit-starter/internal/platform/auth"
)

// AuthMiddleware validates bearer tokens and stores claims in the request context.
type AuthMiddleware struct {
	log      *slog.Logger
	verifier platformauth.TokenVerifier
}

// NewAuthMiddleware creates a new auth middleware.
func NewAuthMiddleware(verifier platformauth.TokenVerifier, log *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		log:      log,
		verifier: verifier,
	}
}

// Require ensures the request has a valid bearer token.
func (m *AuthMiddleware) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			_ = response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing bearer token", nil)
			return
		}

		claims, err := m.verifier.Verify(token)
		if err != nil {
			if m.log != nil {
				m.log.DebugContext(r.Context(), "invalid bearer token", "error", err)
			}

			_ = response.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid bearer token", nil)
			return
		}

		next.ServeHTTP(w, r.WithContext(platformauth.WithClaims(r.Context(), claims)))
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
