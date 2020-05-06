package grpc

import (
	"context"

	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
)

func (s Server) CheckPassword(ctx context.Context, request *passwordV1.CheckPasswordRequest) (*passwordV1.CheckPasswordResponse, error) {
	return s.service.CheckPassword(ctx, *request)
}
