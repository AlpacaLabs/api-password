package service

import (
	"time"

	log "github.com/sirupsen/logrus"
)

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
