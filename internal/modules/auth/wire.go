// Package auth provides the auth module wiring.
package auth

import (
	"log/slog"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/masmuss/gokit-starter/internal/delivery"
	deliverymiddleware "github.com/masmuss/gokit-starter/internal/delivery/middleware"
	"github.com/masmuss/gokit-starter/internal/infra/auth"
	"github.com/masmuss/gokit-starter/internal/infra/cache"
	"github.com/masmuss/gokit-starter/internal/infra/database"
	"github.com/masmuss/gokit-starter/internal/modules/auth/app"
	"github.com/masmuss/gokit-starter/internal/modules/auth/handler"
	"github.com/masmuss/gokit-starter/internal/modules/auth/infra"
	"github.com/masmuss/gokit-starter/internal/pkg/audit"
	"github.com/masmuss/gokit-starter/internal/pkg/doc"
)

// Dependencies holds everything the auth module needs from the outside.
type Dependencies struct {
	DB             *database.DB
	CacheStore     cache.Cache
	PasswordHasher auth.PasswordHasher
	JWTManager     *auth.JWTManager
	Log            *slog.Logger
	Audit          *audit.Logger
	Validator      *validator.Validate
	AccessTTL      int
	RefreshTTL     int
}

// Module exposes the auth module's public outputs.
type Module struct {
	Handler      *handler.AuthHandler
	Registrar    delivery.RouteRegistrar
	DocRegistrar doc.OperationRegistrar
	Middleware   *deliverymiddleware.AuthMiddleware
}

// Wire builds the auth module from its dependencies.
func Wire(deps Dependencies) Module {
	tokenVerifier := auth.TokenVerifier(deps.JWTManager)
	tokenIssuer := auth.TokenIssuer(deps.JWTManager)
	refreshIssuer := auth.RefreshTokenIssuer(deps.JWTManager)

	blacklist := auth.NewTokenBlacklist(deps.CacheStore)

	repo := infra.NewRepositoryFromDB(deps.DB)
	var repoInterface app.Repository = repo

	svc := app.New(
		repoInterface,
		deps.PasswordHasher,
		tokenIssuer,
		refreshIssuer,
		tokenVerifier,
		blacklist,
		deps.AccessTTL,
		deps.RefreshTTL,
	)

	log := deps.Log.With("module", "auth")

	authHandler := handler.NewAuthHandler(
		handler.AuthService(svc),
		log,
		deps.Audit,
		deps.Validator,
		tokenVerifier,
	)

	authMiddleware := deliverymiddleware.NewAuthMiddleware(
		tokenVerifier, blacklist, log, deps.Audit,
	)

	routeRegistrar := delivery.RouteRegistrarFunc(func(r chi.Router) {
		authHandler.RegisterRoutes(r, authMiddleware)
	})

	return Module{
		Handler:      authHandler,
		Registrar:    routeRegistrar,
		DocRegistrar: handler.NewAuthDocRegistrar(),
		Middleware:   authMiddleware,
	}
}
