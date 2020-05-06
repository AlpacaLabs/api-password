package grpc

import (
	"context"

	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
)

func (s Server) GetDeliveryOptions(ctx context.Context, request *passwordV1.GetDeliveryOptionsRequest) (*passwordV1.GetDeliveryOptionsResponse, error) {
	return s.service.GetDeliveryOptions(ctx, *request)
}
