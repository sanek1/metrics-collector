// package flags
package flags

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Options struct {
	FlagRunAddr    string
	CryptoKey      string
	ReportInterval int64
	PollInterval   int64
	RateLimit      int64
}

const (
	defaultReportInterval = 10
	defaultPollInterval   = 2
	defaultRateLimit      = 2
)

func ParseFlags() *Options {
	opt := &Options{}

	flag.StringVar(&opt.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&opt.CryptoKey, "k", "", "key to encrypt/decrypt metrics")
	flag.Int64Var(&opt.ReportInterval, "r", defaultReportInterval, "report interval in seconds")
	flag.Int64Var(&opt.PollInterval, "p", defaultPollInterval, "poll interval in seconds")
	flag.Int64Var(&opt.RateLimit, "l", defaultRateLimit, "outgoing requests to the server ")
	flag.Parse()

	if len(flag.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "Unknown flags:", flag.Args())
		os.Exit(1)
	}

	if addr := os.Getenv("ADDRESS"); addr != "" {
		opt.FlagRunAddr = addr
	}

	if interval := os.Getenv("REPORT_INTERVAL"); interval != "" {
		opt.ReportInterval, _ = strconv.ParseInt(interval, 10, 64)
	}

	if interval := os.Getenv("POLL_INTERVAL"); interval != "" {
		opt.PollInterval, _ = strconv.ParseInt(interval, 10, 64)
	}

	if key := os.Getenv("KEY"); key != "" {
		opt.CryptoKey = key
	}

	if limit := os.Getenv("RATE_LIMIT"); limit != "" {
		opt.RateLimit, _ = strconv.ParseInt(limit, 10, 64)
	}

	return opt
}
