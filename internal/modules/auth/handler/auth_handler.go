// Package handler provides HTTP handlers for the auth module.
package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/masmuss/gokit-starter/internal/inbound/middleware"
	"github.com/masmuss/gokit-starter/internal/inbound/response"
	"github.com/masmuss/gokit-starter/internal/modules/auth/domain"
	"github.com/masmuss/gokit-starter/internal/outbound/authtoken"
	"github.com/masmuss/gokit-starter/internal/pkg/apperr"
	"github.com/masmuss/gokit-starter/internal/pkg/audit"
	"github.com/masmuss/gokit-starter/internal/pkg/validate"
)

// AuthService defines the interface for authentication operations.
type AuthService interface {
	Register(ctx context.Context, input domain.RegisterInput) (domain.Session, domain.Profile, error)
	Login(ctx context.Context, credentials domain.Credentials) (domain.Session, domain.Profile, error)
	Profile(ctx context.Context, userID, orgID uuid.UUID) (domain.Profile, error)
	Logout(ctx context.Context, accessClaims, refreshClaims authtoken.Claims) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
	RefreshAccessToken(ctx context.Context, token string) (domain.Session, error)
}

// AuthHandler handles authentication requests.
type AuthHandler struct {
	service       AuthService
	log           *slog.Logger
	audit         *audit.Logger
	validator     *validator.Validate
	tokenVerifier authtoken.TokenVerifier
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	service AuthService,
	log *slog.Logger,
	audit *audit.Logger,
	v *validator.Validate,
	tokenVerifier authtoken.TokenVerifier,
) *AuthHandler {
	return &AuthHandler{
		service:       service,
		log:           log,
		audit:         audit,
		validator:     v,
		tokenVerifier: tokenVerifier,
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
		h.audit.Warn("user.registered", "failed", audit.Email(req.Email), audit.IP(r))
		_ = response.WriteAppError(w, err)
		return
	}

	h.audit.Info("user.registered", "success",
		audit.UserID(profile.ID), audit.OrgID(profile.Organization.ID), audit.Email(req.Email), audit.IP(r))

	_ = response.WriteJSON(w, http.StatusCreated, response.Envelope{
		Message: "registered",
		Data:    sessionResponse(session, profile),
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
		isAuthFailure := errors.Is(err, domain.ErrUserNotFound) ||
			errors.Is(err, domain.ErrInvalidCredentials) ||
			errors.Is(err, domain.ErrAccountInactive)
		if isAuthFailure {
			err = apperr.Unauthorized("invalid_credentials", "invalid email or password")
		}
		h.audit.Warn("user.login", "failed", audit.Email(req.Email), audit.IP(r))
		_ = response.WriteAppError(w, err)
		return
	}

	h.audit.Info("user.login", "success",
		audit.UserID(profile.ID), audit.OrgID(profile.Organization.ID), audit.Email(req.Email), audit.IP(r))

	_ = response.WriteJSON(w, http.StatusOK, response.Envelope{
		Message: "authenticated",
		Data:    sessionResponse(session, profile),
	})
}

// Profile handles retrieving the current user's profile.
func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	claims, ok := authtoken.ClaimsFromContext(r.Context())
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

// LogoutRequest defines the input for user logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Logout handles user logout by revoking both tokens.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	accessClaims, ok := authtoken.ClaimsFromContext(r.Context())
	if !ok {
		_ = response.WriteAppError(w, apperr.Unauthorized("unauthorized", "unauthorized access"))
		return
	}

	var req LogoutRequest
	if err := validate.BindJSON(r, &req); err != nil {
		_ = response.WriteAppError(w, err)
		return
	}

	if err := validate.Struct(h.validator, req); err != nil {
		_ = response.WriteAppError(w, err)
		return
	}

	refreshClaims, err := h.tokenVerifier.Verify(req.RefreshToken)
	if err != nil {
		_ = response.WriteAppError(w, apperr.BadRequest("invalid_refresh_token", "invalid refresh token"))
		return
	}

	if err = h.service.Logout(r.Context(), accessClaims, refreshClaims); err != nil {
		_ = response.WriteAppError(w, err)
		return
	}

	h.audit.Info("user.logout", "success",
		audit.UserID(accessClaims.UserID), audit.OrgID(accessClaims.OrganizationID), audit.IP(r))

	_ = response.WriteJSON(w, http.StatusOK, response.Envelope{
		Message: "logged out",
	})
}

// ChangePasswordRequest defines the input for changing password.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ChangePassword handles password change for the authenticated user.
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := authtoken.ClaimsFromContext(r.Context())
	if !ok {
		_ = response.WriteAppError(w, apperr.Unauthorized("unauthorized", "unauthorized access"))
		return
	}

	var req ChangePasswordRequest
	if err := validate.BindJSON(r, &req); err != nil {
		_ = response.WriteAppError(w, err)
		return
	}

	if err := validate.Struct(h.validator, req); err != nil {
		_ = response.WriteAppError(w, err)
		return
	}

	if err := h.service.ChangePassword(r.Context(), claims.UserID, req.OldPassword, req.NewPassword); err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			err = apperr.BadRequest("invalid_password", "current password is incorrect")
		}
		h.audit.Warn("user.password_changed", "failed",
			audit.UserID(claims.UserID), audit.OrgID(claims.OrganizationID), audit.IP(r))
		_ = response.WriteAppError(w, err)
		return
	}

	h.audit.Info("user.password_changed", "success",
		audit.UserID(claims.UserID), audit.OrgID(claims.OrganizationID), audit.IP(r))

	_ = response.WriteJSON(w, http.StatusOK, response.Envelope{
		Message: "password changed",
	})
}

// RefreshTokenRequest defines the input for refreshing a token.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshToken handles access token refresh.
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := validate.BindJSON(r, &req); err != nil {
		_ = response.WriteAppError(w, err)
		return
	}

	if err := validate.Struct(h.validator, req); err != nil {
		_ = response.WriteAppError(w, err)
		return
	}

	session, err := h.service.RefreshAccessToken(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			err = apperr.Unauthorized("invalid_refresh_token", "invalid or expired refresh token")
		}
		h.audit.Warn("user.token_refreshed", "failed", audit.IP(r))
		_ = response.WriteAppError(w, err)
		return
	}

	h.audit.Info("user.token_refreshed", "success")

	_ = response.WriteJSON(w, http.StatusOK, response.Envelope{
		Message: "token refreshed",
		Data:    sessionResponse(session, domain.Profile{}),
	})
}

// RegisterRoutes registers auth routes on the given router.
func (h *AuthHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.RefreshToken)
	r.With(authMiddleware.Require).Get("/profile", h.Profile)
	r.With(authMiddleware.Require).Post("/logout", h.Logout)
	r.With(authMiddleware.Require).Put("/password", h.ChangePassword)
}

func sessionResponse(session domain.Session, profile domain.Profile) map[string]any {
	data := map[string]any{
		"access_token":  session.AccessToken,
		"refresh_token": session.RefreshToken,
		"token_type":    session.TokenType,
		"expires_in":    session.ExpiresIn,
	}
	if profile.ID != uuid.Nil {
		data["user"] = profile
	}
	return data
}
