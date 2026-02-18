package domain

import (
	"regexp"
	"strings"
)

// Email validation constants based on RFC 5321 and RFC 5322
const (
	// MaxEmailLength is the maximum total length of an email address (RFC 5321)
	MaxEmailLength = 254
	// MaxLocalPartLength is the maximum length of the local part (before @)
	MaxLocalPartLength = 64
	// MaxDomainLength is the maximum length of the domain part (after @)
	MaxDomainLength = 255
)

// emailRegex is a robust email validation pattern that covers most valid email formats.
// It validates:
// - Local part: alphanumeric, dots, hyphens, underscores, plus signs, percent
// - Domain: alphanumeric, dots, hyphens
// - TLD: at least 2 characters, letters only
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._%+\-]*[a-zA-Z0-9]@[a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?)*\.[a-zA-Z]{2,}$`)

// Email is a value object that guarantees a structurally valid email address.
type Email struct {
	value string
}

// NewEmail validates and creates an Email value object.
// It performs comprehensive validation including:
// - Length limits (RFC 5321)
// - Format validation (RFC 5322)
// - Normalization (lowercase, trim)
// - Special character rules
func NewEmail(value string) (Email, error) {
	// Normalize: trim whitespace and convert to lowercase
	value = strings.TrimSpace(strings.ToLower(value))

	// Check if empty
	if value == "" {
		return Email{}, ErrInvalidEmail
	}

	// Check maximum length (RFC 5321: 254 characters)
	if len(value) > MaxEmailLength {
		return Email{}, ErrInvalidEmail
	}

	// Split into local and domain parts
	parts := strings.Split(value, "@")
	if len(parts) != 2 {
		return Email{}, ErrInvalidEmail
	}

	localPart := parts[0]
	domainPart := parts[1]

	// Validate local part length (RFC 5321: 64 characters max)
	if len(localPart) == 0 || len(localPart) > MaxLocalPartLength {
		return Email{}, ErrInvalidEmail
	}

	// Validate domain part length (RFC 5321: 255 characters max)
	if len(domainPart) == 0 || len(domainPart) > MaxDomainLength {
		return Email{}, ErrInvalidEmail
	}

	// Check for consecutive dots (not allowed)
	if strings.Contains(value, "..") {
		return Email{}, ErrInvalidEmail
	}

	// Check if local part starts or ends with a dot
	if strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") {
		return Email{}, ErrInvalidEmail
	}

	// Check if domain starts or ends with a dot or hyphen
	if strings.HasPrefix(domainPart, ".") || strings.HasSuffix(domainPart, ".") ||
		strings.HasPrefix(domainPart, "-") || strings.HasSuffix(domainPart, "-") {
		return Email{}, ErrInvalidEmail
	}

	// Validate against regex pattern
	if !emailRegex.MatchString(value) {
		return Email{}, ErrInvalidEmail
	}

	return Email{value: value}, nil
}

// String returns the email address as a plain string.
func (e Email) String() string {
	return e.value
}
