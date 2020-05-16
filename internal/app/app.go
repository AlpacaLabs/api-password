package app

import (
	"sync"

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
	dbConn, err := db.Connect(a.config.SQLConfig)
	if err != nil {
		logrus.Fatalf("failed to dial account service: %v", err)
	}
	dbClient := db.NewClient(dbConn)

	accountConn, err := grpc.Dial(a.config.AccountGRPCAddress)
	if err != nil {
		logrus.Fatalf("failed to dial account service: %v", err)
	}
	svc := service.NewService(a.config, dbClient, accountConn)

	var wg sync.WaitGroup

	wg.Add(1)
	httpServer := http.NewServer(a.config, svc)
	go httpServer.Run()

	wg.Wait()
}
