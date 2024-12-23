package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

var Options struct {
	flagRunAddr   string
	storeInterval int64
	path          string
	restore       bool
}

const (
	defaultStoreInterval = 60
	defaultFileName      = "File_Log_Store.json"
	defaultRestore       = true
)

func ParseFlags() {
	flag.StringVar(&Options.flagRunAddr, "a", ":8080", "address and port to run server")

	flag.Int64Var(&Options.storeInterval, "i", defaultStoreInterval, "address and port to run server")
	flag.StringVar(&Options.path, "f", defaultFileName, "address and port to run server")
	flag.BoolVar(&Options.restore, "r", defaultRestore, "address and port to run server")

	flag.Parse()

	if len(flag.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "Unknown flags:", flag.Args())
		os.Exit(1)
	}

	if addr := os.Getenv("ADDRESS"); addr != "" {
		Options.flagRunAddr = addr
	}

	if interval, err := strconv.ParseInt(os.Getenv("STORE_INTERVAL"), 10, 64); err == nil {
		Options.storeInterval = interval
	}

	if path := os.Getenv("FILE_STORAGE_PATH"); path != "" {
		Options.path = path
	}

	if restore, err := strconv.ParseBool(os.Getenv("RESTORE")); err == nil {
		Options.restore = restore
	}
}
