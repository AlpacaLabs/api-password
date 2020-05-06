package service

import (
	"encoding/hex"
	"testing"

	"github.com/AlpacaLabs/api-password/internal/db/entities"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHash(t *testing.T) {
	Convey("Given a salt, iteration count, and password", t, func() {
		salt, err := generateSalt(32)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Salt is", hex.EncodeToString(salt))
		hash := generateHash("potato-tomato-cherry-gun", 1, salt)
		log.Println("Hash is", hex.EncodeToString(hash))
		//salt, _ = hex.DecodeString("99d1745e8fe4f96130c638d733de489c4eb9fad026bb2540a414c113e5a1fe9b")

		tests := []struct {
			password       string
			salt           []byte
			iterationCount int
		}{
			{password: "MyPassword123!", salt: salt, iterationCount: 1000},
			{password: "MyPassword123!", salt: salt, iterationCount: 2000},
			{password: "MyPassword123!", salt: salt, iterationCount: 4000},
			{password: "MyPassword123!", salt: salt, iterationCount: 6000},
			{password: "MyPassword123!", salt: salt, iterationCount: 8000},
			{password: "MyPassword123!", salt: salt, iterationCount: 10000},
			{password: "MyPassword123!", salt: salt, iterationCount: 20000},
			{password: "MyPassword123!", salt: salt, iterationCount: 40000},
			{password: "MyPassword123!", salt: salt, iterationCount: 80000},
			{password: "MyPassword123!", salt: salt, iterationCount: 160000},
			{password: "MyPassword123!", salt: salt, iterationCount: 200000},
		}

		Convey("Hashes should match", func() {
			for _, test := range tests {
				hash := generateHash(test.password, test.iterationCount, test.salt)
				match := matchesHash(test.password, &entities.Password{IterationCount: test.iterationCount, Salt: test.salt, PasswordHash: hash})
				So(match, ShouldBeTrue)
			}
		})
	})
}
