// Package auth provides the Fx module for the auth feature.
package auth

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	"github.com/masmuss/gokit-starter/internal/delivery"
	deliverymiddleware "github.com/masmuss/gokit-starter/internal/delivery/middleware"
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
		app.New,
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

func provideAuthHandler(svc handler.AuthService, log *slog.Logger, v *validator.Validate) *handler.AuthHandler {
	return handler.NewAuthHandler(svc, log.With("module", "auth"), v)
}
