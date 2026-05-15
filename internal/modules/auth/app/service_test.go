package app

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/masmuss/gokit-starter/internal/modules/auth/domain"
	"github.com/stretchr/testify/require"
)

// mock implementations
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
	issueFunc func(_ context.Context, _ uuid.UUID, _ string) (string, error)
}

func (m *mockToken) Issue(ctx context.Context, id uuid.UUID, email string) (string, error) {
	return m.issueFunc(ctx, id, email)
}

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
		Organization: domain.Organization{ID: orgID, Name: "Org", Code: "org-123", Type: "company", Status: "active"},
	}

	// Register success
	t.Run("Register success", func(t *testing.T) {
		repo := &mockRepo{
			createAccount: func(_ context.Context, input domain.RegisterInput, passwordHash string) (domain.User, error) {
				if passwordHash == "" {
					return domain.User{}, errors.New("missing password hash")
				}
				return sampleUser, nil
			},
		}

		hasher := &mockHasher{hashFunc: func(_ string) (string, error) { return "hashed", nil }}
		token := &mockToken{
			issueFunc: func(_ context.Context, _ uuid.UUID, _ string) (string, error) {
				return "tok-123", nil
			},
		}

		svc := New(repo, hasher, token, 3600)

		sess, prof, err := svc.Register(ctx, domain.RegisterInput{Name: "Alice", Email: "alice@example.com", Password: "secret", OrganizationName: "Org"})
		require.NoError(t, err)
		require.Equal(t, "tok-123", sess.AccessToken)
		require.Equal(t, sampleUser.Email, prof.Email)
	})

	// Login success

	t.Run("Login success", func(t *testing.T) {
		repo := &mockRepo{
			findByEmail: func(_ context.Context, _ string) (domain.User, error) { return sampleUser, nil },
		}
		hasher := &mockHasher{compareFunc: func(_ string, _ string) error { return nil }}
		token := &mockToken{issueFunc: func(_ context.Context, _ uuid.UUID, _ string) (string, error) { return "tok-login", nil }}
		svc := New(repo, hasher, token, 3600)

		sess, prof, err := svc.Login(ctx, domain.Credentials{Email: sampleUser.Email, Password: "secret"})
		require.NoError(t, err)
		require.Equal(t, "tok-login", sess.AccessToken)
		require.Equal(t, sampleUser.Email, prof.Email)
	})

	// Login invalid credentials

	t.Run("Login invalid credentials", func(t *testing.T) {
		repo := &mockRepo{
			findByEmail: func(_ context.Context, _ string) (domain.User, error) { return sampleUser, nil },
		}
		hasher := &mockHasher{compareFunc: func(_ string, _ string) error { return errors.New("mismatch") }}
		token := &mockToken{issueFunc: func(_ context.Context, _ uuid.UUID, _ string) (string, error) { return "", nil }}
		svc := New(repo, hasher, token, 3600)

		_, _, err := svc.Login(ctx, domain.Credentials{Email: sampleUser.Email, Password: "badpass"})
		require.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	// Profile user not found

	t.Run("Profile not found", func(t *testing.T) {
		repo := &mockRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (domain.User, error) {
				return domain.User{}, domain.ErrUserNotFound
			},
		}
		hasher := &mockHasher{}
		token := &mockToken{}
		svc := New(repo, hasher, token, 3600)

		_, err := svc.Profile(ctx, uuid.New())
		require.ErrorIs(t, err, domain.ErrUserNotFound)
	})
}
