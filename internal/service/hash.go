package service

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"time"

	"github.com/AlpacaLabs/api-password/internal/db/entities"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/pbkdf2"
)

func matchesHash(passwordText string, password *entities.Password) bool {
	hash := generateHash(passwordText, password.IterationCount, password.Salt)
	return hex.EncodeToString(password.PasswordHash) == hex.EncodeToString(hash)
}

func generateHash(passwordText string, iterationCount int, salt []byte) []byte {
	start := time.Now()
	// TODO use switch statement on scheme string
	hash := pbkdf2.Key([]byte(passwordText), salt, iterationCount, 32, sha1.New)
	log.Printf("Hashing %d iterations took: %s", iterationCount, time.Since(start))
	return hash
}

func generateSalt(byteLength int) ([]byte, error) {
	salt := make([]byte, byteLength)
	_, err := io.ReadFull(rand.Reader, salt)
	return salt, err
}
