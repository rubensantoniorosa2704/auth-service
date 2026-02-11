-include .env

DB_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOSTNAME):5432/$(POSTGRES_DB)?sslmode=disable

.PHONY: migrate-up migrate-down sqlc server test

migrate-up:
	migrate -path database/migrations/ -database "$(DB_URL)" up

migrate-down:
	migrate -path database/migrations/ -database "$(DB_URL)" down

sqlc:
	sqlc generate

server:
	go run cmd/server/main.go

test:
	go test -v ./...

run: sqlc migrate-up server