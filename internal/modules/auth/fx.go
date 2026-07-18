// Package auth provides the Fx module for the auth feature.
package auth

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	"github.com/masmuss/gokit-starter/internal/delivery"
	deliverymiddleware "github.com/masmuss/gokit-starter/internal/delivery/middleware"
	infraauth "github.com/masmuss/gokit-starter/internal/infra/auth"
	"github.com/masmuss/gokit-starter/internal/modules/auth/app"
	"github.com/masmuss/gokit-starter/internal/modules/auth/handler"
	"github.com/masmuss/gokit-starter/internal/modules/auth/infra"
	"github.com/masmuss/gokit-starter/internal/pkg/doc"
)

// Module groups auth dependencies for Fx.
var Module = fx.Module("auth",
	fx.Provide(
		infra.NewRepositoryFromDB,
		fx.Annotate(func(r *infra.Repository) app.Repository { return r }),
		provideAuthService,
		fx.Annotate(
			func(s *app.Service) handler.AuthService { return s },
		),
		provideAuthHandler,
		fx.Annotate(
			func(h *handler.AuthHandler, m *deliverymiddleware.AuthMiddleware) delivery.RouteRegistrar {
				return delivery.RouteRegistrarFunc(func(r chi.Router) {
					h.RegisterRoutes(r, m)
				})
			},
			fx.ResultTags(`group:"routes"`),
		),
		handler.NewAuthDocRegistrar,
		fx.Annotate(
			func(r *handler.AuthDocRegistrar) doc.OperationRegistrar { return r },
			fx.ResultTags(`group:"docRegistrars"`),
		),
	),
)

type authServiceDeps struct {
	fx.In
	Repository     app.Repository
	Hasher         infraauth.PasswordHasher
	Tokens         infraauth.TokenIssuer
	RefreshTokens  infraauth.RefreshTokenIssuer
	TokenVerifier  infraauth.TokenVerifier
	Blacklist      *infraauth.TokenBlacklist
	ExpiresIn      int `name:"authExpiresIn"`
	RefreshExpires int `name:"authRefreshExpiresIn"`
}

func provideAuthService(deps authServiceDeps) *app.Service {
	return app.New(
		deps.Repository,
		deps.Hasher,
		deps.Tokens,
		deps.RefreshTokens,
		deps.TokenVerifier,
		deps.Blacklist,
		deps.ExpiresIn,
		deps.RefreshExpires,
	)
}

type authHandlerDeps struct {
	fx.In
	Service       handler.AuthService
	Log           *slog.Logger
	Validator     *validator.Validate
	TokenVerifier infraauth.TokenVerifier
}

func provideAuthHandler(deps authHandlerDeps) *handler.AuthHandler {
	return handler.NewAuthHandler(
		deps.Service,
		deps.Log.With("module", "auth"),
		deps.Validator,
		deps.TokenVerifier,
	)
}
