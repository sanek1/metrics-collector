// Package flags предоставляет функциональность для работы с флагами командной строки агента
package flags

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Options struct {
	FlagRunAddr    string
	CryptoKey      string
	ReportInterval int64
	PollInterval   int64
	RateLimit      int64
	ConfigPath     string
}

// AgentFileConfig представляет конфигурацию агента из файла
type AgentFileConfig struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}

const (
	defaultReportInterval = 10
	defaultPollInterval   = 2
	defaultRateLimit      = 2
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
func LoadConfig(filePath string) (*AgentFileConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config AgentFileConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return &config, nil
}

// ApplyFileConfig применяет настройки из конфигурационного файла
func ApplyFileConfig(opt *Options, config *AgentFileConfig) error {
	if config.Address != "" {
		opt.FlagRunAddr = config.Address
	}

	if config.ReportInterval != "" {
		reportInterval, err := ParseDuration(config.ReportInterval)
		if err != nil {
			return fmt.Errorf("wrong report_interval: %w", err)
		}
		opt.ReportInterval = reportInterval
	}

	if config.PollInterval != "" {
		pollInterval, err := ParseDuration(config.PollInterval)
		if err != nil {
			return fmt.Errorf("wrong poll_interval: %w", err)
		}
		opt.PollInterval = pollInterval
	}

	if config.CryptoKey != "" {
		opt.CryptoKey = config.CryptoKey
	}

	return nil
}

func ParseFlags() *Options {
	opt := &Options{}

	flag.StringVar(&opt.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&opt.CryptoKey, "k", "", "key to encrypt/decrypt metrics")
	flag.Int64Var(&opt.ReportInterval, "r", defaultReportInterval, "report interval in seconds")
	flag.Int64Var(&opt.PollInterval, "p", defaultPollInterval, "poll interval in seconds")
	flag.Int64Var(&opt.RateLimit, "l", defaultRateLimit, "outgoing requests to the server ")
	flag.StringVar(&opt.CryptoKey, "crypto-key", "", "path to public key file for encryption")
	flag.StringVar(&opt.ConfigPath, "c", "", "path to config file")
	flag.StringVar(&opt.ConfigPath, "config", "", "path to config file")
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

	if path := os.Getenv("CRYPTO_KEY"); path != "" {
		opt.CryptoKey = path
	}

	if limit := os.Getenv("RATE_LIMIT"); limit != "" {
		opt.RateLimit, _ = strconv.ParseInt(limit, 10, 64)
	}

	return opt
}
