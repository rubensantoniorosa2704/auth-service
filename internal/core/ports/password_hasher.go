package ports

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password string, hash string) bool
}
