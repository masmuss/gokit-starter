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

	"github.com/masmuss/gokit-starter/internal/infra/auth"
	"github.com/masmuss/gokit-starter/internal/modules/auth/domain"
	"github.com/masmuss/gokit-starter/internal/pkg/validate"
)

type mockAuthService struct {
	registerFn func(ctx context.Context, input domain.RegisterInput) (domain.Session, domain.Profile, error)
	loginFn    func(ctx context.Context, credentials domain.Credentials) (domain.Session, domain.Profile, error)
	profileFn  func(ctx context.Context, userID, orgID uuid.UUID) (domain.Profile, error)
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

	t.Run("Register handler", func(t *testing.T) {
		svc := &mockAuthService{
			registerFn: func(_ context.Context, _ domain.RegisterInput) (domain.Session, domain.Profile, error) {
				return domain.Session{
					AccessToken: "tok-123",
					TokenType:   "Bearer",
					ExpiresIn:   3600,
				}, sampleProfile, nil
			},
		}
		h := NewAuthHandler(svc, slog.New(slog.DiscardHandler), validate.New())

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
		userMap, ok := env.Data["user"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, "alice@example.com", userMap["email"])
	})

	t.Run("Login handler", func(t *testing.T) {
		svc := &mockAuthService{
			loginFn: func(_ context.Context, _ domain.Credentials) (domain.Session, domain.Profile, error) {
				return domain.Session{
					AccessToken: "tok-login",
					TokenType:   "Bearer",
					ExpiresIn:   3600,
				}, sampleProfile, nil
			},
		}
		h := NewAuthHandler(svc, slog.New(slog.DiscardHandler), validate.New())

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
	})

	t.Run("Profile handler", func(t *testing.T) {
		svc := &mockAuthService{
			profileFn: func(_ context.Context, _, _ uuid.UUID) (domain.Profile, error) {
				return sampleProfile, nil
			},
		}
		h := NewAuthHandler(svc, slog.New(slog.DiscardHandler), validate.New())

		req := httptest.NewRequest("GET", "/auth/profile", nil)
		claims := auth.Claims{UserID: userID, OrganizationID: orgID, Email: sampleProfile.Email}
		req = req.WithContext(auth.WithClaims(req.Context(), claims))
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
