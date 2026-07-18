package app

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/masmuss/gokit-starter/internal/modules/auth/domain"
	"github.com/masmuss/gokit-starter/test/mocks"
)

func TestService_Register_Login_Profile(t *testing.T) {
	ctx := context.Background()

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

	// Register success
	t.Run("Register success", func(t *testing.T) {
		repo := mocks.NewRepositoryMock(t)
		hasher := mocks.NewPasswordHasherMock(t)
		token := mocks.NewTokenIssuerMock(t)

		hasher.EXPECT().Hash("secret").Return("hashed", nil)

		repo.EXPECT().CreateAccount(mock.Anything, mock.MatchedBy(func(input domain.RegisterInput) bool {
			return input.Name == "Alice" && input.Email == "alice@example.com"
		}), "hashed").Return(sampleUser, nil)

		token.EXPECT().
			Issue(mock.Anything, sampleUser.ID, sampleUser.Organization.ID, sampleUser.Email).
			Return("tok-123", nil)

		svc := New(repo, hasher, token, 3600)

		sess, prof, err := svc.Register(ctx, domain.RegisterInput{
			Name:             "Alice",
			Email:            "alice@example.com",
			Password:         "secret",
			OrganizationName: "Org",
		})
		require.NoError(t, err)
		require.Equal(t, "tok-123", sess.AccessToken)
		require.Equal(t, sampleUser.Email, prof.Email)
	})

	// Login success
	t.Run("Login success", func(t *testing.T) {
		repo := mocks.NewRepositoryMock(t)
		hasher := mocks.NewPasswordHasherMock(t)
		token := mocks.NewTokenIssuerMock(t)

		repo.EXPECT().FindByEmail(mock.Anything, "alice@example.com").Return(sampleUser, nil)
		hasher.EXPECT().Compare("hashed", "secret").Return(nil)
		token.EXPECT().
			Issue(mock.Anything, sampleUser.ID, sampleUser.Organization.ID, sampleUser.Email).
			Return("tok-login", nil)

		svc := New(repo, hasher, token, 3600)

		sess, prof, err := svc.Login(ctx, domain.Credentials{
			Email:    sampleUser.Email,
			Password: "secret",
		})
		require.NoError(t, err)
		require.Equal(t, "tok-login", sess.AccessToken)
		require.Equal(t, sampleUser.Email, prof.Email)
	})

	// Login invalid credentials
	t.Run("Login invalid credentials", func(t *testing.T) {
		repo := mocks.NewRepositoryMock(t)
		hasher := mocks.NewPasswordHasherMock(t)
		token := mocks.NewTokenIssuerMock(t)

		repo.EXPECT().FindByEmail(mock.Anything, sampleUser.Email).Return(sampleUser, nil)
		hasher.EXPECT().Compare("hashed", "badpass").Return(errors.New("mismatch"))

		svc := New(repo, hasher, token, 3600)

		_, _, err := svc.Login(ctx, domain.Credentials{
			Email:    sampleUser.Email,
			Password: "badpass",
		})
		require.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	// Profile user not found
	t.Run("Profile not found", func(t *testing.T) {
		repo := mocks.NewRepositoryMock(t)
		hasher := mocks.NewPasswordHasherMock(t)
		token := mocks.NewTokenIssuerMock(t)

		repo.EXPECT().FindByID(mock.Anything, mock.Anything).Return(domain.User{}, domain.ErrUserNotFound)

		svc := New(repo, hasher, token, 3600)

		_, err := svc.Profile(ctx, uuid.New(), uuid.New())
		require.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	// Profile wrong org
	t.Run("Profile wrong org", func(t *testing.T) {
		repo := mocks.NewRepositoryMock(t)
		hasher := mocks.NewPasswordHasherMock(t)
		token := mocks.NewTokenIssuerMock(t)

		repo.EXPECT().FindByID(mock.Anything, sampleUser.ID).Return(sampleUser, nil)

		svc := New(repo, hasher, token, 3600)

		_, err := svc.Profile(ctx, sampleUser.ID, uuid.New())
		require.ErrorIs(t, err, domain.ErrUserNotFound)
	})
}
