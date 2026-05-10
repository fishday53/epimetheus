// Package config is used to get command-line and Env Metrics-Agent settings.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/caarlos0/env/v6"
)

// Config is a Metrics-Agent configuration storage.
type Config struct {
	Addr           string `env:"ADDRESS" json:"address"`
	ReportInterval int    `env:"REPORT_INTERVAL" json:"report_interval"`
	PollInterval   int    `env:"POLL_INTERVAL" json:"poll_interval"`
	HashKey        string `env:"KEY" json:"hash_key"`
	RateLimit      int    `env:"RATE_LIMIT" json:"rate_limit"`
	BufferSize     int
	CryptoKeyPath  string `env:"CRYPTO_KEY" json:"crypto_key"`
	ConfigPath     string `env:"CONFIG"`
	Transport      string `env:"TRANSPORT" json:"transport"`
}

// Get is a single method to get all Metrics-Agent settings.
func (cfg *Config) Get() error {

	// first: read config file with low priority
	fs := flag.NewFlagSet("metrics-agent-preread", flag.ContinueOnError)
	fs.StringVar(&cfg.ConfigPath, "config", "", "Config file path (short: -c)")
	fs.StringVar(&cfg.ConfigPath, "c", "", "short for -config")
	originalOutput := fs.Output()
	fs.SetOutput(io.Discard)
	fs.Parse(os.Args[1:])

	if config, ok := os.LookupEnv("CONFIG"); ok {
		cfg.ConfigPath = config
	}

	if cfg.ConfigPath != "" {
		if err := cfg.readFromFile(); err != nil {
			return fmt.Errorf("cannot read config file:%v", err)
		}
	}

	// read cmd-line params with medium priority
	fs = flag.NewFlagSet("metrics-agent", flag.ContinueOnError)

	Addr := fs.String("a", "localhost:8080", "Server address")
	ReportInterval := fs.Int("r", 10, "Report interval")
	PollInterval := fs.Int("p", 2, "Poll interval")
	HashKey := fs.String("k", "", "Hash Key")
	RateLimit := fs.Int("l", 1, "Rate limit")
	CryptoKeyPath := fs.String("crypto-key", "", "Public Key path")
	Transport := fs.String("x", "http", "Transport: http or grpc")
	fs.SetOutput(originalOutput)
	fs.Parse(os.Args[1:])

	// replace config file options with explicit cmd-line values
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			cfg.Addr = *Addr
		case "r":
			cfg.ReportInterval = *ReportInterval
		case "p":
			cfg.PollInterval = *PollInterval
		case "k":
			cfg.HashKey = *HashKey
		case "l":
			cfg.RateLimit = *RateLimit
		case "crypto-key":
			cfg.CryptoKeyPath = *CryptoKeyPath
		case "x":
			cfg.Transport = *Transport
		}
	})

	// replace only empty config values with cmd-line defaults
	if cfg.Addr == "" {
		cfg.Addr = *Addr
	}
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = *ReportInterval
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = *PollInterval
	}
	if cfg.RateLimit == 0 {
		cfg.RateLimit = *RateLimit
	}
	if cfg.Transport == "" {
		cfg.Transport = *Transport
	}

	// read envs with high priority
	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("config parse error:%v", err)
	}

	return nil
}

// readFromFile reads config from json
func (cfg *Config) readFromFile() error {
	cfgData, err := os.ReadFile(cfg.ConfigPath)
	if err != nil {
		return fmt.Errorf("error reading file:%v", err)
	}

	err = json.Unmarshal(cfgData, &cfg)
	if err != nil {
		return fmt.Errorf("error parsing json config:%v", err)
	}
	return nil
}
