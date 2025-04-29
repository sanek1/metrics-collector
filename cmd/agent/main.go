package main

import (
	"fmt"
	"os"

	app "github.com/sanek1/metrics-collector/internal/app/agent"
	flags "github.com/sanek1/metrics-collector/internal/flags/agent"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

const versionString = "N/A"

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

func defaultIfEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func run() int {
	opt := flags.ParseFlags()
	if err := app.New(opt).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}
