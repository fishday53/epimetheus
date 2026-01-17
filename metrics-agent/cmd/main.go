package main

import (
	"fmt"
	"math/rand/v2"
	"metrics-agent/internal/config"
	"metrics-agent/internal/metrics"
	"time"
)

func main() {
	var cfg config.Config
	cfg.Get()

	url := "http://" + cfg.Addr + "/update"
	metricsBatch := metrics.NewMetricsBatch()
	metricsBatch.Counter["PollCount"] = 0

	for {
		for i := 0; i < (cfg.ReportInterval / cfg.PollInterval); i++ {

			if err := metricsBatch.GetAllRuntimeMetrics(metrics.MetricList); err != nil {
				fmt.Println(err)
			}

			metricsBatch.Counter["PollCount"]++
			metricsBatch.Gauge["RandomValue"] = rand.Float64()

			time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
		}

		if err := metricsBatch.SendAllMetrics(url); err != nil {
			fmt.Println(err)
		}
	}
}
