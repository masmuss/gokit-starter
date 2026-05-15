package app

import (
	"log/slog"
	"net/http"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/masmuss/gokit-starter/docs"
	"github.com/masmuss/gokit-starter/internal/config"
	"github.com/masmuss/gokit-starter/internal/delivery/handler"
	deliverymiddleware "github.com/masmuss/gokit-starter/internal/delivery/middleware"
	authapp "github.com/masmuss/gokit-starter/internal/modules/auth/app"
	authinfra "github.com/masmuss/gokit-starter/internal/modules/auth/infra"
	"github.com/masmuss/gokit-starter/internal/platform/auth"
	"github.com/masmuss/gokit-starter/internal/platform/database"
	"github.com/masmuss/gokit-starter/internal/platform/validation"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// NewRouter builds the HTTP routes and middleware stack.
func NewRouter(cfg *config.Config, db *database.DB, log *slog.Logger) *chi.Mux {
	r := chi.NewRouter()
	healthHandler := handler.NewHealthHandler(cfg.App.Name, log)
	passwordHasher := auth.NewBcryptHasher(cfg.Bcrypt.Rounds)
	tokenManager := auth.NewJWTManager(
		cfg.Auth.JWTSecret,
		cfg.Auth.JWTIssuer,
		time.Duration(cfg.Auth.JWTTTL)*time.Minute,
	)
	authService := authapp.New(
		authinfra.NewRepository(db.Client),
		passwordHasher,
		tokenManager,
		cfg.Auth.JWTTTL*60,
	)
	authHandler := handler.NewAuthHandler(authService, log, validation.New())
	authMiddleware := deliverymiddleware.NewAuthMiddleware(tokenManager, log)

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/health", healthHandler.Handle)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.With(authMiddleware.Require).Get("/profile", authHandler.Profile)
	})
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusMovedPermanently)
	})
	r.Get("/docs/*", httpSwagger.WrapHandler)

	docs.SwaggerInfo.Title = cfg.App.Name
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Version = "0.1.0"
	docs.SwaggerInfo.Description = "Boilerplate API starter with Chi, Ent, and JWT auth."
	docs.SwaggerInfo.Host = ""
	docs.SwaggerInfo.Schemes = []string{"http"}

	return r
}
