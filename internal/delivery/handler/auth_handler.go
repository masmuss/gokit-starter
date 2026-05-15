package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/masmuss/gokit-starter/internal/delivery/response"
	authapp "github.com/masmuss/gokit-starter/internal/modules/auth/app"
	authdomain "github.com/masmuss/gokit-starter/internal/modules/auth/domain"
	"github.com/masmuss/gokit-starter/internal/platform/validation"
)

// LoginRequest defines the payload required to authenticate a user.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginResponse defines the response returned after authentication.
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// AuthHandler serves auth endpoints.
type AuthHandler struct {
	log      *slog.Logger
	validate *validator.Validate
	service  *authapp.Service
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(
	service *authapp.Service,
	log *slog.Logger,
	validate *validator.Validate,
) *AuthHandler {
	return &AuthHandler{
		log:      log,
		validate: validate,
		service:  service,
	}
}

// Login handles user authentication.
// @Summary Login
// @Description Authenticates a user and returns an access token.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body LoginRequest true "Login payload"
// @Success 200 {object} response.Envelope
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest

	if err := validation.BindJSON(r, &request); err != nil {
		_ = response.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_json",
			"request body is invalid",
			nil,
		)
		return
	}

	if err := validation.ValidateStruct(h.validate, request); err != nil {
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

		_ = response.WriteError(
			w,
			http.StatusBadRequest,
			"validation_failed",
			"request validation failed",
			nil,
		)
		return
	}

	session, err := h.service.Login(r.Context(), authdomain.Credentials{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		if errors.Is(err, authdomain.ErrInvalidCredentials) {
			_ = response.WriteError(
				w,
				http.StatusUnauthorized,
				"invalid_credentials",
				"email or password is invalid",
				nil,
			)
			return
		}

		_ = response.WriteError(
			w,
			http.StatusInternalServerError,
			"authentication_failed",
			"failed to authenticate",
			nil,
		)
		return
	}

	_ = response.WriteJSON(w, http.StatusOK, response.OK(LoginResponse{
		AccessToken: session.AccessToken,
		TokenType:   session.TokenType,
		ExpiresIn:   session.ExpiresIn,
	}, "authenticated"))
}
