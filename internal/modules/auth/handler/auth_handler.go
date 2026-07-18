// Package handler provides HTTP handlers for the auth module.
package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/masmuss/gokit-starter/internal/delivery/middleware"
	"github.com/masmuss/gokit-starter/internal/delivery/response"
	"github.com/masmuss/gokit-starter/internal/infra/auth"
	"github.com/masmuss/gokit-starter/internal/modules/auth/domain"
	"github.com/masmuss/gokit-starter/internal/pkg/apperr"
	"github.com/masmuss/gokit-starter/internal/pkg/validate"
)

// AuthService defines the interface for authentication operations.
type AuthService interface {
	Register(ctx context.Context, input domain.RegisterInput) (domain.Session, domain.Profile, error)
	Login(ctx context.Context, credentials domain.Credentials) (domain.Session, domain.Profile, error)
	Profile(ctx context.Context, userID, orgID uuid.UUID) (domain.Profile, error)
}

// AuthHandler handles authentication requests.
type AuthHandler struct {
	service   AuthService
	log       *slog.Logger
	validator *validator.Validate
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	service AuthService,
	log *slog.Logger,
	v *validator.Validate,
) *AuthHandler {
	return &AuthHandler{
		service:   service,
		log:       log,
		validator: v,
	}
}

// RegisterRequest defines the input for user registration.
type RegisterRequest struct {
	Name             string `json:"name"              validate:"required,min=2,max=128"`
	Email            string `json:"email"             validate:"required,email"`
	Password         string `json:"password"          validate:"required,min=8"`
	OrganizationName string `json:"organization_name" validate:"omitempty,min=3"`
}

// Register handles user registration.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := validate.BindJSON(r, &req); err != nil {
		_ = response.WriteAppError(w, err)
		return
	}

	if err := validate.Struct(h.validator, req); err != nil {
		_ = response.WriteAppError(w, err)
		return
	}

	session, profile, err := h.service.Register(r.Context(), domain.RegisterInput{
		Name:             req.Name,
		Email:            req.Email,
		Password:         req.Password,
		OrganizationName: req.OrganizationName,
	})
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyUsed) {
			err = apperr.Conflict("registration_failed", "registration failed")
		}
		_ = response.WriteAppError(w, err)
		return
	}

	_ = response.WriteJSON(w, http.StatusCreated, response.Envelope{
		Message: "registered",
		Data: map[string]any{
			"user":         profile,
			"access_token": session.AccessToken,
		},
	})
}

// LoginRequest defines the input for user login.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Login handles user authentication.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := validate.BindJSON(r, &req); err != nil {
		_ = response.WriteAppError(w, err)
		return
	}

	if err := validate.Struct(h.validator, req); err != nil {
		_ = response.WriteAppError(w, err)
		return
	}

	session, profile, err := h.service.Login(r.Context(), domain.Credentials{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) ||
			errors.Is(err, domain.ErrInvalidCredentials) ||
			errors.Is(err, domain.ErrAccountInactive) {
			err = apperr.Unauthorized("invalid_credentials", "invalid email or password")
		}
		_ = response.WriteAppError(w, err)
		return
	}

	_ = response.WriteJSON(w, http.StatusOK, response.Envelope{
		Message: "authenticated",
		Data: map[string]any{
			"user":         profile,
			"access_token": session.AccessToken,
		},
	})
}

// Profile handles retrieving the current user's profile.
func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		_ = response.WriteAppError(w, apperr.Unauthorized("unauthorized", "unauthorized access"))
		return
	}

	profile, err := h.service.Profile(r.Context(), claims.UserID, claims.OrganizationID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			err = apperr.NotFound("user_not_found", err.Error())
		}
		_ = response.WriteAppError(w, err)
		return
	}

	_ = response.WriteJSON(w, http.StatusOK, response.Envelope{
		Message: "profile",
		Data:    profile,
	})
}

// RegisterRoutes registers auth routes on the given router.
func (h *AuthHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.With(authMiddleware.Require).Get("/profile", h.Profile)
}
