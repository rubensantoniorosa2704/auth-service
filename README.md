# Auth Service

A production-grade authentication microservice built with Go, designed to demonstrate clean architecture,
gRPC communication, and industry-standard backend engineering practices.

## Features

- **User Registration** — with email and password validation (domain value objects)
- **User Authentication** — login with JWT token generation
- **Argon2id Password Hashing** — memory-hard, timing-safe password storage
- **gRPC API** — strongly typed contracts with Protocol Buffers
- **PostgreSQL** — persistent storage with [sqlc](https://sqlc.dev) type-safe queries
- **Hexagonal Architecture** — strict separation of domain, ports, and adapters
- **Structured Logging** — using Go's standard `log/slog` (JSON output)
- **Context Propagation** — `context.Context` flows from gRPC handler through every layer
- **Graceful Shutdown** — signal-aware server lifecycle management
- **Table-Driven Tests** — comprehensive unit tests with parallel execution

## Tech Stack

| Category       | Technology                                |
|----------------|-------------------------------------------|
| Language       | Go 1.22+                                  |
| Transport      | gRPC + Protocol Buffers                   |
| Database       | PostgreSQL (pgx v5 + sqlc)                |
| Authentication | JWT (golang-jwt/v5)                       |
| Hashing        | Argon2id (golang.org/x/crypto)            |
| Logging        | log/slog (standard library)               |
| Infrastructure | Docker, Docker Compose, DevContainers     |

## Project Structure

```
cmd/server/             Application entrypoint
internal/
  core/
    domain/             Value objects, entities, domain errors
    ports/              Interfaces (driven + driving ports)
    services/           Use cases / application services
  adapters/
    db/                 PostgreSQL repository (pgx + sqlc)
    encryption/         Argon2id password hasher
    handlers/grpc/      gRPC transport handler
    tokens/             JWT token service
proto/auth/v1/          Protocol Buffer definitions
```

## Getting Started

### Prerequisites

- Go 1.22 or higher
- Docker and Docker Compose

### Running with Docker

```bash
docker compose up
```

### Running locally

```bash
make server
```

### Running Tests

```bash
make test
```

## Environment Variables

| Variable        | Description                  | Default     |
|-----------------|------------------------------|-------------|
| `DB_HOST`       | PostgreSQL host              | `localhost` |
| `DB_PORT`       | PostgreSQL port              | `5432`      |
| `DB_USER`       | PostgreSQL user              | `postgres`  |
| `DB_PASSWORD`   | PostgreSQL password          | —           |
| `DB_NAME`       | Database name                | `auth`      |
| `DB_SSLMODE`    | SSL mode                     | `disable`   |
| `JWT_SECRET`    | Secret key for JWT signing (min 32 bytes) | —           |
| `GRPC_PORT`     | Port for gRPC server         | `50051`     |

### Generating a Secure JWT Secret

The JWT secret must be at least 32 bytes long for security. You can generate a secure random secret using:

```bash
# Using OpenSSL (recommended)
openssl rand -base64 32

# Using Go
go run -e 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"); func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(base64.StdEncoding.EncodeToString(b)) }'

# Using Python
python3 -c "import secrets; print(secrets.token_urlsafe(32))"
```

Copy the generated value to your `.env` file:

```bash
JWT_SECRET=your-generated-secret-here
```

## License

This project is licensed under the [MIT License](./LICENSE.md).
