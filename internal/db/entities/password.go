package entities

import (
	"time"
)

// Password is a representation of a user's password.
type Password struct {
	ID        string
	CreatedAt time.Time
	Salt      []byte
	Hash      []byte
	AccountID string
	ArgonConfiguration
}

type ArgonConfiguration struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}
