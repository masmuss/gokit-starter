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

	authapp "github.com/masmuss/gokit-starter/internal/modules/auth/app"
	"github.com/masmuss/gokit-starter/internal/modules/auth/domain"
	"github.com/masmuss/gokit-starter/internal/platform/auth"
	"github.com/masmuss/gokit-starter/internal/platform/validation"
)

// reuse simple mocks similar to service tests
type mockRepo struct {
	createAccount func(_ context.Context, _ domain.RegisterInput, _ string) (domain.User, error)
	findByEmail   func(_ context.Context, _ string) (domain.User, error)
	findByID      func(_ context.Context, _ uuid.UUID) (domain.User, error)
}

func (m *mockRepo) CreateAccount(
	ctx context.Context,
	input domain.RegisterInput,
	passwordHash string,
) (domain.User, error) {
	return m.createAccount(ctx, input, passwordHash)
}

func (m *mockRepo) FindByEmail(
	ctx context.Context,
	email string,
) (domain.User, error) {
	return m.findByEmail(ctx, email)
}

func (m *mockRepo) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.User, error) {
	return m.findByID(ctx, id)
}

type mockHasher struct {
	hashFunc    func(_ string) (string, error)
	compareFunc func(_ string, _ string) error
}

func (m *mockHasher) Hash(password string) (string, error) { return m.hashFunc(password) }
func (m *mockHasher) Compare(hash, password string) error  { return m.compareFunc(hash, password) }

type mockToken struct {
	issueFunc func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string) (string, error)
}

func (m *mockToken) Issue(ctx context.Context, id uuid.UUID, orgID uuid.UUID, email string) (string, error) {
	return m.issueFunc(ctx, id, orgID, email)
}

func TestAuthHandler_Register_Login_Profile(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()

	sampleUser := domain.User{
		ID:           userID,
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed",
		Status:       "active",
		Organization: domain.Organization{
			ID:     orgID,
			Name:   "Org",
			Code:   "org-123",
			Type:   "company",
			Status: "active",
		},
	}

	// Register
	t.Run("Register handler", func(t *testing.T) {
		repo := &mockRepo{
			createAccount: func(_ context.Context, _ domain.RegisterInput, _ string) (domain.User, error) {
				return sampleUser, nil
			},
		}
		hasher := &mockHasher{hashFunc: func(_ string) (string, error) { return "hashed", nil }}
		token := &mockToken{
			issueFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string) (string, error) {
				return "tok-123", nil
			},
		}
		svc := authapp.New(repo, hasher, token, 3600)
		h := NewAuthHandler(
			svc,
			slog.New(slog.DiscardHandler),
			validation.New(),
		)

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
		reqBodyBytes := rr.Body.Bytes()
		require.NoError(t, json.Unmarshal(reqBodyBytes, &env))
		require.Equal(t, "registered", env.Message)
		data := env.Data
		require.Equal(t, "tok-123", data["access_token"])
		userMap, ok := data["user"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, sampleUser.Email, userMap["email"])
	})

	// Login

	t.Run("Login handler", func(t *testing.T) {
		repo := &mockRepo{findByEmail: func(_ context.Context, _ string) (domain.User, error) {
			return sampleUser, nil
		}}
		hasher := &mockHasher{compareFunc: func(_ string, _ string) error { return nil }}
		token := &mockToken{
			issueFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string) (string, error) {
				return "tok-login", nil
			},
		}
		svc := authapp.New(repo, hasher, token, 3600)
		h := NewAuthHandler(
			svc,
			slog.New(slog.DiscardHandler),
			validation.New(),
		)

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

	// Profile

	t.Run("Profile handler", func(t *testing.T) {
		repo := &mockRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (domain.User, error) { return sampleUser, nil },
		}
		hasher := &mockHasher{}
		token := &mockToken{}
		svc := authapp.New(repo, hasher, token, 3600)
		h := NewAuthHandler(
			svc,
			slog.New(slog.DiscardHandler),
			validation.New(),
		)

		req := httptest.NewRequest("GET", "/auth/profile", nil)
		// inject claims into context as middleware would
		claims := auth.Claims{UserID: userID, OrganizationID: orgID, Email: sampleUser.Email}
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
		// profile handler returns the profile object directly in data
		require.Equal(t, sampleUser.Email, env.Data["email"], "body=%s", rr.Body.String())
	})
}
