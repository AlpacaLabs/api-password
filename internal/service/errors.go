package service

import "errors"

var (
	ErrCodeExpired = errors.New("code has expired")
)
