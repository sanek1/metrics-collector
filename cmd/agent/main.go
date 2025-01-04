package main

import (
	"log"

	app "github.com/sanek1/metrics-collector/internal/app/agent"
	flags "github.com/sanek1/metrics-collector/internal/flags/agent"
)

func main() {
	opt := flags.ParseFlags()
	if err := app.New(opt).Run(); err != nil {
		log.Fatal(err)
	}
}
