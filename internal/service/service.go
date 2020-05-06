package service

import (
	"github.com/AlpacaLabs/api-password/internal/configuration"
	"github.com/AlpacaLabs/api-password/internal/db"
	"google.golang.org/grpc"
)

type Service struct {
	config   configuration.Config
	dbClient db.Client
	authConn *grpc.ClientConn
}

func NewService(config configuration.Config, dbClient db.Client) Service {
	return Service{
		config:   config,
		dbClient: dbClient,
	}
}
