package grpc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpchandler "github.com/rubensantoniorosa2704/auth-service/internal/adapters/handlers/grpc"
	"github.com/rubensantoniorosa2704/auth-service/internal/core/domain"
	pb "github.com/rubensantoniorosa2704/auth-service/proto/auth/v1"
)

// =====================
// Fake AuthService
// =====================

type fakeAuthService struct {
	registerFn func(ctx context.Context, email, password string) (string, error)
	loginFn    func(ctx context.Context, email, password string) (string, error)
}

func (f *fakeAuthService) Register(ctx context.Context, email, password string) (string, error) {
	return f.registerFn(ctx, email, password)
}

func (f *fakeAuthService) Login(ctx context.Context, email, password string) (string, error) {
	return f.loginFn(ctx, email, password)
}

// =====================
// Register tests
// =====================

func TestAuthGRPCHandler_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		req       *pb.RegisterRequest
		serviceFn func(ctx context.Context, email, password string) (string, error)
		wantCode  codes.Code
		wantResp  *pb.RegisterResponse
	}{
		{
			name: "success",
			req:  &pb.RegisterRequest{Email: "user@example.com", Password: "securepassword"},
			serviceFn: func(_ context.Context, email, _ string) (string, error) {
				return "user-id-123", nil
			},
			wantCode: codes.OK,
			wantResp: &pb.RegisterResponse{
				UserId: "user-id-123",
				Email:  "user@example.com",
			},
		},
		{
			name: "duplicate email returns AlreadyExists",
			req:  &pb.RegisterRequest{Email: "dup@example.com", Password: "securepassword"},
			serviceFn: func(_ context.Context, _, _ string) (string, error) {
				return "", domain.ErrUserAlreadyExists
			},
			wantCode: codes.AlreadyExists,
		},
		{
			name: "invalid email returns InvalidArgument",
			req:  &pb.RegisterRequest{Email: "not-an-email", Password: "securepassword"},
			serviceFn: func(_ context.Context, _, _ string) (string, error) {
				return "", domain.ErrInvalidEmail
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "weak password returns InvalidArgument",
			req:  &pb.RegisterRequest{Email: "user@example.com", Password: "123"},
			serviceFn: func(_ context.Context, _, _ string) (string, error) {
				return "", domain.ErrWeakPassword
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "unexpected error returns Internal",
			req:  &pb.RegisterRequest{Email: "user@example.com", Password: "securepassword"},
			serviceFn: func(_ context.Context, _, _ string) (string, error) {
				return "", assert.AnError
			},
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := &fakeAuthService{registerFn: tt.serviceFn}
			handler := grpchandler.NewAuthGRPCHandler(svc)

			resp, err := handler.Register(context.Background(), tt.req)

			if tt.wantCode != codes.OK {
				require.Error(t, err)
				assert.Nil(t, resp)
				assert.Equal(t, tt.wantCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.wantResp.UserId, resp.UserId)
			assert.Equal(t, tt.wantResp.Email, resp.Email)
		})
	}
}

// =====================
// Login tests
// =====================

func TestAuthGRPCHandler_Login(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		req       *pb.LoginRequest
		serviceFn func(ctx context.Context, email, password string) (string, error)
		wantCode  codes.Code
		wantToken string
	}{
		{
			name: "success",
			req:  &pb.LoginRequest{Email: "user@example.com", Password: "securepassword"},
			serviceFn: func(_ context.Context, _, _ string) (string, error) {
				return "jwt-token-abc", nil
			},
			wantCode:  codes.OK,
			wantToken: "jwt-token-abc",
		},
		{
			name: "invalid credentials returns Unauthenticated",
			req:  &pb.LoginRequest{Email: "user@example.com", Password: "wrongpassword"},
			serviceFn: func(_ context.Context, _, _ string) (string, error) {
				return "", domain.ErrInvalidCredentials
			},
			wantCode: codes.Unauthenticated,
		},
		{
			name: "user not found returns NotFound",
			req:  &pb.LoginRequest{Email: "ghost@example.com", Password: "securepassword"},
			serviceFn: func(_ context.Context, _, _ string) (string, error) {
				return "", domain.ErrUserNotFound
			},
			wantCode: codes.NotFound,
		},
		{
			name: "unexpected error returns Internal",
			req:  &pb.LoginRequest{Email: "user@example.com", Password: "securepassword"},
			serviceFn: func(_ context.Context, _, _ string) (string, error) {
				return "", assert.AnError
			},
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := &fakeAuthService{loginFn: tt.serviceFn}
			handler := grpchandler.NewAuthGRPCHandler(svc)

			resp, err := handler.Login(context.Background(), tt.req)

			if tt.wantCode != codes.OK {
				require.Error(t, err)
				assert.Nil(t, resp)
				assert.Equal(t, tt.wantCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.wantToken, resp.AccessToken)
		})
	}
}

// =====================
// mapDomainError tests
// =====================

func TestAuthGRPCHandler_ErrorMapping(t *testing.T) {
	t.Parallel()

	// Each domain error must map to exactly one gRPC code.
	// This test makes the mapping explicit and catches regressions.
	tests := []struct {
		domainErr error
		wantCode  codes.Code
	}{
		{domain.ErrUserAlreadyExists, codes.AlreadyExists},
		{domain.ErrInvalidEmail, codes.InvalidArgument},
		{domain.ErrWeakPassword, codes.InvalidArgument},
		{domain.ErrInvalidCredentials, codes.Unauthenticated},
		{domain.ErrUserNotFound, codes.NotFound},
		{domain.ErrInvalidUUID, codes.InvalidArgument},
		{assert.AnError, codes.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.domainErr.Error(), func(t *testing.T) {
			t.Parallel()

			// We exercise mapDomainError indirectly via Register,
			// since the function is unexported.
			svc := &fakeAuthService{
				registerFn: func(_ context.Context, _, _ string) (string, error) {
					return "", tt.domainErr
				},
			}
			handler := grpchandler.NewAuthGRPCHandler(svc)

			_, err := handler.Register(context.Background(), &pb.RegisterRequest{
				Email:    "user@example.com",
				Password: "securepassword",
			})

			require.Error(t, err)
			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}
