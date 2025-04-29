package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	_ "net/http/pprof"

	app "github.com/sanek1/metrics-collector/internal/app/server"
	flags "github.com/sanek1/metrics-collector/internal/flags/server"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

const (
	readTimeout   = 5 * time.Second
	writeTimeout  = 60 * time.Second
	idleTimeout   = 15 * time.Second
	versionString = "N/A"
)

// @title Metrics Collector API
// @version 1.0
// @description API для сбора и обработки метрик
// @host localhost:8080
// @BasePath /
func main() {
	printBuildInfo()

	code := run()
	if code != 0 {
		if err := os.Setenv("EXIT_CODE", fmt.Sprintf("%d", code)); err != nil {
			fmt.Fprintf(os.Stderr, "not set EXIT_CODE: %v\n", err)
		}
		return
	}
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", defaultIfEmpty(buildVersion, versionString))
	fmt.Printf("Build date: %s\n", defaultIfEmpty(buildDate, versionString))
	fmt.Printf("Build commit: %s\n", defaultIfEmpty(buildCommit, versionString))
}

func defaultIfEmpty(value, versionString string) string {
	if value == "" {
		return versionString
	}
	return value
}

func run() int {
	errCh := make(chan error, 2)

	go func() {
		server := &http.Server{
			Addr:         "localhost:6060",
			Handler:      nil,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		}
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("error running profiler: %w", err)
		}
	}()

	opt := flags.ParseServerFlags()
	go func() {
		if err := app.New(opt, opt.UseDatabase).Run(); err != nil {
			errCh <- fmt.Errorf("error running server: %w", err)
		}
	}()

	err := <-errCh
	fmt.Fprintf(os.Stderr, "critical error: %v\n", err)
	return 1
}
