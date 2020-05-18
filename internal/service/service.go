package service

import (
	"time"

	"github.com/AlpacaLabs/api-password/internal/configuration"
	"github.com/AlpacaLabs/api-password/internal/db"
	"github.com/AlpacaLabs/api-password/internal/db/entities"
	"google.golang.org/grpc"
)

type Service struct {
	config             configuration.Config
	dbClient           db.Client
	accountConn        *grpc.ClientConn
	argonConfiguration entities.ArgonConfiguration
}

func NewService(config configuration.Config, dbClient db.Client, accountConn *grpc.ClientConn) Service {
	argonConfiguration := GetHashConfiguration(time.Millisecond * 500)
	return Service{
		config:             config,
		argonConfiguration: argonConfiguration,
		dbClient:           dbClient,
		accountConn:        accountConn,
	}
}
