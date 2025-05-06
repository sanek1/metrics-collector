// Package flags
package flags

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type ServerOptions struct {
	FlagRunAddr   string
	StoreInterval int64
	Path          string
	Restore       bool
	DBPath        string
	UseDatabase   bool
	CryptoKey     string
	ConfigPath    string
	TrustedSubnet string `json:"trusted_subnet"`
}

// ServerFileConfig представляет конфигурацию сервера из файла
type ServerFileConfig struct {
	Address       string `json:"address"`
	Restore       bool   `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
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

// ParseDuration преобразует строку длительности в секунды
func ParseDuration(duration string) (int64, error) {
	d, err := time.ParseDuration(duration)
	if err != nil {
		return 0, err
	}
	return int64(d.Seconds()), nil
}

// LoadConfig загружает конфигурацию из JSON-файла
func LoadConfig(filePath string) (*ServerFileConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config ServerFileConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return &config, nil
}

// ApplyFileConfig применяет настройки из конфигурационного файла
func ApplyFileConfig(opt *ServerOptions, config *ServerFileConfig) error {
	if config.Address != "" {
		opt.FlagRunAddr = config.Address
	}

	opt.Restore = config.Restore

	if config.StoreInterval != "" {
		storeInterval, err := ParseDuration(config.StoreInterval)
		if err != nil {
			return fmt.Errorf("wrong store_interval: %w", err)
		}
		opt.StoreInterval = storeInterval
	}

	if config.StoreFile != "" {
		opt.Path = config.StoreFile
	}

	if config.DatabaseDSN != "" {
		opt.DBPath = config.DatabaseDSN
		opt.UseDatabase = true
	}

	if config.CryptoKey != "" {
		opt.CryptoKey = config.CryptoKey
	}

	return nil
}

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
	flag.StringVar(&opt.ConfigPath, "c", "", "path to config file")
	flag.StringVar(&opt.ConfigPath, "config", "", "path to config file")
	flag.StringVar(&opt.TrustedSubnet, "t", "", "Trusted subnet in CIDR format")

	flag.Parse()
	if len(flag.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "Unknown flags:", flag.Args())
		os.Exit(1)
	}

	configPath := opt.ConfigPath
	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	if configPath != "" {
		if cfg, err := LoadConfig(configPath); err == nil {
			if err := ApplyFileConfig(opt, cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Error applying configuration: %v\n", err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		}
	}
	if sbnt := os.Getenv("TRUSTED_SUBNET"); sbnt != "" {
		opt.TrustedSubnet = sbnt
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
