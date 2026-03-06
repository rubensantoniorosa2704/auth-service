package grpc

import (
	"context"
	"errors"

	"github.com/rubensantoniorosa2704/auth-service/internal/core/domain"
	"github.com/rubensantoniorosa2704/auth-service/internal/core/services"
	pb "github.com/rubensantoniorosa2704/auth-service/proto/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	maxEmailLength    = 254 // RFC 5321 maximum
	maxPasswordLength = 128 // Reasonable upper bound
)

// AuthGRPCHandler translates gRPC requests into application service calls.
type AuthGRPCHandler struct {
	pb.UnimplementedAuthServiceServer
	authService services.AuthServicer
}

// NewAuthGRPCHandler creates a handler that depends on the AuthServicer interface.
func NewAuthGRPCHandler(svc services.AuthServicer) *AuthGRPCHandler {
	return &AuthGRPCHandler{
		authService: svc,
	}
}

// validateInputSizes checks that email and password do not exceed maximum lengths.
// Returns a gRPC InvalidArgument error if validation fails.
func validateInputSizes(email, password string) error {
	if len(email) > maxEmailLength {
		return status.Errorf(codes.InvalidArgument, "email exceeds maximum length of 254 characters")
	}
	if len(password) > maxPasswordLength {
		return status.Errorf(codes.InvalidArgument, "password exceeds maximum length of 128 characters")
	}
	return nil
}

// Register handles user registration requests.
func (h *AuthGRPCHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// Validate input sizes before processing
	if err := validateInputSizes(req.GetEmail(), req.GetPassword()); err != nil {
		return nil, err
	}

	userID, err := h.authService.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &pb.RegisterResponse{
		UserId: userID,
		Email:  req.GetEmail(),
	}, nil
}

// Login handles user authentication requests.
func (h *AuthGRPCHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Validate input sizes before processing
	if err := validateInputSizes(req.GetEmail(), req.GetPassword()); err != nil {
		return nil, err
	}

	token, err := h.authService.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &pb.LoginResponse{
		AccessToken: token,
	}, nil
}

// mapDomainError translates domain errors into appropriate gRPC status codes.
func mapDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, "email already registered")
	case errors.Is(err, domain.ErrInvalidEmail):
		return status.Error(codes.InvalidArgument, "invalid email format")
	case errors.Is(err, domain.ErrWeakPassword):
		return status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
