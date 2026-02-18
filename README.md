# Auth Service

[![Go Version](https://img.shields.io/badge/Go-1.25.6-00ADD8?style=flat&logo=go)](https://go.dev/)
[![gRPC](https://img.shields.io/badge/gRPC-Protocol%20Buffers-244c5a?style=flat&logo=grpc)](https://grpc.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](./LICENSE.md)

A production-grade authentication microservice built with Go, demonstrating clean architecture, gRPC communication, and industry-standard backend engineering practices.

> 🎯 **Portfolio Project** — Showcasing professional Go development with focus on security, maintainability, and architectural clarity.

---

## 📋 Table of Contents

- [Features](#-features)
- [Architecture](#-architecture)
- [Tech Stack](#-tech-stack)
- [Project Structure](#-project-structure)
- [Getting Started](#-getting-started)
- [API Usage](#-api-usage)
- [Development](#-development)
- [Testing](#-testing)
- [Environment Variables](#-environment-variables)
- [Design Decisions](#-design-decisions)
- [License](#-license)

---

## ✨ Features

- **User Registration** — Email/password validation using domain value objects
- **User Authentication** — Secure login with JWT token generation
- **Argon2id Password Hashing** — Memory-hard, timing-safe password storage
- **gRPC API** — Strongly typed contracts with Protocol Buffers
- **PostgreSQL Storage** — Type-safe queries with [sqlc](https://sqlc.dev)
- **Hexagonal Architecture** — Clean separation of domain, ports, and adapters
- **Structured Logging** — JSON output using Go's `log/slog`
- **Context Propagation** — Request context flows through all layers
- **Graceful Shutdown** — Signal-aware server lifecycle management
- **Timing Attack Protection** — Constant-time authentication checks
- **Docker Support** — Full containerization with Docker Compose

---

## 🏗️ Architecture

This project follows **Hexagonal Architecture (Ports & Adapters)** with strict layer separation:

```
┌─────────────────────────────────────────────────────────┐
│                    gRPC Handler                         │
│                   (Transport Layer)                     │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                  Auth Service                           │
│              (Application Layer)                        │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Domain Layer (Entities, Value Objects, Errors)  │  │
│  └──────────────────────────────────────────────────┘  │
└────────┬────────────────────┬────────────────┬─────────┘
         │                    │                │
         ▼                    ▼                ▼
┌────────────────┐  ┌─────────────────┐  ┌──────────────┐
│   PostgreSQL   │  │  Argon2 Hasher  │  │  JWT Service │
│   Repository   │  │                 │  │              │
└────────────────┘  └─────────────────┘  └──────────────┘
```

**Key Principles:**
- Domain logic has zero external dependencies
- All dependencies point inward toward the domain
- Adapters implement port interfaces defined by the core
- Easy to test, maintain, and extend

---

## 🛠️ Tech Stack

| Category       | Technology                                |
|----------------|-------------------------------------------|
| Language       | Go 1.25.6                                 |
| Transport      | gRPC + Protocol Buffers                   |
| Database       | PostgreSQL 16 (pgx v5 driver)             |
| Query Builder  | sqlc (type-safe SQL)                      |
| Authentication | JWT (golang-jwt/v5)                       |
| Password Hash  | Argon2id (golang.org/x/crypto)            |
| Logging        | log/slog (standard library, JSON)        |
| Infrastructure | Docker, Docker Compose                    |

**Key Dependencies:**
```
github.com/google/uuid
github.com/golang-jwt/jwt/v5
github.com/jackc/pgx/v5
github.com/joho/godotenv
golang.org/x/crypto
google.golang.org/grpc
google.golang.org/protobuf
```

---

## 📁 Project Structure

```
auth-service/
├── cmd/server/                    # Application entrypoint
│   └── main.go                    # Server initialization, dependency injection
│
├── internal/
│   ├── core/                      # Business logic (no external dependencies)
│   │   ├── domain/                # Entities, value objects, domain errors
│   │   │   ├── user.go
│   │   │   ├── email.go
│   │   │   ├── password.go
│   │   │   └── errors.go
│   │   ├── ports/                 # Interfaces (contracts)
│   │   │   ├── user_repository.go
│   │   │   ├── password_hasher.go
│   │   │   └── token_service.go
│   │   └── services/              # Use cases / application services
│   │       ├── auth_service.go
│   │       └── auth_service_test.go
│   │
│   └── adapters/                  # External integrations
│       ├── db/                    # PostgreSQL adapter
│       │   ├── connection.go
│       │   ├── postgres_repository.go
│       │   ├── queries/           # SQL definitions (sqlc input)
│       │   └── generated/         # sqlc generated code
│       ├── encryption/            # Argon2id hasher
│       ├── tokens/                # JWT service
│       └── handlers/grpc/         # gRPC transport handler
│
├── proto/auth/v1/                 # Protocol Buffer definitions
│   ├── auth.proto
│   ├── auth.pb.go                 # Generated protobuf code
│   └── auth_grpc.pb.go            # Generated gRPC code
│
├── database/migrations/           # SQL migration files
│   ├── 000001_create_users_table.up.sql
│   └── 000001_create_users_table.down.sql
│
├── Dockerfile                     # Multi-stage Docker build
├── docker-compose.yml             # Service orchestration
├── Makefile                       # Common development tasks
└── .env.example                   # Environment variables template
```

---

## 🚀 Getting Started

### Prerequisites

- **Go 1.22+** — [Install Go](https://go.dev/doc/install)
- **Docker & Docker Compose** — [Install Docker](https://docs.docker.com/get-docker/)
- **Make** (optional) — For convenience commands

### Quick Start with Docker

1. **Clone the repository**
   ```bash
   git clone https://github.com/rubensantoniorosa2704/auth-service.git
   cd auth-service
   ```

2. **Create environment file**
   ```bash
   cp .env.example .env
   ```

3. **Generate a secure JWT secret**
   ```bash
   openssl rand -base64 32
   ```
   Copy the output and update `JWT_SECRET` in `.env`

4. **Start all services**
   ```bash
   docker compose up --build
   ```

   This will:
   - Start PostgreSQL database
   - Run database migrations
   - Start the gRPC server on port `50051`

5. **Verify the server is running**
   ```bash
   docker compose logs server
   ```
   You should see:
   ```json
   {"level":"INFO","msg":"database connection established","host":"postgres","database":"auth"}
   {"level":"INFO","msg":"server started","port":"50051","transport":"grpc"}
   ```

### Running Locally (without Docker)

1. **Start PostgreSQL** (ensure it's running on `localhost:5432`)

2. **Set environment variables**
   ```bash
   export POSTGRES_HOST=localhost
   export POSTGRES_PORT=5432
   export POSTGRES_USER=postgres
   export POSTGRES_PASSWORD=your_password
   export POSTGRES_DB=auth
   export JWT_SECRET=$(openssl rand -base64 32)
   export GRPC_PORT=50051
   ```

3. **Run migrations**
   ```bash
   make migrate-up
   ```

4. **Generate sqlc code** (if needed)
   ```bash
   make sqlc
   ```

5. **Start the server**
   ```bash
   make server
   ```

---

## 📡 API Usage

The service exposes a gRPC API with two endpoints:

### 1. Register User

**Method:** `auth.v1.AuthService/Register`

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Response:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com"
}
```

### 2. Login User

**Method:** `auth.v1.AuthService/Login`

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Testing with grpcurl

**Install grpcurl:**
```bash
# macOS
brew install grpcurl

# Linux
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

**List available services:**
```bash
grpcurl -plaintext localhost:50051 list
```

**Register a new user:**
```bash
grpcurl -plaintext -d '{
  "email": "test@example.com",
  "password": "MySecurePass123!"
}' localhost:50051 auth.v1.AuthService/Register
```

**Login:**
```bash
grpcurl -plaintext -d '{
  "email": "test@example.com",
  "password": "MySecurePass123!"
}' localhost:50051 auth.v1.AuthService/Login
```

### Testing with Postman

1. Create a new **gRPC Request** in Postman
2. Set server URL: `localhost:50051`
3. Enable **"Use server reflection"**
4. Select method: `auth.v1.AuthService/Register` or `Login`
5. Enter JSON payload in the message body

---

## 🔧 Development

### Available Make Commands

```bash
make server          # Run the gRPC server locally
make test            # Run all tests with verbose output
make run             # Full workflow: sqlc generate → migrate → run server
make migrate-up      # Apply database migrations
make migrate-down    # Rollback database migrations
make sqlc            # Generate type-safe Go code from SQL queries
```

### Docker Commands

```bash
# Start all services
docker compose up

# Rebuild and start
docker compose up --build

# Stop all services
docker compose down

# View logs
docker compose logs -f server

# Clean up (including volumes)
docker compose down -v
```

---

## 🧪 Testing

### Run All Tests

```bash
make test
```

Or directly with Go:
```bash
go test -v ./...
```

### Run Tests with Coverage

```bash
go test -v -cover ./...
```

### Run Specific Package Tests

```bash
go test -v ./internal/core/services/
```

### Test Structure

- **Unit Tests** — Table-driven tests with parallel execution
- **Mocks** — Defined in `*_test.go` files alongside production code
- **Focus** — Business logic in `services/` layer

---

## 🔐 Environment Variables

| Variable          | Description                              | Required | Default     |
|-------------------|------------------------------------------|----------|-------------|
| `POSTGRES_HOST`   | PostgreSQL host                          | Yes      | —           |
| `POSTGRES_PORT`   | PostgreSQL port                          | Yes      | —           |
| `POSTGRES_USER`   | PostgreSQL username                      | Yes      | —           |
| `POSTGRES_PASSWORD` | PostgreSQL password                    | Yes      | —           |
| `POSTGRES_DB`     | Database name                            | Yes      | —           |
| `DB_SSLMODE`      | SSL mode (`disable`, `require`, etc.)    | No       | `disable`   |
| `JWT_SECRET`      | Secret key for JWT signing (min 32 bytes)| Yes      | —           |
| `GRPC_PORT`       | Port for gRPC server                     | No       | `50051`     |

### Generating a Secure JWT Secret

```bash
# Using OpenSSL (recommended)
openssl rand -base64 32

# Using Python
python3 -c "import secrets; print(secrets.token_urlsafe(32))"
```

Copy the output to your `.env` file:
```bash
JWT_SECRET=your-generated-secret-here
```

---

## 💡 Design Decisions

### Why Hexagonal Architecture?

- **Testability** — Business logic isolated from infrastructure
- **Flexibility** — Easy to swap adapters (e.g., switch from PostgreSQL to MongoDB)
- **Maintainability** — Clear boundaries between layers
- **Domain-Driven** — Core domain logic remains pure and focused

### Why Argon2id?

- **Memory-hard** — Resistant to GPU/ASIC attacks
- **Timing-safe** — Constant-time verification prevents timing attacks
- **Industry standard** — Recommended by OWASP for password hashing

### Why gRPC?

- **Performance** — Binary protocol with HTTP/2
- **Type safety** — Strongly typed contracts with Protocol Buffers
- **Code generation** — Auto-generated client/server code
- **Streaming support** — Built-in support for bidirectional streaming

### Timing Attack Protection

The login implementation uses a **dummy hash** to ensure constant-time response, preventing attackers from enumerating valid email addresses through timing analysis.

```go
// Always perform password verification, even when user doesn't exist
hashToVerify := s.dummyHash
if userExists {
    hashToVerify = user.PasswordHash.String()
}
ok, _ := s.hasher.Verify(ctx, password, hashToVerify)
```

---

## 📄 License

This project is licensed under the [MIT License](./LICENSE.md).

---

## 🤝 Contributing

This is a portfolio project, but feedback and suggestions are welcome! Feel free to open an issue or submit a pull request.

---

**Built with ❤️ by [Rubens Antonio Rosa](https://github.com/rubensantoniorosa2704)**
