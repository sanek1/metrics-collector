package main

import (
	"log"

	app "github.com/sanek1/metrics-collector/internal/app/server"
	flags "github.com/sanek1/metrics-collector/internal/flags/server"
)

func main() {
	opt := flags.ParseServerFlags()
	if err := app.New(opt).Run(); err != nil {
		log.Fatal(err)
	}
}
