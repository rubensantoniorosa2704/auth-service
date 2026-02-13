package domain

// Password is a value object for a raw (unhashed) password.
// It enforces minimum strength requirements before hashing.
type Password struct {
	value string
}

// NewPassword validates a raw password and returns a Password value object.
func NewPassword(raw string) (Password, error) {
	if len(raw) < 8 {
		return Password{}, ErrWeakPassword
	}

	return Password{value: raw}, nil
}

// String returns the raw password. Only used during the hashing step.
func (p Password) String() string {
	return p.value
}

// PasswordHash is a value object for a hashed password stored in the database.
type PasswordHash struct {
	value string
}

// NewPasswordHash wraps a pre-computed hash string.
func NewPasswordHash(hash string) (PasswordHash, error) {
	if hash == "" {
		return PasswordHash{}, ErrWeakPassword
	}

	return PasswordHash{value: hash}, nil
}

// String returns the hashed password string.
func (p PasswordHash) String() string {
	return p.value
}
