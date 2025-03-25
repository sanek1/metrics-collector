package main

import (
	"log"
	"net/http"
	"time"

	app "github.com/sanek1/metrics-collector/internal/app/server"
	flags "github.com/sanek1/metrics-collector/internal/flags/server"
)

const (
	readTimeout  = 5 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 15 * time.Second
)

// @title Metrics Collector API
// @version 1.0
// @description API для сбора и обработки метрик
// @host localhost:8080
// @BasePath /
func main() {
	go func() {
		server := &http.Server{
			Addr:         "localhost:6060",
			Handler:      nil,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		}
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	opt := flags.ParseServerFlags()
	if err := app.New(opt, opt.UseDatabase).Run(); err != nil {
		log.Fatal(err)
	}
}
