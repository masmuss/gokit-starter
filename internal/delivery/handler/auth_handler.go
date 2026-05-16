package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/masmuss/gokit-starter/internal/delivery/response"
	authapp "github.com/masmuss/gokit-starter/internal/modules/auth/app"
	authdomain "github.com/masmuss/gokit-starter/internal/modules/auth/domain"
	platformauth "github.com/masmuss/gokit-starter/internal/platform/auth"
	"github.com/masmuss/gokit-starter/internal/platform/validation"
)

// RegisterRequest defines the payload required to register a user.
type RegisterRequest struct {
	Name             string `json:"name"              validate:"required"`
	Email            string `json:"email"             validate:"required,email"`
	Password         string `json:"password"          validate:"required,min=8"`
	OrganizationName string `json:"organization_name" validate:"omitempty"`
}

// LoginRequest defines the payload required to authenticate a user.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// OrganizationResponse represents the organization data in transport layer.
type OrganizationResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

// ProfileResponse represents the public user profile.
type ProfileResponse struct {
	ID           string               `json:"id"`
	Name         string               `json:"name"`
	Email        string               `json:"email"`
	Status       string               `json:"status"`
	Organization OrganizationResponse `json:"organization"`
}

// AuthResponse represents authentication success payloads.
type AuthResponse struct {
	AccessToken string          `json:"access_token"`
	TokenType   string          `json:"token_type"`
	ExpiresIn   int             `json:"expires_in"`
	User        ProfileResponse `json:"user"`
}

// AuthHandler serves auth endpoints.
type AuthHandler struct {
	log      *slog.Logger
	validate *validator.Validate
	service  *authapp.Service
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(service *authapp.Service, log *slog.Logger, validate *validator.Validate) *AuthHandler {
	return &AuthHandler{
		log:      log,
		validate: validate,
		service:  service,
	}
}

// Register handles account creation.
// @Summary Register
// @Description Creates a new account and organization.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body RegisterRequest true "Register payload"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 409 {object} response.ErrorEnvelope
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest

	if err := validation.BindJSON(r, &request); err != nil {
		_ = response.WriteError(w, http.StatusBadRequest, "invalid_json", "request body is invalid", nil)
		return
	}

	if err := validation.ValidateStruct(h.validate, request); err != nil {
		h.writeValidationError(w, err)
		return
	}

	session, profile, err := h.service.Register(r.Context(), authdomain.RegisterInput{
		Name:             request.Name,
		Email:            request.Email,
		Password:         request.Password,
		OrganizationName: request.OrganizationName,
	})
	if err != nil {
		h.writeAuthError(w, err, true)
		return
	}

	_ = response.WriteJSON(w, http.StatusCreated, response.OK(AuthResponse{
		AccessToken: session.AccessToken,
		TokenType:   session.TokenType,
		ExpiresIn:   session.ExpiresIn,
		User:        toProfileResponse(profile),
	}, "registered"))
}

// Login handles user authentication.
// @Summary Login
// @Description Authenticates a user and returns an access token.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body LoginRequest true "Login payload"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest

	if err := validation.BindJSON(r, &request); err != nil {
		_ = response.WriteError(w, http.StatusBadRequest, "invalid_json", "request body is invalid", nil)
		return
	}

	if err := validation.ValidateStruct(h.validate, request); err != nil {
		h.writeValidationError(w, err)
		return
	}

	session, profile, err := h.service.Login(r.Context(), authdomain.Credentials{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		h.writeAuthError(w, err, false)
		return
	}

	_ = response.WriteJSON(w, http.StatusOK, response.OK(AuthResponse{
		AccessToken: session.AccessToken,
		TokenType:   session.TokenType,
		ExpiresIn:   session.ExpiresIn,
		User:        toProfileResponse(profile),
	}, "authenticated"))
}

// Profile returns the current authenticated user profile.
// @Summary Profile
// @Description Returns the current authenticated user profile.
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ProfileResponse
// @Failure 401 {object} response.ErrorEnvelope
// @Router /auth/profile [get]
func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	userID, ok := platformauth.CurrentUserID(r.Context())
	if !ok {
		_ = response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing authentication context", nil)
		return
	}

	profile, err := h.service.Profile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, authdomain.ErrUserNotFound) {
			_ = response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found", nil)
			return
		}

		_ = response.WriteError(w, http.StatusInternalServerError, "profile_failed", "failed to load profile", nil)
		return
	}

	_ = response.WriteJSON(w, http.StatusOK, response.OK(toProfileResponse(profile), "profile loaded"))
}

func (h *AuthHandler) writeValidationError(w http.ResponseWriter, err error) {
	var validationError validation.Error
	if errors.As(err, &validationError) {
		_ = response.WriteError(
			w,
			http.StatusBadRequest,
			"validation_failed",
			validationError.Message,
			validationError.Fields,
		)
		return
	}

	_ = response.WriteError(w, http.StatusBadRequest, "validation_failed", "request validation failed", nil)
}

func (h *AuthHandler) writeAuthError(w http.ResponseWriter, err error, isRegister bool) {
	switch {
	case errors.Is(err, authdomain.ErrEmailAlreadyUsed):
		_ = response.WriteError(
			w,
			http.StatusConflict,
			"email_already_used",
			"email is already registered",
			nil,
		)
	case errors.Is(err, authdomain.ErrInvalidCredentials):
		_ = response.WriteError(
			w,
			http.StatusUnauthorized,
			"invalid_credentials",
			"email or password is invalid",
			nil,
		)
	default:
		code := "authentication_failed"
		message := "failed to authenticate"
		status := http.StatusInternalServerError
		if isRegister {
			code = "registration_failed"
			message = "failed to register account"
		}

		_ = response.WriteError(w, status, code, message, nil)
	}
}

func toProfileResponse(profile authdomain.Profile) ProfileResponse {
	return ProfileResponse{
		ID:     profile.ID.String(),
		Name:   profile.Name,
		Email:  profile.Email,
		Status: profile.Status,
		Organization: OrganizationResponse{
			ID:     profile.Organization.ID.String(),
			Name:   profile.Organization.Name,
			Code:   profile.Organization.Code,
			Type:   profile.Organization.Type,
			Status: profile.Organization.Status,
		},
	}
}
