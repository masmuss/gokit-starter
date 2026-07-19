package middleware

import (
	"net/http"
	"slices"

	"github.com/masmuss/gokit-starter/internal/delivery/response"
	"github.com/masmuss/gokit-starter/internal/infra/auth"
)

// RequireRole returns a middleware that checks the authenticated user has one of the given roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := auth.ClaimsFromContext(r.Context())
			if !ok {
				_ = response.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required", nil)
				return
			}

			if !slices.Contains(roles, claims.Role) {
				_ = response.WriteError(w, http.StatusForbidden, "forbidden", "insufficient permissions", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
