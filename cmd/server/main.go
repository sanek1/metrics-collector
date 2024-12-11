package main

import (
	"log"

	"github.com/sanek1/metrics-collector/internal/app"
)

func main() {
	ParseFlags()

	application := app.New(Options.flagRunAddr)
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
