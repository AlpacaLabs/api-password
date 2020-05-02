package main

import (
	"sync"

	"github.com/AlpacaLabs/password-reset/internal/app"
	"github.com/AlpacaLabs/password-reset/internal/config"
)

func main() {
	c := config.LoadConfig()
	a := app.NewApp(c)

	var wg sync.WaitGroup

	wg.Add(1)
	go a.Run()

	wg.Wait()
}
