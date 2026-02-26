package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/rubensantoniorosa2704/auth-service/internal/adapters/db"
	"github.com/rubensantoniorosa2704/auth-service/internal/adapters/encryption"
	grpcHandler "github.com/rubensantoniorosa2704/auth-service/internal/adapters/handlers/grpc"
	"github.com/rubensantoniorosa2704/auth-service/internal/adapters/tokens"
	"github.com/rubensantoniorosa2704/auth-service/internal/core/services"
	pb "github.com/rubensantoniorosa2704/auth-service/proto/auth/v1"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database
	dbCfg := db.Config{
		Host:            requireEnv("POSTGRES_HOST"),
		Port:            requireEnv("POSTGRES_PORT"),
		User:            requireEnv("POSTGRES_USER"),
		Password:        requireEnv("POSTGRES_PASSWORD"),
		DBName:          requireEnv("POSTGRES_DB"),
		SSLMode:         getEnv("DB_SSLMODE", "disable"),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	pool, err := db.NewConnection(ctx, dbCfg)
	if err != nil {
		slog.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// Adapters
	userRepo := db.NewPostgresUserRepository(pool)
	hasher := encryption.NewArgon2Hasher()

	jwtSecret := requireEnv("JWT_SECRET")

	// Load JWT expiration hours from environment variable
	// Default to 1 hour per OWASP security guidelines
	jwtExpirationHoursStr := getEnv("JWT_EXPIRATION_HOURS", "1")
	jwtExpirationHours, err := strconv.Atoi(jwtExpirationHoursStr)
	if err != nil {
		slog.Error("invalid JWT_EXPIRATION_HOURS value",
			slog.String("value", jwtExpirationHoursStr),
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	tokenSvc, err := tokens.NewJWTService(jwtSecret, "auth-service", jwtExpirationHours)
	if err != nil {
		slog.Error("failed to initialize JWT service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Service
	authService := services.NewAuthService(userRepo, hasher, tokenSvc, logger)
	authHandler := grpcHandler.NewAuthGRPCHandler(authService)

	// gRPC server
	grpcPort := getEnv("GRPC_PORT", "50051")

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		slog.Error("failed to listen", slog.String("port", grpcPort), slog.String("error", err.Error()))
		os.Exit(1)
	}

	server := grpc.NewServer()
	pb.RegisterAuthServiceServer(server, authHandler)
	reflection.Register(server)

	go func() {
		slog.Info("server started", slog.String("port", grpcPort), slog.String("transport", "grpc"))
		if err := server.Serve(lis); err != nil {
			slog.Error("grpc serve failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sig := <-stop
	slog.Info("shutdown signal received", slog.String("signal", sig.String()))

	server.GracefulStop()
	cancel()
	slog.Info("server stopped gracefully")
}

// requireEnv returns the value of a required environment variable or exits if not set.
func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("required environment variable not set", slog.String("variable", key))
		os.Exit(1)
	}
	return v
}

// getEnv returns the value of an environment variable or a default fallback.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
