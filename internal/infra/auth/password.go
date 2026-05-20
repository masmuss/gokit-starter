// Package auth provides reusable authentication primitives.
package auth

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/masmuss/gokit-starter/internal/config"
)

const defaultBcryptCost = bcrypt.DefaultCost

// PasswordHasher hashes and verifies passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

// BcryptHasher hashes passwords with bcrypt.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new bcrypt hasher.
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost <= 0 {
		cost = defaultBcryptCost
	}

	return &BcryptHasher{cost: cost}
}

// NewBcryptHasherFromConfig creates a Bcrypthasher from config.
func NewBcryptHasherFromConfig(cfg *config.Config) *BcryptHasher {
	return NewBcryptHasher(cfg.Bcrypt.Rounds)
}

// Hash returns a bcrypt hash of the password.
func (h *BcryptHasher) Hash(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

// Compare verifies a plaintext password against a bcrypt hash.
func (h *BcryptHasher) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
