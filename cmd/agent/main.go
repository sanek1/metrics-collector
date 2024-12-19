package main

import (
	"log"

	"github.com/sanek1/metrics-collector/internal/app"
	"github.com/sanek1/metrics-collector/internal/flags"
)

func main() {
	opt := flags.ParseFlags()
	if err := app.Run(opt); err != nil {
		log.Fatal(err)
	}
}
