//go:build integration

package infra

import (
	"context"
	"testing"

	"github.com/masmuss/gokit-starter/internal/modules/auth/domain"
	"github.com/masmuss/gokit-starter/internal/test"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func TestRepository_CreateAccount(t *testing.T) {
	client := test.NewEntClient(t)
	defer client.Close()

	repo := NewRepository(client)
	ctx := context.Background()

	t.Run("Create Personal Account", func(t *testing.T) {
		input := domain.RegisterInput{
			Name:             "Personal User",
			Email:            "personal@example.com",
			Password:         "secret123",
			OrganizationName: "", // Empty for personal
		}

		user, err := repo.CreateAccount(ctx, input, "hashed_pass")
		require.NoError(t, err)
		require.Equal(t, "Personal User", user.Name)
		require.Equal(t, "personal@example.com", user.Email)
		require.Equal(t, "personal", user.Organization.Type)
		require.Equal(t, "Personal User", user.Organization.Name)
	})

	t.Run("Create Company Account", func(t *testing.T) {
		input := domain.RegisterInput{
			Name:             "Admin User",
			Email:            "admin@company.com",
			Password:         "secret123",
			OrganizationName: "My Company",
		}

		user, err := repo.CreateAccount(ctx, input, "hashed_pass")
		require.NoError(t, err)
		require.Equal(t, "Admin User", user.Name)
		require.Equal(t, "company", user.Organization.Type)
		require.Equal(t, "My Company", user.Organization.Name)
	})
}
