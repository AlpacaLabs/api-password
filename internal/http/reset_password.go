package http

import (
	"net/http"

	authV1 "github.com/AlpacaLabs/protorepo-auth-go/alpacalabs/auth/v1"
	"github.com/golang/protobuf/jsonpb"
)

func (s Server) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// De-serialize request body
	var requestBody authV1.ResetPasswordRequest
	if err := jsonpb.Unmarshal(r.Body, &requestBody); err != nil {
		// TODO return response
		return
	}
	defer r.Body.Close()

	// Call the service function
	err := s.service.ResetPassword(ctx, requestBody)

	// Return any errors
	if err != nil {
		// TODO return response
	}

	w.Write(nil)
}
