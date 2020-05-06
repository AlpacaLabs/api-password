package main

import (
	"sync"

	"github.com/AlpacaLabs/api-password/internal/app"
	"github.com/AlpacaLabs/api-password/internal/configuration"
)

func main() {
	c := configuration.LoadConfig()
	a := app.NewApp(c)

	var wg sync.WaitGroup

	wg.Add(1)
	go a.Run()

	wg.Wait()
}
