package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

var Options struct {
	flagRunAddr    string
	reportInterval int64
	pollInterval   int64
}

func ParseFlags() {
	flag.StringVar(&Options.flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Int64Var(&Options.reportInterval, "r", 10, "report interval in seconds")
	flag.Int64Var(&Options.pollInterval, "p", 2, "poll interval in seconds")
	flag.Parse()

	if len(flag.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "Unknown flags:", flag.Args())
		os.Exit(1)
	}

	if addr := os.Getenv("ADDRESS"); addr != "" {
		Options.flagRunAddr = addr
	}

	if interval := os.Getenv("REPORT_INTERVAL"); interval != "" {
		Options.reportInterval, _ = strconv.ParseInt(interval, 10, 64)
	}

	if interval := os.Getenv("POLL_INTERVAL"); interval != "" {
		Options.pollInterval, _ = strconv.ParseInt(interval, 10, 64)
	}
}
