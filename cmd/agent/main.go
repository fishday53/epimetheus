package main

import (
	"flag"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Addr           string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func main() {
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
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

	url := "http://" + cfg.Addr + "/update"
	metrics := NewMetricsBatch()
	metrics.Counter["PollCount"] = 0

	for {
		for i := 0; i < (cfg.ReportInterval / cfg.PollInterval); i++ {

			if err := metrics.getAllRuntimeMetrics(metricList); err != nil {
				fmt.Println(err)
			}

			metrics.Counter["PollCount"]++
			metrics.Gauge["RandomValue"] = rand.Float64()

			time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
		}

		if err := metrics.sendAllMetrics(url); err != nil {
			fmt.Println(err)
		}
	}
}
