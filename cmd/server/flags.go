package main

import (
	"flag"
	"fmt"
	"os"
)

var Options struct {
	flagRunAddr string
}

func ParseFlags() {
	flag.StringVar(&Options.flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Parse()

	if len(flag.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "Unknown flags:", flag.Args())
		os.Exit(1)
	}

	if addr := os.Getenv("ADDRESS"); addr != "" {
		Options.flagRunAddr = addr
	}
}
