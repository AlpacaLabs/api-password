package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s Server) VerifyCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	code := vars["code"]

	valid, err := s.service.VerifyCode(ctx, code)
	if err != nil || !valid {
		// TODO return error response
		return
	}

	if valid {
		// TODO
	}
}
