package tokens

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService implements ports.TokenService using JSON Web Tokens.
type JWTService struct {
	secretKey          []byte
	issuer             string
	expirationDuration time.Duration
}

// CustomClaims extends the standard JWT claims with a user identifier.
type CustomClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// NewJWTService returns a new JWTService with the given signing secret, issuer, and expiration hours.
// It validates that the secret meets minimum security requirements and that expirationHours is positive.
func NewJWTService(secret string, issuer string, expirationHours int) (*JWTService, error) {
	if err := validateSecret(secret); err != nil {
		return nil, fmt.Errorf("invalid JWT secret: %w", err)
	}

	if expirationHours <= 0 {
		return nil, fmt.Errorf("expiration hours must be positive, got %d", expirationHours)
	}

	return &JWTService{
		secretKey:          []byte(secret),
		issuer:             issuer,
		expirationDuration: time.Duration(expirationHours) * time.Hour,
	}, nil
}

// validateSecret ensures the JWT secret meets minimum security requirements.
// For HMAC-SHA256 (HS256), NIST recommends a key length equal to the hash output (256 bits = 32 bytes).
func validateSecret(secret string) error {
	const minSecretLength = 32

	if secret == "" {
		return errors.New("secret cannot be empty")
	}

	if len(secret) < minSecretLength {
		return fmt.Errorf("secret must be at least %d bytes long, got %d", minSecretLength, len(secret))
	}

	return nil
}

// Generate creates a signed JWT token for the given user ID.
func (s *JWTService) Generate(_ context.Context, userID string) (string, error) {
	claims := &CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expirationDuration)),
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return signed, nil
}

// Validate parses and verifies a JWT token, returning the embedded user ID.
func (s *JWTService) Validate(_ context.Context, tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return "", fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token")
	}

	return claims.UserID, nil
}
