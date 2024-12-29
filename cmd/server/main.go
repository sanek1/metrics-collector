package main

import (
	"log"

	"github.com/sanek1/metrics-collector/internal/app"
	"github.com/sanek1/metrics-collector/internal/flags"
)

func main() {
	opt := flags.ParseServerFlags()

	application := app.New(opt.FlagRunAddr, opt.StoreInterval, opt.Path, opt.Restore)
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
