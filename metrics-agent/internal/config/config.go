package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Addr           string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	HashKey        string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
	BufferSize     int
}

func (cfg *Config) Get() error {
	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("config parse error:%v", err)
	}

	fs := flag.NewFlagSet("metrics-agent", flag.ContinueOnError)

	addr := fs.String("a", "localhost:8080", "Server address")
	reportInterval := fs.Int("r", 10, "Report interval")
	pollInterval := fs.Int("p", 2, "Poll interval")
	hashKey := fs.String("k", "", "Hash Key")
	rateLimit := fs.Int("l", 1, "Rate limit")
	fs.Parse(os.Args[1:])

	if cfg.Addr == "" {
		cfg.Addr = *addr
	}
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = *reportInterval
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = *pollInterval
	}
	if cfg.HashKey == "" {
		cfg.HashKey = *hashKey
	}
	if cfg.RateLimit == 0 {
		cfg.RateLimit = *rateLimit
	}

	return nil
}
