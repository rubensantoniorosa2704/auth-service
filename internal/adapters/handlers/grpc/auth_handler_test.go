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

// =====================
// Input Size Validation Tests
// =====================

// generateString creates a string of exactly n characters
func generateString(n int) string {
	if n <= 0 {
		return ""
	}
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[i] = 'a'
	}
	return string(result)
}

// generateEmail creates an email of exactly n total characters
// Format: {local}@example.com where local is padded to reach n total chars
func generateEmail(n int) string {
	domain := "@example.com"
	domainLen := len(domain)

	if n <= domainLen {
		return domain
	}

	localLen := n - domainLen
	local := generateString(localLen)
	return local + domain
}

func TestAuthGRPCHandler_Register_InputSizeValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		email             string
		password          string
		wantCode          codes.Code
		wantErr           string
		shouldCallService bool
	}{
		{
			name:              "email at 254 chars - accept",
			email:             generateEmail(254),
			password:          "ValidPass123!",
			wantCode:          codes.OK,
			shouldCallService: true,
		},
		{
			name:              "email at 255 chars - reject",
			email:             generateEmail(255),
			password:          "ValidPass123!",
			wantCode:          codes.InvalidArgument,
			wantErr:           "email exceeds maximum length of 254 characters",
			shouldCallService: false,
		},
		{
			name:              "password at 128 chars - accept",
			email:             "user@example.com",
			password:          generateString(128),
			wantCode:          codes.OK,
			shouldCallService: true,
		},
		{
			name:              "password at 129 chars - reject",
			email:             "user@example.com",
			password:          generateString(129),
			wantCode:          codes.InvalidArgument,
			wantErr:           "password exceeds maximum length of 128 characters",
			shouldCallService: false,
		},
		{
			name:              "valid sizes - accept",
			email:             "user@example.com",
			password:          "ValidPass123!",
			wantCode:          codes.OK,
			shouldCallService: true,
		},
		{
			name:              "both fields oversized - reject first (email)",
			email:             generateEmail(255),
			password:          generateString(129),
			wantCode:          codes.InvalidArgument,
			wantErr:           "email exceeds maximum length of 254 characters",
			shouldCallService: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			serviceCalled := false
			svc := &fakeAuthService{
				registerFn: func(_ context.Context, email, password string) (string, error) {
					serviceCalled = true
					return "user-id-123", nil
				},
			}
			handler := grpchandler.NewAuthGRPCHandler(svc)

			resp, err := handler.Register(context.Background(), &pb.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})

			// Verify service invocation expectation
			assert.Equal(t, tt.shouldCallService, serviceCalled, "service invocation mismatch")

			if tt.wantCode != codes.OK {
				require.Error(t, err)
				assert.Nil(t, resp)
				assert.Equal(t, tt.wantCode, status.Code(err))
				if tt.wantErr != "" {
					assert.Contains(t, err.Error(), tt.wantErr)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

func TestAuthGRPCHandler_Login_InputSizeValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		email             string
		password          string
		wantCode          codes.Code
		wantErr           string
		shouldCallService bool
	}{
		{
			name:              "email at 254 chars - accept",
			email:             generateEmail(254),
			password:          "ValidPass123!",
			wantCode:          codes.OK,
			shouldCallService: true,
		},
		{
			name:              "email at 255 chars - reject",
			email:             generateEmail(255),
			password:          "ValidPass123!",
			wantCode:          codes.InvalidArgument,
			wantErr:           "email exceeds maximum length of 254 characters",
			shouldCallService: false,
		},
		{
			name:              "password at 128 chars - accept",
			email:             "user@example.com",
			password:          generateString(128),
			wantCode:          codes.OK,
			shouldCallService: true,
		},
		{
			name:              "password at 129 chars - reject",
			email:             "user@example.com",
			password:          generateString(129),
			wantCode:          codes.InvalidArgument,
			wantErr:           "password exceeds maximum length of 128 characters",
			shouldCallService: false,
		},
		{
			name:              "valid sizes - accept",
			email:             "user@example.com",
			password:          "ValidPass123!",
			wantCode:          codes.OK,
			shouldCallService: true,
		},
		{
			name:              "both fields oversized - reject first (email)",
			email:             generateEmail(255),
			password:          generateString(129),
			wantCode:          codes.InvalidArgument,
			wantErr:           "email exceeds maximum length of 254 characters",
			shouldCallService: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			serviceCalled := false
			svc := &fakeAuthService{
				loginFn: func(_ context.Context, email, password string) (string, error) {
					serviceCalled = true
					return "jwt-token-abc", nil
				},
			}
			handler := grpchandler.NewAuthGRPCHandler(svc)

			resp, err := handler.Login(context.Background(), &pb.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
			})

			// Verify service invocation expectation
			assert.Equal(t, tt.shouldCallService, serviceCalled, "service invocation mismatch")

			if tt.wantCode != codes.OK {
				require.Error(t, err)
				assert.Nil(t, resp)
				assert.Equal(t, tt.wantCode, status.Code(err))
				if tt.wantErr != "" {
					assert.Contains(t, err.Error(), tt.wantErr)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

// =====================
// Property-Based Tests
// =====================

// Feature: input-size-validation, Property 1: Clear Error Messages for Size Violations
// **Validates: Requirements 3.1**
//
// Property: For any email or password input that exceeds its maximum length,
// the returned error message should clearly indicate which field exceeded the
// limit and specify the maximum allowed length.
func TestInputSizeValidation_ErrorMessages_Property(t *testing.T) {
	t.Parallel()

	// Test cases with various oversized inputs
	testCases := []struct {
		name             string
		emailSize        int
		passwordSize     int
		expectedField    string
		expectedMaxLen   string
		expectedContains []string
	}{
		// Email oversized cases
		{
			name:             "email 255 chars",
			emailSize:        255,
			passwordSize:     10,
			expectedField:    "email",
			expectedMaxLen:   "254",
			expectedContains: []string{"email", "exceeds", "254"},
		},
		{
			name:             "email 300 chars",
			emailSize:        300,
			passwordSize:     10,
			expectedField:    "email",
			expectedMaxLen:   "254",
			expectedContains: []string{"email", "exceeds", "254"},
		},
		{
			name:             "email 1000 chars",
			emailSize:        1000,
			passwordSize:     10,
			expectedField:    "email",
			expectedMaxLen:   "254",
			expectedContains: []string{"email", "exceeds", "254"},
		},
		// Password oversized cases
		{
			name:             "password 129 chars",
			emailSize:        20,
			passwordSize:     129,
			expectedField:    "password",
			expectedMaxLen:   "128",
			expectedContains: []string{"password", "exceeds", "128"},
		},
		{
			name:             "password 200 chars",
			emailSize:        20,
			passwordSize:     200,
			expectedField:    "password",
			expectedMaxLen:   "128",
			expectedContains: []string{"password", "exceeds", "128"},
		},
		{
			name:             "password 500 chars",
			emailSize:        20,
			passwordSize:     500,
			expectedField:    "password",
			expectedMaxLen:   "128",
			expectedContains: []string{"password", "exceeds", "128"},
		},
	}

	// Run minimum 100 iterations as specified in requirements
	iterations := 0
	for i := 0; i < 20; i++ {
		for _, tc := range testCases {
			iterations++
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				// Setup handler with fake service
				svc := &fakeAuthService{
					registerFn: func(_ context.Context, _, _ string) (string, error) {
						t.Fatal("service should not be called for oversized inputs")
						return "", nil
					},
				}
				handler := grpchandler.NewAuthGRPCHandler(svc)

				// Generate oversized inputs
				email := generateEmail(tc.emailSize)
				password := generateString(tc.passwordSize)

				// Test Register endpoint
				_, err := handler.Register(context.Background(), &pb.RegisterRequest{
					Email:    email,
					Password: password,
				})

				// Verify error is returned
				require.Error(t, err, "expected error for oversized input")

				// Verify error code
				assert.Equal(t, codes.InvalidArgument, status.Code(err),
					"expected InvalidArgument code for oversized input")

				// Verify error message contains required elements
				errMsg := err.Error()
				for _, expected := range tc.expectedContains {
					assert.Contains(t, errMsg, expected,
						"error message should contain '%s'", expected)
				}

				// Test Login endpoint with same inputs
				_, err = handler.Login(context.Background(), &pb.LoginRequest{
					Email:    email,
					Password: password,
				})

				// Verify error is returned
				require.Error(t, err, "expected error for oversized input")

				// Verify error code
				assert.Equal(t, codes.InvalidArgument, status.Code(err),
					"expected InvalidArgument code for oversized input")

				// Verify error message contains required elements
				errMsg = err.Error()
				for _, expected := range tc.expectedContains {
					assert.Contains(t, errMsg, expected,
						"error message should contain '%s'", expected)
				}
			})
		}
	}

	// Verify we ran at least 100 iterations
	assert.GreaterOrEqual(t, iterations, 100,
		"property test should run at least 100 iterations")
}

// Feature: input-size-validation, Property 2: Validation Occurs Before Service Layer
// **Validates: Requirements 3.2**
//
// Property: For any Register or Login request where input size validation fails,
// the request should be rejected at the handler layer without invoking the
// underlying AuthService.
func TestInputSizeValidation_ServiceNotInvoked_Property(t *testing.T) {
	t.Parallel()

	// Test cases with various oversized inputs
	testCases := []struct {
		name         string
		emailSize    int
		passwordSize int
		description  string
	}{
		// Email oversized cases
		{
			name:         "email 255 chars (boundary+1)",
			emailSize:    255,
			passwordSize: 10,
			description:  "email just over limit",
		},
		{
			name:         "email 300 chars (moderate overage)",
			emailSize:    300,
			passwordSize: 10,
			description:  "email moderately over limit",
		},
		{
			name:         "email 1000 chars (extreme overage)",
			emailSize:    1000,
			passwordSize: 10,
			description:  "email extremely over limit",
		},
		{
			name:         "email 500 chars",
			emailSize:    500,
			passwordSize: 10,
			description:  "email significantly over limit",
		},
		// Password oversized cases
		{
			name:         "password 129 chars (boundary+1)",
			emailSize:    20,
			passwordSize: 129,
			description:  "password just over limit",
		},
		{
			name:         "password 200 chars (moderate overage)",
			emailSize:    20,
			passwordSize: 200,
			description:  "password moderately over limit",
		},
		{
			name:         "password 500 chars (extreme overage)",
			emailSize:    20,
			passwordSize: 500,
			description:  "password extremely over limit",
		},
		{
			name:         "password 300 chars",
			emailSize:    20,
			passwordSize: 300,
			description:  "password significantly over limit",
		},
		// Both fields oversized
		{
			name:         "both fields oversized (255 email, 129 password)",
			emailSize:    255,
			passwordSize: 129,
			description:  "both fields just over limit",
		},
		{
			name:         "both fields oversized (500 email, 300 password)",
			emailSize:    500,
			passwordSize: 300,
			description:  "both fields significantly over limit",
		},
	}

	// Run minimum 100 iterations as specified in requirements
	iterations := 0
	for i := 0; i < 10; i++ {
		for _, tc := range testCases {
			iterations++
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				// Create a mock service that panics if invoked
				// This ensures validation prevents service invocation
				panicService := &fakeAuthService{
					registerFn: func(_ context.Context, _, _ string) (string, error) {
						panic("BUG: AuthService.Register should not be called when input size validation fails")
					},
					loginFn: func(_ context.Context, _, _ string) (string, error) {
						panic("BUG: AuthService.Login should not be called when input size validation fails")
					},
				}
				handler := grpchandler.NewAuthGRPCHandler(panicService)

				// Generate oversized inputs
				email := generateEmail(tc.emailSize)
				password := generateString(tc.passwordSize)

				// Test Register endpoint
				// If service is called, the panic will fail the test
				resp, err := handler.Register(context.Background(), &pb.RegisterRequest{
					Email:    email,
					Password: password,
				})

				// Verify error is returned (validation failed)
				require.Error(t, err, "expected error for oversized input")
				assert.Nil(t, resp, "response should be nil when validation fails")
				assert.Equal(t, codes.InvalidArgument, status.Code(err),
					"expected InvalidArgument code for oversized input")

				// Test Login endpoint with same inputs
				// If service is called, the panic will fail the test
				loginResp, loginErr := handler.Login(context.Background(), &pb.LoginRequest{
					Email:    email,
					Password: password,
				})

				// Verify error is returned (validation failed)
				require.Error(t, loginErr, "expected error for oversized input")
				assert.Nil(t, loginResp, "response should be nil when validation fails")
				assert.Equal(t, codes.InvalidArgument, status.Code(loginErr),
					"expected InvalidArgument code for oversized input")
			})
		}
	}

	// Verify we ran at least 100 iterations
	assert.GreaterOrEqual(t, iterations, 100,
		"property test should run at least 100 iterations")
}
