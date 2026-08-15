package handler

import (
	"net/http"

	"github.com/swaggest/openapi-go/openapi31"

	"github.com/masmuss/gokit-starter/internal/inbound/response"
	"github.com/masmuss/gokit-starter/internal/pkg/doc"
)

// AuthDocRegistrar registers auth endpoints in the OpenAPI spec.
type AuthDocRegistrar struct{}

// NewAuthDocRegistrar creates a new auth doc registrar.
func NewAuthDocRegistrar() *AuthDocRegistrar {
	return &AuthDocRegistrar{}
}

// RegisterOperations implements doc.OperationRegistrar.
func (a *AuthDocRegistrar) RegisterOperations(r *openapi31.Reflector) error {
	ops := []struct {
		method, path, summary, description string
		tags                               []string
		req                                any
		resps                              []doc.RespSpec
		secured                            bool
	}{
		{
			method: http.MethodPost, path: "/auth/register",
			summary: "Register a new user", description: "Register a new user and create an organization",
			tags: []string{"auth"}, req: RegisterRequest{},
			resps: []doc.RespSpec{
				{Status: http.StatusCreated, StructType: response.Envelope{}},
				{Status: http.StatusBadRequest, StructType: response.ErrorEnvelope{}},
				{Status: http.StatusConflict, StructType: response.ErrorEnvelope{}},
			},
		},
		{
			method: http.MethodPost, path: "/auth/login",
			summary: "User login", description: "Authenticate user and return JWT token",
			tags: []string{"auth"}, req: LoginRequest{},
			resps: []doc.RespSpec{
				{Status: http.StatusOK, StructType: response.Envelope{}},
				{Status: http.StatusUnauthorized, StructType: response.ErrorEnvelope{}},
			},
		},
		{
			method: http.MethodGet, path: "/auth/profile",
			summary: "Get current user profile", description: "Retrieve profile details for the authenticated user",
			tags: []string{"auth"}, secured: true,
			resps: []doc.RespSpec{
				{Status: http.StatusOK, StructType: response.Envelope{}},
				{Status: http.StatusUnauthorized, StructType: response.ErrorEnvelope{}},
			},
		},
		{
			method: http.MethodPost, path: "/auth/logout",
			summary: "Logout", description: "Revoke access and refresh tokens",
			tags: []string{"auth"}, req: LogoutRequest{}, secured: true,
			resps: []doc.RespSpec{
				{Status: http.StatusOK, StructType: response.Envelope{}},
				{Status: http.StatusUnauthorized, StructType: response.ErrorEnvelope{}},
				{Status: http.StatusBadRequest, StructType: response.ErrorEnvelope{}},
			},
		},
		{
			method: http.MethodPut, path: "/auth/password",
			summary: "Change password", description: "Update password for the authenticated user",
			tags: []string{"auth"}, req: ChangePasswordRequest{}, secured: true,
			resps: []doc.RespSpec{
				{Status: http.StatusOK, StructType: response.Envelope{}},
				{Status: http.StatusBadRequest, StructType: response.ErrorEnvelope{}},
				{Status: http.StatusUnauthorized, StructType: response.ErrorEnvelope{}},
			},
		},
		{
			method: http.MethodPost, path: "/auth/refresh",
			summary: "Refresh access token", description: "Issue a new access token using a valid refresh token",
			tags: []string{"auth"}, req: RefreshTokenRequest{},
			resps: []doc.RespSpec{
				{Status: http.StatusOK, StructType: response.Envelope{}},
				{Status: http.StatusUnauthorized, StructType: response.ErrorEnvelope{}},
			},
		},
	}

	for _, op := range ops {
		if err := doc.AddOperation(
			r, op.method, op.path, op.summary, op.description,
			op.tags, op.req, op.resps, op.secured,
		); err != nil {
			return err
		}
	}

	return nil
}
