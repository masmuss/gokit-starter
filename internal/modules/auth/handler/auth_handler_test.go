package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/masmuss/gokit-starter/internal/infra/authtoken"
	"github.com/masmuss/gokit-starter/internal/modules/auth/domain"
	"github.com/masmuss/gokit-starter/internal/pkg/audit"
	"github.com/masmuss/gokit-starter/internal/pkg/validate"
)

type mockAuthService struct {
	registerFn       func(ctx context.Context, input domain.RegisterInput) (domain.Session, domain.Profile, error)
	loginFn          func(ctx context.Context, credentials domain.Credentials) (domain.Session, domain.Profile, error)
	profileFn        func(ctx context.Context, userID, orgID uuid.UUID) (domain.Profile, error)
	logoutFn         func(ctx context.Context, accessClaims, refreshClaims authtoken.Claims) error
	changePasswordFn func(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
	refreshTokenFn   func(ctx context.Context, token string) (domain.Session, error)
}

func (m *mockAuthService) Register(
	ctx context.Context,
	input domain.RegisterInput,
) (domain.Session, domain.Profile, error) {
	return m.registerFn(ctx, input)
}

func (m *mockAuthService) Login(
	ctx context.Context,
	credentials domain.Credentials,
) (domain.Session, domain.Profile, error) {
	return m.loginFn(ctx, credentials)
}

func (m *mockAuthService) Profile(ctx context.Context, userID, orgID uuid.UUID) (domain.Profile, error) {
	return m.profileFn(ctx, userID, orgID)
}

func (m *mockAuthService) Logout(ctx context.Context, accessClaims, refreshClaims authtoken.Claims) error {
	return m.logoutFn(ctx, accessClaims, refreshClaims)
}

func (m *mockAuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	return m.changePasswordFn(ctx, userID, oldPassword, newPassword)
}

func (m *mockAuthService) RefreshAccessToken(ctx context.Context, token string) (domain.Session, error) {
	return m.refreshTokenFn(ctx, token)
}

type mockTokenVerifier struct {
	verifyFn func(token string) (authtoken.Claims, error)
}

func (m *mockTokenVerifier) Verify(token string) (authtoken.Claims, error) {
	return m.verifyFn(token)
}

func TestAuthHandler_Register_Login_Profile(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()

	sampleProfile := domain.Profile{
		ID:     userID,
		Name:   "Alice",
		Email:  "alice@example.com",
		Status: "active",
		Organization: domain.Organization{
			ID:     orgID,
			Name:   "Org",
			Code:   "org-123",
			Type:   "company",
			Status: "active",
		},
	}

	newHandler := func(svc *mockAuthService, verifier authtoken.TokenVerifier) *AuthHandler {
		if verifier == nil {
			verifier = &mockTokenVerifier{}
		}
		return NewAuthHandler(
			svc,
			slog.New(slog.DiscardHandler),
			audit.New(slog.New(slog.DiscardHandler)),
			validate.New(),
			verifier,
		)
	}

	t.Run("Register handler", func(t *testing.T) {
		svc := &mockAuthService{
			registerFn: func(_ context.Context, _ domain.RegisterInput) (domain.Session, domain.Profile, error) {
				return domain.Session{
					AccessToken:  "tok-123",
					RefreshToken: "ref-123",
					TokenType:    "Bearer",
					ExpiresIn:    3600,
				}, sampleProfile, nil
			},
		}
		h := newHandler(svc, nil)

		reqBody := `{"name":"Alice","email":"alice@example.com","password":"secret123","organization_name":"Org"}`
		req := httptest.NewRequest("POST", "/auth/register", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.Register(rr, req)
		require.Equalf(t, 201, rr.Code, "body=%s", rr.Body.String())

		var env struct {
			Message string         `json:"message"`
			Data    map[string]any `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &env))
		require.Equal(t, "registered", env.Message)
		require.Equal(t, "tok-123", env.Data["access_token"])
		require.Equal(t, "ref-123", env.Data["refresh_token"])
		_, hasProfile := env.Data["user"]
		require.True(t, hasProfile)
	})

	t.Run("Login handler", func(t *testing.T) {
		svc := &mockAuthService{
			loginFn: func(_ context.Context, _ domain.Credentials) (domain.Session, domain.Profile, error) {
				return domain.Session{
					AccessToken:  "tok-login",
					RefreshToken: "ref-login",
					TokenType:    "Bearer",
					ExpiresIn:    3600,
				}, sampleProfile, nil
			},
		}
		h := newHandler(svc, nil)

		reqBody := `{"email":"alice@example.com","password":"secret123"}`
		req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.Login(rr, req)
		require.Equalf(t, 200, rr.Code, "body=%s", rr.Body.String())

		var env struct {
			Message string         `json:"message"`
			Data    map[string]any `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &env))
		require.Equal(t, "authenticated", env.Message)
		require.Equal(t, "tok-login", env.Data["access_token"])
		require.Equal(t, "ref-login", env.Data["refresh_token"])
		_, hasProfile := env.Data["user"]
		require.True(t, hasProfile, "login response should include user profile")
	})

	t.Run("Login inactive user", func(t *testing.T) {
		svc := &mockAuthService{
			loginFn: func(_ context.Context, _ domain.Credentials) (domain.Session, domain.Profile, error) {
				return domain.Session{}, domain.Profile{}, domain.ErrAccountInactive
			},
		}
		h := newHandler(svc, nil)

		reqBody := `{"email":"banned@example.com","password":"secret123"}`
		req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.Login(rr, req)
		require.Equal(t, 401, rr.Code)
	})

	t.Run("Register email already used", func(t *testing.T) {
		svc := &mockAuthService{
			registerFn: func(_ context.Context, _ domain.RegisterInput) (domain.Session, domain.Profile, error) {
				return domain.Session{}, domain.Profile{}, domain.ErrEmailAlreadyUsed
			},
		}
		h := newHandler(svc, nil)

		reqBody := `{"name":"Alice","email":"alice@example.com","password":"secret123"}`
		req := httptest.NewRequest("POST", "/auth/register", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.Register(rr, req)
		require.Equal(t, 409, rr.Code)

		var env struct {
			Error string `json:"error"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &env))
		require.Equal(t, "registration_failed", env.Error)
	})

	t.Run("Profile handler", func(t *testing.T) {
		svc := &mockAuthService{
			profileFn: func(_ context.Context, _, _ uuid.UUID) (domain.Profile, error) {
				return sampleProfile, nil
			},
		}
		h := newHandler(svc, nil)

		req := httptest.NewRequest("GET", "/auth/profile", nil)
		authClaims := authtoken.Claims{UserID: userID, OrganizationID: orgID, Email: sampleProfile.Email}
		req = req.WithContext(authtoken.WithClaims(req.Context(), authClaims))
		rr := httptest.NewRecorder()

		h.Profile(rr, req)
		require.Equal(t, 200, rr.Code)

		var env struct {
			Message string         `json:"message"`
			Data    map[string]any `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &env))
		require.NotNil(t, env.Data, "body=%s", rr.Body.String())
		require.Equal(t, "alice@example.com", env.Data["email"], "body=%s", rr.Body.String())
	})
}
