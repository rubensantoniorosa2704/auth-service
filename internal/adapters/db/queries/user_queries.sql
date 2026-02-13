-- name: CreateUser :one
INSERT INTO users (
    id, email, password_hash, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, created_at, updated_at
FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT id, email, password_hash, created_at, updated_at
FROM users
WHERE id = $1 LIMIT 1;