package service

import (
	"runtime"
	"time"

	"github.com/AlpacaLabs/api-password/internal/db/entities"
	"golang.org/x/crypto/argon2"

	log "github.com/sirupsen/logrus"
)

// GetHashConfiguration calibrates the Argon password hashing configuration.
// The goal is to set memory, iteration count, parallelism, salt length, and
// key length such that generating a hash takes approximately 500ms, the
// amount of time a login should take.
// The only input to this function is how long a hash should take.
func GetHashConfiguration(targetHashTime time.Duration) entities.ArgonConfiguration {
	c := entities.ArgonConfiguration{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 4,
		KeyLength:   32,
		SaltLength:  32,
	}

	salt, err := generateSalt(c.SaltLength)
	if err != nil {
		log.Fatal(err)
	}

	maxThreads := runtime.NumCPU()
	log.Infof("Num CPU = %d", maxThreads)

	log.Infof("Calibrating password iteration count, starting at %d iterations...", c.Iterations)
	for {
		start := time.Now()

		passwordText := "MyPassword123!"
		argon2.IDKey([]byte(passwordText), salt, c.Iterations, c.Memory, c.Parallelism, c.KeyLength)

		elapsed := time.Since(start)
		if elapsed > targetHashTime {
			log.Printf("Took %s to do %d iterations\n", elapsed, c.Iterations)
			break
		}

		percentage := elapsed.Seconds() / targetHashTime.Seconds()
		if percentage < 0.2 {
			log.Println("Less than 20% of the way there...")
			c.Iterations = c.Iterations * 4
		} else if percentage < 0.3 {
			log.Println("Less than 30% of the way there...")
			c.Iterations = c.Iterations * 3
		} else if percentage < 0.4 {
			log.Println("Less than 40% of the way there...")
			c.Iterations = c.Iterations * 2
		} else if percentage < 0.5 {
			log.Println("Less than 50% of the way there...")
			c.Iterations = uint32(float64(c.Iterations) * 1.75)
		} else if percentage < 0.6 {
			log.Println("Less than 60% of the way there...")
			c.Iterations = uint32(float64(c.Iterations) * 1.55)
		} else if percentage < 0.7 {
			log.Println("Less than 70% of the way there...")
			c.Iterations = uint32(float64(c.Iterations) * 1.35)
		} else if percentage < 0.8 {
			log.Println("Less than 80% of the way there...")
			c.Iterations = uint32(float64(c.Iterations) * 1.20)
		} else if percentage < 0.9 {
			log.Println("Less than 90% of the way there...")
			c.Iterations = uint32(float64(c.Iterations) * 1.07)
		} else if percentage < 0.95 {
			log.Println("Less than 95% of the way there...")
			c.Iterations = uint32(float64(c.Iterations) * 1.04)
		} else {
			log.Printf("We're close enough... Took %s to do %d iterations\n", elapsed, c.Iterations)
			break
		}
	}
	return c
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	log.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	log.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	log.Printf("\tSys = %v MiB", bToMb(m.Sys))
	log.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
