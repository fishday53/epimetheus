package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Addr           string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func (cfg *Config) Get() error {

	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("config parse error:%v", err)
	}

	addr := flag.String("a", "localhost:8080", "Server address")
	reportInterval := flag.Int("r", 10, "Report interval")
	pollInterval := flag.Int("p", 2, "Poll interval")
	flag.Parse()

	if cfg.Addr == "" {
		cfg.Addr = *addr
	}
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = *reportInterval
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = *pollInterval
	}

	return nil
}
