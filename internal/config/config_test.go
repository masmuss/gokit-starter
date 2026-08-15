package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid environment",
			env: map[string]string{
				"APP_ENV":         "production",
				"APP_URL":         "https://example.com",
				"AUTH_JWT_SECRET": "this-is-a-very-long-secret-key-32-chars",
			},
			wantErr: false,
		},
		{
			name: "invalid APP_ENV",
			env: map[string]string{
				"APP_ENV": "invalid_env",
			},
			wantErr: true,
			errMsg:  "config validation failed",
		},
		{
			name: "short JWT secret",
			env: map[string]string{
				"AUTH_JWT_SECRET": "short",
				"APP_ENV":         "production",
				"APP_URL":         "https://example.com",
			},
			wantErr: true,
			errMsg:  "JWTSecret",
		},
		{
			name: "invalid APP_PORT",
			env: map[string]string{
				"APP_PORT": "-1",
				"APP_ENV":  "production",
				"APP_URL":  "https://example.com",
			},
			wantErr: true,
		},
		{
			name: "refresh TTL less than access TTL",
			env: map[string]string{
				"APP_ENV":              "production",
				"APP_URL":              "https://example.com",
				"AUTH_JWT_SECRET":      "this-is-a-very-long-secret-key-32-chars",
				"AUTH_JWT_TTL":         "120",
				"AUTH_JWT_REFRESH_TTL": "60",
			},
			wantErr: true,
			errMsg:  "must be greater than",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.env {
					os.Unsetenv(k)
				}
			}()

			_, err := LoadConfig()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
