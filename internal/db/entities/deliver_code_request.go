package entities

import (
	eventV1 "github.com/AlpacaLabs/protorepo-event-go/alpacalabs/event/v1"
	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
	"github.com/rs/xid"
)

type DeliverCodeRequest struct {
	eventV1.EventInfo
	eventV1.TraceInfo
	Sent bool
	passwordV1.DeliverCodeRequest
}

func NewDeliverCodeRequest(traceInfo eventV1.TraceInfo, payload passwordV1.DeliverCodeRequest) DeliverCodeRequest {
	return DeliverCodeRequest{
		EventInfo: eventV1.EventInfo{
			EventId: xid.New().String(),
		},
		TraceInfo:          traceInfo,
		Sent:               false,
		DeliverCodeRequest: payload,
	}
}
