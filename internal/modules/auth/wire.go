// Package auth provides the auth module wiring.
package auth

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	delivery "github.com/masmuss/gokit-starter/internal/inbound"
	inboundmw "github.com/masmuss/gokit-starter/internal/inbound/middleware"
	"github.com/masmuss/gokit-starter/internal/modules/auth/app"
	"github.com/masmuss/gokit-starter/internal/modules/auth/handler"
	"github.com/masmuss/gokit-starter/internal/modules/auth/repository"
	"github.com/masmuss/gokit-starter/internal/outbound/authtoken"
	"github.com/masmuss/gokit-starter/internal/outbound/cache"
	"github.com/masmuss/gokit-starter/internal/outbound/database"
	"github.com/masmuss/gokit-starter/internal/pkg/audit"
	"github.com/masmuss/gokit-starter/internal/pkg/doc"
)

// Dependencies holds everything the auth module needs from the outside.
type Dependencies struct {
	DB             *database.DB
	CacheStore     cache.Cache
	PasswordHasher authtoken.PasswordHasher
	JWTManager     *authtoken.JWTManager
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
	Middleware   *inboundmw.AuthMiddleware
}

// Wire builds the auth module from its dependencies.
func Wire(deps Dependencies) Module {
	tokenVerifier := authtoken.TokenVerifier(deps.JWTManager)
	tokenIssuer := authtoken.TokenIssuer(deps.JWTManager)
	refreshIssuer := authtoken.RefreshTokenIssuer(deps.JWTManager)

	blacklist := authtoken.NewTokenBlacklist(deps.CacheStore)

	repo := repository.NewRepositoryFromDB(deps.DB)
	var repoInterface app.Repository = repo

	svc := app.New(app.Config{
		Repository:     repoInterface,
		Hasher:         deps.PasswordHasher,
		Tokens:         tokenIssuer,
		RefreshTokens:  refreshIssuer,
		TokenVerifier:  tokenVerifier,
		Blacklist:      blacklist,
		ExpiresIn:      deps.AccessTTL,
		RefreshExpires: deps.RefreshTTL,
	})

	log := deps.Log.With("module", "auth")

	authHandler := handler.NewAuthHandler(
		handler.AuthService(svc),
		log,
		deps.Audit,
		deps.Validator,
		tokenVerifier,
	)

	authMiddleware := inboundmw.NewAuthMiddleware(
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
