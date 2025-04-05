package main

import (
	"fmt"
	"os"

	app "github.com/sanek1/metrics-collector/internal/app/agent"
	flags "github.com/sanek1/metrics-collector/internal/flags/agent"
)

func main() {
	code := run()
	if code != 0 {
		if err := os.Setenv("EXIT_CODE", fmt.Sprintf("%d", code)); err != nil {
			fmt.Fprintf(os.Stderr, "not set EXIT_CODE: %v\n", err)
		}
		return
	}
}

func run() int {
	opt := flags.ParseFlags()
	if err := app.New(opt).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}
