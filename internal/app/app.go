package app

import (
	"sync"

	"github.com/AlpacaLabs/api-password/internal/async"

	"google.golang.org/grpc"

	"github.com/AlpacaLabs/api-password/internal/configuration"
	"github.com/AlpacaLabs/api-password/internal/db"
	"github.com/AlpacaLabs/api-password/internal/http"
	"github.com/AlpacaLabs/api-password/internal/service"
	"github.com/sirupsen/logrus"
)

type App struct {
	config configuration.Config
}

func NewApp(c configuration.Config) App {
	return App{
		config: c,
	}
}

func (a App) Run() {
	config := a.config

	// Connect to the database
	dbConn, err := db.Connect(config.SQLConfig)
	if err != nil {
		logrus.Fatalf("failed to dial account service: %v", err)
	}
	dbClient := db.NewClient(dbConn)

	// Connect to the Account service
	accountConn, err := grpc.Dial(config.AccountGRPCAddress)
	if err != nil {
		logrus.Fatalf("failed to dial account service: %v", err)
	}

	// Create our service layer
	svc := service.NewService(config, dbClient, accountConn)

	var wg sync.WaitGroup

	wg.Add(1)
	httpServer := http.NewServer(config, svc)
	go httpServer.Run()

	wg.Add(1)
	go async.RelayMessagesForSendEmail(config, dbClient)

	wg.Add(1)
	go async.RelayMessagesForSendSms(config, dbClient)

	wg.Wait()
}
