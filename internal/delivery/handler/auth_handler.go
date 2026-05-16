// Package handler provides HTTP handlers for auth operations.
package handler

import (
	"log/slog"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/masmuss/gokit-starter/internal/delivery/middleware"
	"github.com/masmuss/gokit-starter/internal/delivery/response"
	auth_app "github.com/masmuss/gokit-starter/internal/modules/auth/app"
	"github.com/masmuss/gokit-starter/internal/modules/auth/domain"
	"github.com/masmuss/gokit-starter/internal/platform/auth"
	"github.com/masmuss/gokit-starter/internal/platform/validation"
	"github.com/masmuss/gokit-starter/internal/shared/errors"
)

// AuthHandler handles authentication requests.
type AuthHandler struct {
	service   *auth_app.Service
	log       *slog.Logger
	validator *validator.Validate
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	service *auth_app.Service,
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
	Name             string `json:"name"              validate:"required"`
	Email            string `json:"email"             validate:"required,email"`
	Password         string `json:"password"          validate:"required,min=8"`
	OrganizationName string `json:"organization_name" validate:"omitempty,min=3"`
}

// Register handles user registration.
//
//	@Summary		Register a new user
//	@Description	Register a new user and create an organization
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RegisterRequest	true	"Registration details"
//	@Success		201		{object}	response.Envelope
//	@Failure		400		{object}	response.ErrorEnvelope
//	@Router			/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := validation.BindJSON(r, &req); err != nil {
		_ = response.WriteAppError(w, err)
		return
	}

	if err := validation.ValidateStruct(h.validator, req); err != nil {
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
//
//	@Summary		User login
//	@Description	Authenticate user and return JWT token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LoginRequest	true	"Login credentials"
//	@Success		200		{object}	response.Envelope
//	@Failure		401		{object}	response.ErrorEnvelope
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := validation.BindJSON(r, &req); err != nil {
		_ = response.WriteAppError(w, err)
		return
	}

	if err := validation.ValidateStruct(h.validator, req); err != nil {
		_ = response.WriteAppError(w, err)
		return
	}

	session, _, err := h.service.Login(r.Context(), domain.Credentials{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		_ = response.WriteAppError(w, err)
		return
	}

	_ = response.WriteJSON(w, http.StatusOK, response.Envelope{
		Message: "authenticated",
		Data: map[string]any{
			"access_token": session.AccessToken,
		},
	})
}

// Profile handles retrieving the current user's profile.
//
//	@Summary		Get current user profile
//	@Description	Retrieve profile details for the authenticated user
//	@Tags			auth
//	@Produce		json
//	@Security		Bearer
//	@Success		200	{object}	response.Envelope
//	@Failure		401	{object}	response.ErrorEnvelope
//	@Router			/auth/profile [get]
func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		_ = response.WriteAppError(w, errors.Unauthorized("unauthorized", "unauthorized access"))
		return
	}

	profile, err := h.service.Profile(r.Context(), claims.UserID)
	if err != nil {
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
