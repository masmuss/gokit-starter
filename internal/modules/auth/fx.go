// Package auth provides the Fx module for the auth feature.
package auth

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/masmuss/gokit-starter/internal/delivery"
	delivery_middleware "github.com/masmuss/gokit-starter/internal/delivery/middleware"
	infraauth "github.com/masmuss/gokit-starter/internal/infra/auth"
	"github.com/masmuss/gokit-starter/internal/modules/auth/app"
	"github.com/masmuss/gokit-starter/internal/modules/auth/handler"
	"github.com/masmuss/gokit-starter/internal/modules/auth/infra"
)

// Module groups auth dependencies for Fx.
var Module = fx.Module("auth",
	fx.Provide(
		infra.NewRepositoryFromDB,
		fx.Annotate(func(r *infra.Repository) app.Repository { return r }),
		fx.Annotate(
			func(h *infraauth.BcryptHasher) app.PasswordHasher { return h },
		),
		fx.Annotate(
			func(m *infraauth.JWTManager) app.TokenIssuer { return m },
		),
		app.New,
		fx.Annotate(
			func(s *app.Service) handler.AuthService { return s },
		),
		handler.NewAuthHandler,
		fx.Annotate(
			func(h *handler.AuthHandler, m *delivery_middleware.AuthMiddleware) delivery.RouteRegistrar {
				return delivery.RouteRegistrarFunc(func(r chi.Router) {
					h.RegisterRoutes(r, m)
				})
			},
			fx.ResultTags(`group:"routes"`),
		),
	),
)
