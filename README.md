# Auth Service

Auth Service is a simple and professional authentication microservice built with Go.
It was created as a portfolio project to demonstrate clean architecture, gRPC communication,
and good backend engineering practices.

## ✨ Features

- User registration
- User authentication (login)
- Password hashing using Argon2id
- JWT token generation
- gRPC API using Protocol Buffers
- PostgreSQL as database
- Clean Architecture (Hexagonal Architecture)
- Unit tests with mocks
- Environment configuration with Viper
- Docker and DevContainer support

## 🛠 Tech Stack

- Go 1.22+
- gRPC + Protocol Buffers
- PostgreSQL
- Docker & Docker Compose
- Viper
- JWT
- Argon2id

## 📂 Project Structure

```
cmd/                # Application entrypoints
internal/
  core/
    domain/         # Domain entities and business rules
    ports/          # Interfaces (ports)
    services/       # Use cases / application services
  adapters/         # Infrastructure adapters (database, grpc, etc.)
```

## 🚀 Getting Started

### Prerequisites

- Go 1.22 or higher
- Docker and Docker Compose

### Running the project

```
docker compose up
```

Or run locally:

```
go run ./cmd/auth
```

## 🧪 Running Tests

```
go test ./internal/...
```

## 📜 License

This project is licensed under the [MIT License](./LICENSE.md).
