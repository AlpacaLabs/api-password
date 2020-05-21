package grpc

import (
	"context"

	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
)

func (s Server) DeliverCode(ctx context.Context, request *passwordV1.DeliverCodeRequest) (*passwordV1.DeliverCodeResponse, error) {
	return s.service.DeliverCode(ctx, request)
}
