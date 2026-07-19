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
		Role:         "admin",
		Organization: domain.Organization{
			ID:     orgID,
			Name:   "Org",
			Code:   "org-123",
			Type:   "company",
			Status: "active",
		},
	}

	newMocks := func() (
		*mocks.RepositoryMock,
		*mocks.PasswordHasherMock,
		*mocks.TokenIssuerMock,
		*mocks.RefreshTokenIssuerMock,
		*mocks.TokenVerifierMock,
	) {
		return mocks.NewRepositoryMock(t),
			mocks.NewPasswordHasherMock(t),
			mocks.NewTokenIssuerMock(t),
			mocks.NewRefreshTokenIssuerMock(t),
			mocks.NewTokenVerifierMock(t)
	}

	newService := func(
		repo *mocks.RepositoryMock,
		hasher *mocks.PasswordHasherMock,
		tokens *mocks.TokenIssuerMock,
		refreshTokens *mocks.RefreshTokenIssuerMock,
		verifier *mocks.TokenVerifierMock,
	) *Service {
		return New(repo, hasher, tokens, refreshTokens, verifier, nil, 3600, 604800)
	}

	// Register success
	t.Run("Register success", func(t *testing.T) {
		repo, hasher, tokens, refreshTokens, verifier := newMocks()

		hasher.EXPECT().Hash("secret").Return("hashed", nil)

		repo.EXPECT().CreateAccount(mock.Anything, mock.MatchedBy(func(input domain.RegisterInput) bool {
			return input.Name == "Alice" && input.Email == "alice@example.com"
		}), "hashed").Return(sampleUser, nil)

		tokens.EXPECT().
			Issue(mock.Anything, sampleUser.ID, sampleUser.Organization.ID, sampleUser.Email, sampleUser.Role).
			Return("tok-123", nil)

		refreshTokens.EXPECT().
			IssueRefresh(mock.Anything, sampleUser.ID, sampleUser.Organization.ID, sampleUser.Email, sampleUser.Role).
			Return("ref-123", nil)

		svc := newService(repo, hasher, tokens, refreshTokens, verifier)

		sess, prof, err := svc.Register(ctx, domain.RegisterInput{
			Name:             "Alice",
			Email:            "alice@example.com",
			Password:         "secret",
			OrganizationName: "Org",
		})
		require.NoError(t, err)
		require.Equal(t, "tok-123", sess.AccessToken)
		require.Equal(t, "ref-123", sess.RefreshToken)
		require.Equal(t, sampleUser.Email, prof.Email)
	})

	// Login success
	t.Run("Login success", func(t *testing.T) {
		repo, hasher, tokens, refreshTokens, verifier := newMocks()

		repo.EXPECT().FindByEmail(mock.Anything, "alice@example.com").Return(sampleUser, nil)
		hasher.EXPECT().Compare("hashed", "secret").Return(nil)
		tokens.EXPECT().
			Issue(mock.Anything, sampleUser.ID, sampleUser.Organization.ID, sampleUser.Email, sampleUser.Role).
			Return("tok-login", nil)

		refreshTokens.EXPECT().
			IssueRefresh(mock.Anything, sampleUser.ID, sampleUser.Organization.ID, sampleUser.Email, sampleUser.Role).
			Return("ref-login", nil)

		svc := newService(repo, hasher, tokens, refreshTokens, verifier)

		sess, prof, err := svc.Login(ctx, domain.Credentials{
			Email:    sampleUser.Email,
			Password: "secret",
		})
		require.NoError(t, err)
		require.Equal(t, "tok-login", sess.AccessToken)
		require.Equal(t, "ref-login", sess.RefreshToken)
		require.Equal(t, sampleUser.Email, prof.Email)
	})

	// Login invalid credentials
	t.Run("Login invalid credentials", func(t *testing.T) {
		repo, hasher, tokens, refreshTokens, verifier := newMocks()

		repo.EXPECT().FindByEmail(mock.Anything, sampleUser.Email).Return(sampleUser, nil)
		hasher.EXPECT().Compare("hashed", "badpass").Return(errors.New("mismatch"))

		svc := newService(repo, hasher, tokens, refreshTokens, verifier)

		_, _, err := svc.Login(ctx, domain.Credentials{
			Email:    sampleUser.Email,
			Password: "badpass",
		})
		require.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	// Profile user not found
	t.Run("Profile not found", func(t *testing.T) {
		repo, hasher, tokens, refreshTokens, verifier := newMocks()

		repo.EXPECT().FindByID(mock.Anything, mock.Anything).Return(domain.User{}, domain.ErrUserNotFound)

		svc := newService(repo, hasher, tokens, refreshTokens, verifier)

		_, err := svc.Profile(ctx, uuid.New(), uuid.New())
		require.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	// Profile wrong org
	t.Run("Profile wrong org", func(t *testing.T) {
		repo, hasher, tokens, refreshTokens, verifier := newMocks()

		repo.EXPECT().FindByID(mock.Anything, sampleUser.ID).Return(sampleUser, nil)

		svc := newService(repo, hasher, tokens, refreshTokens, verifier)

		_, err := svc.Profile(ctx, sampleUser.ID, uuid.New())
		require.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	// Change password success
	t.Run("ChangePassword success", func(t *testing.T) {
		repo, hasher, tokens, refreshTokens, verifier := newMocks()

		repo.EXPECT().FindByID(mock.Anything, sampleUser.ID).Return(sampleUser, nil)
		hasher.EXPECT().Compare("hashed", "oldpass").Return(nil)
		hasher.EXPECT().Hash("newpass").Return("newhash", nil)
		repo.EXPECT().UpdatePassword(mock.Anything, sampleUser.ID, "newhash").Return(nil)

		svc := newService(repo, hasher, tokens, refreshTokens, verifier)

		err := svc.ChangePassword(ctx, sampleUser.ID, "oldpass", "newpass")
		require.NoError(t, err)
	})

	// Change password wrong old password
	t.Run("ChangePassword wrong old", func(t *testing.T) {
		repo, hasher, tokens, refreshTokens, verifier := newMocks()

		repo.EXPECT().FindByID(mock.Anything, sampleUser.ID).Return(sampleUser, nil)
		hasher.EXPECT().Compare("hashed", "wrong").Return(errors.New("mismatch"))

		svc := newService(repo, hasher, tokens, refreshTokens, verifier)

		err := svc.ChangePassword(ctx, sampleUser.ID, "wrong", "newpass")
		require.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})
}
