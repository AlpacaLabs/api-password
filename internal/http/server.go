package http

import (
	"fmt"
	"net/http"

	"github.com/AlpacaLabs/password-reset/internal/configuration"
	"github.com/AlpacaLabs/password-reset/internal/service"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	config  configuration.Config
	service service.Service
}

func NewServer(config configuration.Config, service service.Service) Server {
	return Server{
		config:  config,
		service: service,
	}
}

func (s Server) Run() {
	r := mux.NewRouter()

	r.HandleFunc("/password-reset", s.SendCodeOptions).Methods(http.MethodPost)
	r.HandleFunc("/password-reset/{code:[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}}", s.VerifyCode).Methods(http.MethodGet)
	r.HandleFunc("/password-reset", s.ResetPassword).Methods(http.MethodPut)

	addr := fmt.Sprintf(":%d", s.config.HTTPPort)
	log.Infof("Listening for HTTP on %s...\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
