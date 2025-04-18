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
	version := buildVersion
	if version == "" {
		version = "N/A"
	}

	date := buildDate
	if date == "" {
		date = "N/A"
	}

	commit := buildCommit
	if commit == "" {
		commit = "N/A"
	}

	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)
}

func run() int {
	opt := flags.ParseFlags()
	if err := app.New(opt).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}
