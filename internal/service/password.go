package service

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"time"

	"github.com/AlpacaLabs/api-password/internal/db/entities"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/pbkdf2"
)

// Update the password for the given account ID.
func (s *Service) UpdatePassword(ctx context.Context, accountID string) {}

// CalibrateIterationCount finds the given
func CalibrateIterationCount(hashTime time.Duration) int {
	salt, err := generateSalt(32)
	if err != nil {
		log.Fatal(err)
	}

	iterationCount := 10000
	log.Printf("Calibrating password iteration count, starting at %d iterations...\n", iterationCount)
	for {
		start := time.Now()
		generateHash("MyPassword123!", iterationCount, salt)
		elapsed := time.Since(start)
		if elapsed > hashTime {
			log.Printf("Took %s to do %d iterations\n", elapsed, iterationCount)
			break
		}

		percentage := elapsed.Seconds() / hashTime.Seconds()
		if percentage < 0.2 {
			log.Println("Less than 20% of the way there...")
			iterationCount = iterationCount * 4
		} else if percentage < 0.3 {
			log.Println("Less than 30% of the way there...")
			iterationCount = iterationCount * 3
		} else if percentage < 0.4 {
			log.Println("Less than 40% of the way there...")
			iterationCount = iterationCount * 2
		} else if percentage < 0.5 {
			log.Println("Less than 50% of the way there...")
			iterationCount = int(float64(iterationCount) * 1.75)
		} else if percentage < 0.6 {
			log.Println("Less than 60% of the way there...")
			iterationCount = int(float64(iterationCount) * 1.55)
		} else if percentage < 0.7 {
			log.Println("Less than 70% of the way there...")
			iterationCount = int(float64(iterationCount) * 1.35)
		} else if percentage < 0.8 {
			log.Println("Less than 80% of the way there...")
			iterationCount = int(float64(iterationCount) * 1.20)
		} else if percentage < 0.9 {
			log.Println("Less than 90% of the way there...")
			iterationCount = int(float64(iterationCount) * 1.07)
		} else if percentage < 0.95 {
			log.Println("Less than 95% of the way there...")
			iterationCount = int(float64(iterationCount) * 1.04)
		} else {
			log.Printf("We're close enough... Took %s to do %d iterations\n", elapsed, iterationCount)
			break
		}
	}
	return iterationCount
}

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
