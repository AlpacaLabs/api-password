package service

import (
	"crypto/rand"
	"encoding/hex"
	"io"

	"github.com/AlpacaLabs/api-password/internal/db/entities"
	"golang.org/x/crypto/argon2"
)

func matchesHash(passwordText string, password entities.Password) bool {
	hash := generateHash(passwordText, password.Salt, password.ArgonConfiguration)
	return hex.EncodeToString(password.Hash) == hex.EncodeToString(hash)
}

func generateHash(passwordText string, salt []byte, c entities.ArgonConfiguration) []byte {
	return argon2.IDKey([]byte(passwordText), salt, c.Iterations, c.Memory, c.Parallelism, c.KeyLength)
}

// generateSalt generates a cryptographically secure random salt.
func generateSalt(byteLength uint32) ([]byte, error) {
	salt := make([]byte, byteLength)
	_, err := io.ReadFull(rand.Reader, salt)
	return salt, err
}
