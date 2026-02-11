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

type AuthGRPCHandler struct {
	pb.UnimplementedAuthServiceServer
	authService *services.AuthService
}

func NewAuthGRPCHandler(svc *services.AuthService) *AuthGRPCHandler {
	return &AuthGRPCHandler{
		authService: svc,
	}
}

func (h *AuthGRPCHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	err := h.authService.Register(req.GetEmail(), req.GetPassword())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, "este e-mail já está cadastrado")
		case err.Error() == "invalid email": // Ou se você tiver um ErrInvalidEmail no domain
			return nil, status.Error(codes.InvalidArgument, "formato de e-mail inválido")
		default:
			return nil, status.Error(codes.Internal, "erro interno ao processar registro")
		}
	}

	return &pb.RegisterResponse{
		Email: req.GetEmail(),
	}, nil
}

func (h *AuthGRPCHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	token, err := h.authService.Login(req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "credenciais inválidas")
		}
		return nil, status.Error(codes.Internal, "erro interno ao processar login")
	}

	return &pb.LoginResponse{
		AccessToken: token,
	}, nil
}
