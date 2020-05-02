package app

import (
	"sync"

	"github.com/AlpacaLabs/password-reset/internal/config"
	"github.com/AlpacaLabs/password-reset/internal/db"
	"github.com/AlpacaLabs/password-reset/internal/http"
	"github.com/AlpacaLabs/password-reset/internal/services"
	_ "github.com/lib/pq"
)

type App struct {
	config config.Config
}

func NewApp(c config.Config) App {
	return App{
		config: c,
	}
}

func (a App) Run() {
	dbConn := db.Connect(a.config.DBUser, a.config.DBPass, a.config.DBHost, a.config.DBName)
	dbClient := db.NewClient(dbConn)
	svc := services.NewService(a.config, dbClient)

	var wg sync.WaitGroup

	wg.Add(1)
	httpServer := http.NewServer(a.config, svc)
	httpServer.Run()

	wg.Wait()
}
