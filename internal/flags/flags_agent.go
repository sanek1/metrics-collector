package flags

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Options struct {
	FlagRunAddr    string
	ReportInterval int64
	PollInterval   int64
}

const (
	defaultReportInterval = 10
	defaultPollInterval   = 2
)

func ParseFlags() *Options {
	flagRunAddr := ""
	reportInterval := int64(defaultReportInterval)
	pollInterval := int64(defaultPollInterval)

	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Int64Var(&reportInterval, "r", defaultReportInterval, "report interval in seconds")
	flag.Int64Var(&pollInterval, "p", defaultPollInterval, "poll interval in seconds")
	flag.Parse()

	if len(flag.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "Unknown flags:", flag.Args())
		os.Exit(1)
	}

	if addr := os.Getenv("ADDRESS"); addr != "" {
		flagRunAddr = addr
	}

	if interval := os.Getenv("REPORT_INTERVAL"); interval != "" {
		reportInterval, _ = strconv.ParseInt(interval, 10, 64)
	}

	if interval := os.Getenv("POLL_INTERVAL"); interval != "" {
		pollInterval, _ = strconv.ParseInt(interval, 10, 64)
	}
	return &Options{
		FlagRunAddr:    flagRunAddr,
		ReportInterval: reportInterval,
		PollInterval:   pollInterval,
	}
}
