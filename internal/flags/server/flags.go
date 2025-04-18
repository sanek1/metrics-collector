// Package flags
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
	DBPath        string
	UseDatabase   bool
	CryptoKey     string
}

type DBSettings struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

const (
	defaultStoreInterval = 60
	defaultFileName      = "File_Log_Store.json"
	defaultRestore       = true
	defaultDatabase      = "MetricStore"
	defaultHost          = "localhost"
	defaultPort          = "5432"
	defaultUser          = "postgres"
	defaultPassword      = "admin"
	defaultSSLMode       = "disable"
)

func ParseServerFlags() *ServerOptions {
	opt := &ServerOptions{}
	defaultPathDB := initDefaulthPathDB()
	flag.StringVar(&opt.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.Int64Var(&opt.StoreInterval, "i", defaultStoreInterval, "address and port to run server")
	flag.StringVar(&opt.Path, "f", defaultFileName, "address and port to run server")
	flag.BoolVar(&opt.Restore, "r", defaultRestore, "address and port to run server")
	flag.StringVar(&opt.DBPath, "d", defaultPathDB, "address and port to run server")
	flag.StringVar(&opt.CryptoKey, "k", "", "key to encrypt/decrypt metrics")
	flag.StringVar(&opt.CryptoKey, "crypto-key", "", "path to private key file for decryption")

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

	if dbPath := os.Getenv("DATABASE_DSN"); dbPath != "" {
		opt.DBPath = dbPath
	}

	if opt.DBPath != "" {
		opt.UseDatabase = true
	} else {
		opt.UseDatabase = false
	}

	if key := os.Getenv("KEY"); key != "" {
		opt.CryptoKey = key
	}

	if path := os.Getenv("CRYPTO_KEY"); path != "" {
		opt.CryptoKey = path
	}

	return opt
}

func initDefaulthPathDB() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		defaultHost, defaultPort, defaultUser, defaultPassword, defaultDatabase, defaultSSLMode)
}
