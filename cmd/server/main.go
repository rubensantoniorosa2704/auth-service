package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
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
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables.")
	}

	dbCfg := db.Config{
		Host:            os.Getenv("DB_HOST"),
		Port:            os.Getenv("DB_PORT"),
		User:            os.Getenv("DB_USER"),
		Password:        os.Getenv("DB_PASSWORD"),
		DBName:          os.Getenv("DB_NAME"),
		SSLMode:         os.Getenv("DB_SSLMODE"),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	pool, err := db.NewConnection(dbCfg)
	if err != nil {
		log.Fatalf("❌ Fatal error connecting to database: %v", err)
	}
	defer pool.Close()

	userRepo := db.NewPostgresUserRepository(pool)
	hasher := encryption.NewArgon2Hasher()
	tokenSvc := tokens.NewJWTService(os.Getenv("JWT_SECRET"), "auth-service")

	authService := services.NewAuthService(userRepo, hasher, tokenSvc)
	authHandler := grpcHandler.NewAuthGRPCHandler(authService)

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("❌ Failed to listen on port %s: %v", grpcPort, err)
	}

	server := grpc.NewServer()
	pb.RegisterAuthServiceServer(server, authHandler)
	reflection.Register(server)

	go func() {
		log.Printf("🚀 Auth-Service running on port %s", grpcPort)
		if err := server.Serve(lis); err != nil {
			log.Fatalf("❌ Failed to serve gRPC: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	log.Println("\n🛑 Gracefully shutting down server...")
	server.GracefulStop()
	log.Println("👋 Server stopped.")
}
