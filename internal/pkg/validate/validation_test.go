package validate

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestRequest struct {
	Name  string `json:"name"  validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

func TestDecodeAndValidate(t *testing.T) {
	v := New()

	t.Run("success", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name": "Alice", "email": "alice@example.com"}`)
		req := httptest.NewRequest(http.MethodPost, "/", body)
		req.Header.Set("Content-Type", "application/json")

		dst, err := DecodeAndValidate[TestRequest](v, req)
		require.NoError(t, err)
		require.Equal(t, "Alice", dst.Name)
		require.Equal(t, "alice@example.com", dst.Email)
	})

	t.Run("decode error (bad json)", func(t *testing.T) {
		body := bytes.NewBufferString(`{bad json}`)
		req := httptest.NewRequest(http.MethodPost, "/", body)
		req.Header.Set("Content-Type", "application/json")

		_, err := DecodeAndValidate[TestRequest](v, req)
		require.Error(t, err)
	})

	t.Run("validation error", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name": "", "email": "not-an-email"}`)
		req := httptest.NewRequest(http.MethodPost, "/", body)
		req.Header.Set("Content-Type", "application/json")

		_, err := DecodeAndValidate[TestRequest](v, req)
		require.Error(t, err)
		require.IsType(t, Error{}, err)
	})
}
