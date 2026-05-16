package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Validation(t *testing.T) {
	// Helper to clear env after test
	clearEnv := func() {
		os.Unsetenv("APP_ENV")
		os.Unsetenv("AUTH_JWT_SECRET")
		os.Unsetenv("DB_PORT")
	}

	t.Run("Success with valid env", func(t *testing.T) {
		defer clearEnv()
		os.Setenv("APP_ENV", "production")
		os.Setenv("APP_URL", "https://example.com")
		os.Setenv("AUTH_JWT_SECRET", "this-is-a-very-long-secret-key-32-chars")
		
		cfg, err := LoadConfig()
		require.NoError(t, err)
		require.Equal(t, "production", cfg.App.Env)
	})

	t.Run("Fail with invalid APP_ENV", func(t *testing.T) {
		defer clearEnv()
		os.Setenv("APP_ENV", "invalid_env")
		
		_, err := LoadConfig()
		require.Error(t, err)
		require.Contains(t, err.Error(), "config validation failed")
	})

	t.Run("Fail with short JWT secret", func(t *testing.T) {
		defer clearEnv()
		os.Setenv("AUTH_JWT_SECRET", "short")
		
		_, err := LoadConfig()
		require.Error(t, err)
		require.Contains(t, err.Error(), "JWTSecret")
	})

	t.Run("Fail with invalid Port", func(t *testing.T) {
		defer clearEnv()
		os.Setenv("APP_PORT", "-1")
		
		_, err := LoadConfig()
		require.Error(t, err)
	})
}
