package flags

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type ServerOptions struct {
	FlagRunAddr   string
	StoreInterval int64
	Path          string
	Restore       bool
}

const (
	defaultStoreInterval = 60
	defaultFileName      = "File_Log_Store.json"
	defaultRestore       = true
)

func ParseServerFlags() *ServerOptions {
	opt := &ServerOptions{}
	flag.StringVar(&opt.FlagRunAddr, "a", ":8080", "address and port to run server")

	flag.Int64Var(&opt.StoreInterval, "i", defaultStoreInterval, "address and port to run server")
	flag.StringVar(&opt.Path, "f", defaultFileName, "address and port to run server")
	flag.BoolVar(&opt.Restore, "r", defaultRestore, "address and port to run server")

	flag.Parse()

	if len(flag.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "Unknown flags:", flag.Args())
		os.Exit(1)
	}

	if addr := os.Getenv("ADDRESS"); addr != "" {
		opt.FlagRunAddr = addr
	}

	if interval, err := strconv.ParseInt(os.Getenv("STORE_INTERVAL"), 10, 64); err == nil {
		opt.StoreInterval = interval
	}

	if path := os.Getenv("FILE_STORAGE_PATH"); path != "" {
		opt.Path = path
	}

	if restore, err := strconv.ParseBool(os.Getenv("RESTORE")); err == nil {
		opt.Restore = restore
	}
	return opt
}
