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
	var parse bool

	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}
	if cfg.Addr == "" {
		cfg.Addr = *flag.String("a", "localhost:8080", "Server address")
		parse = true
	}
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = *flag.Int("r", 10, "Report interval")
		parse = true
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = *flag.Int("p", 2, "Poll interval")
		parse = true
	}
	if parse {
		flag.Parse()
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
