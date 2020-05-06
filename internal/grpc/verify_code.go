package grpc

import (
	"context"

	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
)

func (s Server) VerifyCode(ctx context.Context, request *passwordV1.VerifyCodeRequest) (*passwordV1.VerifyCodeResponse, error) {
	return s.service.VerifyCode(ctx, *request)
}
